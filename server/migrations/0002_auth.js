// 0002_auth.js – создаём пользователя MongoDB и назначаем роли
db = db.getSiblingDB('admin');   // пользователи хранятся в базе admin

if (!db.getUser('kent')) {
    db.createUser({
        user: "kent",
        pwd:  "Fq0cqDWQiuuLRPAlEyrhbbK5sHGtYNxK",   // тот же пароль, что в PostgreSQL
        roles: [
            { role: "readWrite", db: "kent_afcg" },
            { role: "dbAdmin",   db: "kent_afcg" }
        ]
    });
    print('User "kent" created with readWrite+dbAdmin on kent_afcg');
} else {
    print('User "kent" already exists – skipping');
}