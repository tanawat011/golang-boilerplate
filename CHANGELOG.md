# Changelog

All notable changes to this template. Earlier changes are in the git history.

## 2026-11-25: refresh

### Breaking

- `PUT /api/v1/users/:id` is now `PATCH /api/v1/users/:id` (it always was a partial update).
- Every `/api/v1/users` route needs a bearer token. `GET /users` is admin only;
  `GET`, `PATCH` and `DELETE /users/:id` are allowed for that user or an admin.
- `POST /api/v1/users` is gone: sign up with `POST /api/v1/auth/register`, then
  `POST /api/v1/auth/login` for a token.
- Only admins can change `role` or `is_active` (before, any signed-in user could
  make themselves admin).
- Token claims changed (`sub` + `role`, issuer checked), so tokens issued by the
  old version are refused.
- Configuration: `JWT_SECRET` is required (at least 32 characters; the example
  value is refused in production); `JWT_EXPIRATION_HOURS` became
  `JWT_EXPIRATION` (a Go duration, default `1h`); the `REDIS_*` variables are
  gone; `DB_PORT` defaults to `15432`; the database time zone is `DB_TIMEZONE`
  (default `UTC`, was hard-coded to Asia/Bangkok).
- Validation errors name fields by their JSON name (`full_name`, not `FullName`).
- docker-compose publishes PostgreSQL on `127.0.0.1:15432` (was `5432`) and no
  longer starts Redis. `make migrate` is gone (it did nothing).
- Requires Go 1.26.10 or newer.

### Added

- `POST /auth/register`, `POST /auth/login`, `GET /auth/me`.
- Per-IP rate limit on register / login, security headers (HSTS in production),
  1 MiB body limit, `CORS_ALLOWED_ORIGINS`, `TRUSTED_PROXIES`, `BCRYPT_COST`.
- `/health` checks the database and answers 503 when it is down.
- Unit tests, integration tests against PostgreSQL, GitHub Actions CI.

### Changed

- All dependencies updated (govulncheck: 9 reachable vulnerabilities to 0).
- 500 responses no longer include internal error text; an empty user list is
  `[]`, not `null`; users include `created_at` / `updated_at`.
- Docker image: static binary on distroless, non-root, no `.env` baked in.
