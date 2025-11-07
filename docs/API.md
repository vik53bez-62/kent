# Kent API (fragment)

POST /v1/auth/otp/request
  body: { phone: string }
  200: { ok: true }

POST /v1/auth/otp/verify
  body: { phone: string, code: string }
  200: { access: string, refresh: string }
