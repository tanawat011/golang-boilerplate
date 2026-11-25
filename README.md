# Go REST API Boilerplate

[![CI](https://github.com/tanawat011/golang-boilerplate/actions/workflows/ci.yml/badge.svg)](https://github.com/tanawat011/golang-boilerplate/actions/workflows/ci.yml)

A small, tested starting point for a JSON REST API in Go: Gin, GORM + PostgreSQL,
JWT authentication with owner / admin rules, and a layered
(handler → service → repository) structure that keeps HTTP, business logic and
SQL apart.

It is meant to be cloned and changed, not imported as a framework.

## Features

- **Auth**: register, login (JWT access token, HS256), `GET /auth/me`.
  Passwords hashed with bcrypt (cost 12 by default); login failures take the
  same time whether or not the email exists.
- **Access rules**: users can read, edit and delete only themselves; admins can
  do anything; only admins can change roles or deactivate accounts.
- **Secure defaults**: the app refuses to start without a JWT secret of at
  least 32 characters (and refuses the example secret in production);
  per-IP rate limit on register / login; security headers (HSTS in
  production); CORS closed unless you list origins; 1 MiB body limit;
  trusted proxies explicit.
- **PostgreSQL via GORM**: UUID keys, soft deletes, unique email / username
  enforced by the database, errors translated so callers never see GORM types.
- **Ops**: structured logs (zap), `/api/v1/health` that checks the database
  (503 when it is down), graceful shutdown, multi-stage Docker image on
  distroless running as non-root.
- **Tests**: unit tests for config, tokens, services, middleware and every
  route (in-memory repository), plus integration tests against a real
  PostgreSQL. CI runs lint, tests, `govulncheck` and a `docker compose` smoke
  test.

## Quick start

Requires Docker.

```bash
git clone https://github.com/tanawat011/golang-boilerplate.git
cd golang-boilerplate
docker compose up --build
```

The API is on <http://localhost:8080>. Try it:

```bash
curl localhost:8080/api/v1/health

curl -X POST localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","email":"alice@example.com","password":"correct-horse"}'

curl -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"correct-horse"}'
# -> {"data":{"access_token":"...","token_type":"Bearer","expires_at":"...","user":{...}}}

curl localhost:8080/api/v1/auth/me -H "Authorization: Bearer <access_token>"
```

### Run on the host instead

Requires Go 1.26.10 or newer.

```bash
cp .env.example .env   # then set JWT_SECRET (openssl rand -base64 48)
make db-up             # PostgreSQL only, on 127.0.0.1:15432
make run
```

## API

All routes are under `/api/v1`. Responses use one envelope:
`{"success": bool, "message": "...", "data": ..., "error": ..., "meta": ...}`.

| Method   | Path              | Who                       | Notes |
|----------|-------------------|---------------------------|-------|
| `GET`    | `/health`         | anyone                    | 200, or 503 if the database is down |
| `POST`   | `/auth/register`  | anyone (rate limited)     | `username`, `email`, `password` (8–72 bytes), `full_name` |
| `POST`   | `/auth/login`     | anyone (rate limited)     | `email`, `password` → access token |
| `GET`    | `/auth/me`        | signed in                 | the current user |
| `GET`    | `/users`          | admin                     | `?page=1&limit=10` (max 100), newest first |
| `GET`    | `/users/:id`      | that user or admin        | |
| `PATCH`  | `/users/:id`      | that user or admin        | `username`, `email`, `full_name`; admin only: `role`, `is_active` |
| `DELETE` | `/users/:id`      | that user or admin        | soft delete |

Status codes: 400 malformed JSON, 401 missing / bad token or wrong password,
403 not allowed or account disabled, 404 not found, 409 email / username
taken, 422 validation errors (one entry per field), 429 rate limited (with
`Retry-After`).

**Making the first admin**: register normally, then promote the account in
the database and log in again (the role is part of the token):

```bash
docker compose exec postgres psql -U postgres -d golang_boilerplate \
  -c "UPDATE users SET role = 'admin' WHERE email = 'alice@example.com';"
```

## Project structure

```
cmd/api/                 main: config, logger, database, server, shutdown
internal/
  auth/                  JWT issue / verify (HS256, issuer, expiry required)
  config/                env + .env loading and validation
  handler/               HTTP handlers, response envelope, error mapping
  middleware/            auth, CORS, security headers, rate limit, logging, recovery
  model/                 GORM models
  repository/            data access (+ repositorytest: in-memory fake for tests)
  router/                routes and middleware order
  service/               business logic and access rules (no HTTP)
  testutil/              throwaway PostgreSQL database for integration tests
pkg/
  pagination/            page / limit query parameters
  validator/             struct validation with JSON field names
test/integration/        the whole API against PostgreSQL
```

## Configuration

Set through environment variables; `make run` also reads `.env` from the
working directory (real environment variables win). See `.env.example`.

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | HTTP port |
| `APP_MODE` | `development` | `production` = release-mode Gin, JSON logs, HSTS, example secret refused |
| `APP_VERSION` | `dev` | Shown by `/health` |
| `CORS_ALLOWED_ORIGINS` | *(empty)* | Comma-separated origins; empty = same-origin only; `*` = any |
| `TRUSTED_PROXIES` | *(empty)* | Proxies whose `X-Forwarded-For` is trusted (used for rate limiting) |
| `DB_HOST` / `DB_PORT` | `localhost` / `15432` | PostgreSQL address |
| `DB_USER` / `DB_PASSWORD` / `DB_NAME` | `postgres` / *(empty)* / `golang_boilerplate` | |
| `DB_SSL_MODE` | `disable` | Use `require` or `verify-full` outside local development |
| `DB_TIMEZONE` | `UTC` | Session time zone |
| `JWT_SECRET` | *(required)* | At least 32 characters |
| `JWT_EXPIRATION` | `1h` | Access token lifetime (Go duration) |
| `JWT_ISSUER` | `golang-boilerplate` | `iss` claim, checked on every request |
| `BCRYPT_COST` | `12` | 10–31 |
| `AUTH_RATE_LIMIT_PER_MINUTE` | `10` | Per client IP, on register and login |

## Testing

```bash
make test               # unit tests, no database needed
make db-up              # PostgreSQL on 127.0.0.1:15432
make test-integration   # unit + integration tests
make lint               # golangci-lint v2
make vuln               # govulncheck
```

Integration tests are behind the `integration` build tag. Each test creates
its own database on the server in `TEST_DATABASE_URL` and drops it afterwards,
so they never touch your development data.

## Deployment notes

- Build the image with `docker build -t golang-boilerplate .` and run it with
  `APP_MODE=production`, a random `JWT_SECRET`, real database credentials and
  `DB_SSL_MODE=require` (or `verify-full`).
- Put it behind HTTPS and set `TRUSTED_PROXIES` to your load balancer, or the
  rate limiter will see the proxy's IP for every client.
- The schema is created with GORM `AutoMigrate` at startup. For anything
  beyond a prototype, switch to versioned migrations (e.g. goose or
  golang-migrate).
- The rate limiter is in memory, per instance. With several replicas, limit
  at the load balancer or move it to a shared store.
- Access tokens are stateless: a role change, deactivation or deletion takes
  effect when the user's current token expires (`JWT_EXPIRATION`). Add
  refresh tokens or a deny-list if you need instant revocation.

## Upgrading from the old version

The November 2026 refresh changes the API. In short:

- `PUT /api/v1/users/:id` is now `PATCH`.
- Every `/users` route needs sign-in (`Authorization: Bearer <token>`);
  `GET /users` is admin only.
- `POST /users` is replaced by `POST /auth/register` + `POST /auth/login`.
- `JWT_SECRET` (32+ characters) is required; `JWT_EXPIRATION_HOURS` became
  `JWT_EXPIRATION`.

See [CHANGELOG.md](CHANGELOG.md) for the full list.

## History

The repository was created in December 2018; that first version, a
gorilla/mux skeleton, is kept on the `master` branch. The current Gin / Clean
Architecture codebase on `main` dates from January 2024. Refreshed in November
2026: Go 1.26+, updated dependencies, auth endpoints and access rules, secure
defaults, tests and CI.

## License

MIT, see [LICENSE](LICENSE).
