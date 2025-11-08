package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID          uuid.UUID
	Phone       string
	DisplayName *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Device struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Label      *string
	PushToken  *string
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, phone, display_name, created_at, updated_at FROM users WHERE id=$1`, id)
	var u User
	if err := row.Scan(&u.ID, &u.Phone, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) GetByPhone(ctx context.Context, phone string) (*User, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, phone, display_name, created_at, updated_at FROM users WHERE phone=$1`, phone)
	var u User
	if err := row.Scan(&u.ID, &u.Phone, &u.DisplayName, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpsertUserAndDevice(ctx context.Context, phone string, label, pushToken *string) (user *User, device *Device, err error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	user, err = upsertUser(ctx, tx, phone)
	if err != nil {
		return nil, nil, err
	}

	device, err = upsertDevice(ctx, tx, user.ID, label, pushToken)
	if err != nil {
		return nil, nil, err
	}

	return user, device, nil
}

func (r *Repository) UpdateDeviceLastSeen(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE devices SET last_seen_at = now() WHERE id=$1`, deviceID)
	return err
}

func upsertUser(ctx context.Context, q pgx.Tx, phone string) (*User, error) {
	var user User
	err := q.QueryRow(ctx, `INSERT INTO users (phone) VALUES ($1)
      ON CONFLICT (phone) DO UPDATE SET updated_at = now()
      RETURNING id, phone, display_name, created_at, updated_at`, phone).Scan(&user.ID, &user.Phone, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func upsertDevice(ctx context.Context, q pgx.Tx, userID uuid.UUID, label, pushToken *string) (*Device, error) {
	var device Device
	err := q.QueryRow(ctx, `INSERT INTO devices (user_id, label, push_token)
      VALUES ($1, $2, $3)
      ON CONFLICT (user_id, COALESCE(label, '')) DO UPDATE
        SET push_token = EXCLUDED.push_token, last_seen_at = now()
      RETURNING id, user_id, label, push_token, created_at, last_seen_at`,
		userID, label, pushToken,
	).Scan(&device.ID, &device.UserID, &device.Label, &device.PushToken, &device.CreatedAt, &device.LastSeenAt)
	if err != nil {
		return nil, err
	}
	return &device, nil
}
