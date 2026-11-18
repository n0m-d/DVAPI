# How DVAPI fits together

DVAPI is a deliberately vulnerable API, but underneath the bugs it's a pretty
ordinary web app: an online learning platform called **Schole**. Instructors
create courses, lessons and assignments. Students enroll, work through lessons,
submit work and get graded. Admins look after users and keep an eye on the logs.

That ordinariness is on purpose. The weaknesses only teach you something if
they live in code that looks like code you'd find at work: layered services,
generated SQL, background jobs, a queue, a GraphQL endpoint nobody remembers
adding. This document walks through how all of that fits together.

---

## The big picture

```
                         ┌──────────────────────────┐
     Browser ───────────▶│  Frontend (Vue + Caddy)  │  :5173
        │                └──────────────────────────┘
        │ REST / GraphQL / SSE
        ▼
┌───────────────────────────────────────────────┐
│  Backend (Go + gin)                     :8080 │
│                                               │       ┌─────────────┐
│   /api/v1   /api/v2   /api/v3/beta   GraphQL  │──────▶│ PostgreSQL  │ :5433
│        │                                      │       └─────────────┘
│        ▼                                      │       ┌─────────────┐
│   handler → service → repository → sqlc (db)  │──────▶│    Redis    │ :6380
│        └──── query (read model) ───┘          │       └─────────────┘
│                                               │         ▲  cache, token blacklist,
│   mail worker (asynq)   job scheduler (cron)  │         │  rate limits, pub/sub,
│                                               │         │  mail queue
│   ┌──────────────────────────────────────┐    │         │
│   │ Library service (Flask)        :5000 │    │       ┌─────────────┐
│   │ shares the backend's network         │    │──────▶│   MailHog   │ :1025 / UI :8025
│   └──────────────────────────────────────┘    │       └─────────────┘
└───────────────────────────────────────────────┘
        │ access.log (JSON)
        ▼
   Promtail ──▶ Loki :3100 ──▶ Grafana :3001
```

Everything runs from `docker-compose.yml`, and `make up` takes you from
nothing to a seeded, working stack.

---

## The pieces

### Frontend: `frontend/`

A Vue 3 + TypeScript single-page app built with Vite and served by Caddy.

- `src/views/` is split by role: `student/`, `instructor/`, `admin/`, and
  `shared/` for things like the profile page. There are also the auth pages
  (login, register, forgot-password).
- `src/api/` has one small module per backend area (`courses.ts`,
  `assignments.ts`, `notes.ts`…). They all share one HTTP client that points at
  `VITE_API_BASE_URL`, which is `/api/v2` by default.
- `src/stores/` holds the Pinia stores. `auth` keeps the JWT and the current
  user, and `notifications` holds the live notification feed.
- `src/router/index.ts` decides which role can see which page. That's only a
  convenience for the UI. The backend is what enforces access.

The API base URLs are baked in at build time (Docker build args), so changing
them means rebuilding the frontend image.

### Backend: `backend/`

A Go service built on gin. This is where nearly everything happens.

```
backend/
├── cmd/
│   ├── server/     entrypoint: config → logger → Postgres → Redis → server + jobs
│   ├── migrate/    runs goose migrations from migrations/
│   └── seed/       fills the database with fake users, courses, submissions…
├── migrations/     SQL schema history (goose)
├── sql/queries/    hand-written SQL that sqlc turns into Go
├── graph/          GraphQL schema + resolvers (gqlgen)
└── internal/
    ├── api/v1, v2, v3   route tables, one package per API version
    ├── middleware/      request IDs, logging, JWT checks, role checks, rate limits
    ├── handler/         HTTP in/out: parse the request, call a service, write JSON
    ├── service/         business rules: who can do what, and what happens next
    ├── query/           read-only "dashboard" queries (stats, lists, analytics)
    ├── repository/      thin wrappers that map sqlc rows to domain types
    ├── db/              sqlc-generated code (don't edit by hand)
    ├── domain/          plain structs and shared error values used everywhere
    ├── email/           SMTP sender, HTML templates, asynq queue + worker
    ├── pubsub/          Redis pub/sub for real-time notifications
    ├── cache/           Redis-backed user cache and JWT blacklist
    ├── jobs/            scheduled background work (gocron)
    ├── config/          environment variables → Config
    ├── logger/          slog: pretty text to stderr, JSON to the access log
    └── utils/           JWT, password hashing, OTPs, email helpers
```

### Library service: `library/`

A tiny Flask app that serves a catalogue of authors and books from
`library.xml`. It has no auth and no database, just an XML file and some XPath.

There's one detail worth knowing: in Compose it runs with
`network_mode: "service:backend"`. It shares the backend container's network
stack, so to the backend it's simply `localhost:5000`. The frontend never calls
it directly. Library requests go through a backend endpoint that fetches a URL
on its behalf.

### Supporting services

| Service  | What it's for |
| -------- | ------------- |
| PostgreSQL | Everything durable: users, courses, lessons, enrollments, assignments, submissions, grades, announcements, notifications, notes, OTPs. |
| Redis | Wears many hats: user cache, the JWT blacklist for logout, per-IP rate limits on v2, the pub/sub channel for live notifications, and the asynq mail queue. |
| MailHog | Catches every email the app sends. Open `http://localhost:8025` to read password-reset codes and enrollment emails. |
| Promtail → Loki → Grafana | Promtail tails the backend's JSON access log and ships it to Loki. Grafana (`:3001`, anonymous access) lets you search and graph it. |

---

## Following a request through the backend

Take a student opening a course: `GET /api/v2/courses/:courseId`.

1. **Global middleware** runs first, for every route:
   - a panic recovery handler;
   - a **request ID** (sent back in a header, and attached to every log line);
   - the **access logger**;
   - **CORS** for the frontend origins.
2. **v2 middleware:** the request is **rate limited** per IP (60/min, counted
   in Redis).
3. **Auth:** `JWTVerify` takes the token from the `Authorization` header (or
   `?token=` for SSE, since `EventSource` can't set headers), checks it, makes
   sure it isn't on the Redis blacklist, and stores the claims on the request.
4. **Role gate:** some route groups also go through `RoleRestrict("student")`,
   `("instructor")` or `("admin")`.
5. **Handler** (`handler/course_handler.go`) parses the path and query
   parameters, pulls the current user from the claims, and calls the service.
6. **Service** (`service/course_service.go`) applies the business rules (is it
   published? is this user enrolled? do they own it?) and asks the repository
   for data.
7. **Repository → sqlc → Postgres.** Most SQL is written in `sql/queries/*.sql`
   and compiled by sqlc into type-safe Go in `internal/db`. Repositories convert
   those rows into `domain` structs.
8. On the way back, the handler writes JSON. If something went wrong, the error
   goes through `handler/errors.go`, which turns domain errors
   (`domain.ErrNotFound`, `domain.ErrForbidden`…) into status codes and messages
   in one place.

### Reads vs. writes

There are two paths into the database:

- **`service/`** handles anything that *changes* things or needs real business
  rules: enroll, submit, grade, create a course.
- **`query/`** is a read model for the "show me a list or a dashboard" kind of
  request: admin stats, user lists, course catalogues, analytics, progress. It
  goes straight from handler to `query.Service` to sqlc. That keeps the write
  services small and puts the reporting SQL in one place.

Handlers can use both. `CourseHandler`, for example, writes through the course
service and reads lists through `query`.

---

## The three API versions

The versions all run at the same time, like a real product that never quite
retires its old endpoints.

| Prefix | What lives there |
| ------ | ---------------- |
| `/api/v2` | The main API the frontend uses: auth, users, courses, lessons, assignments, submissions, grades, announcements, notifications, admin tools, the library proxy, and GraphQL (`/api/v2/query`, playground at `/api/v2/playground`). |
| `/api/v1` | Older endpoints that are still around: an earlier password-reset flow, an older profile update, and an admin-only notes page. |
| `/api/v3/beta` | Newer "beta" endpoints: enroll/unenroll by request body, and token refresh. |

Swagger docs are generated from handler comments (`swag`) and served by the
backend.

### GraphQL

`graph/schema.graphqls` defines **Notes** (personal notes users keep) and a
user listing. gqlgen generates `graph/generated.go`. You write the resolvers in
`graph/schema.resolvers.go`, and they call into the note service and the
`query` read model.

---

## Things that happen in the background

### Emails: queued, not sent inline

When the app needs to send an email (welcome, password-reset OTP, enrollment
confirmation, grade posted), the service doesn't talk to SMTP itself. It drops
a job on an **asynq** queue in Redis (`email/queue.go`) and returns
immediately.

A **mail worker** (`email/worker.go`) runs inside the same backend process. It
picks up jobs, renders an HTML template from `email/templates/`, and sends it
through `email/email.go` to MailHog. OTP emails go on a `critical` queue that
gets priority. Everything else uses `default`. Failed sends are retried.

The upshot is that API responses don't wait on the mail server, and a
mail-server failure doesn't fail the request. It shows up in the logs instead.

### Live notifications: Redis pub/sub + SSE

When something happens that a user should hear about (a new announcement, a
grade, an enrollment), the service:

1. saves a notification row in Postgres, then
2. publishes it to the Redis channel `notifications:user:<id>`.

The frontend keeps a Server-Sent Events connection open to
`/api/v2/notifications/stream`. That handler subscribes to the user's channel
and forwards each event as it arrives, which is how the bell icon updates
without a page reload.

### Scheduled jobs

`internal/jobs` uses gocron to run two chores. They run once at startup and
then on a schedule:

- **close_overdue**: closes assignments whose due date has passed.
- **purge_otps**: deletes expired password-reset codes.

---

## Auth in one paragraph

Users register and log in through `/api/v2/auth/*` and get back an HS256 JWT
carrying their user ID and role. **The signing secret is randomly generated
every time the backend starts** (`config.go`), so restarting the backend logs
everyone out. Logout puts the token's ID on a Redis blacklist until it would
have expired anyway. Password reset is a three-step flow (request → verify OTP
→ confirm new password), with the OTP delivered by email to MailHog.

---

## Logging and observability

The logger (`internal/logger`) writes to two places:

- **stderr**, as readable, coloured text, which is what you see in `make logs`;
- **`backend/.log/access.log`**, as JSON, rotated by lumberjack.

Promtail reads only the active `access.log` and pushes it to Loki. Grafana
already has Loki set up as a data source (`grafana/provisioning/`). Every line
carries the request ID, so you can follow one request from start to finish.

---

## Data: schema, migrations, seeding

- **Schema** lives in `backend/migrations/` as goose SQL files, applied in
  order by `make migrate-up` (and automatically by `make up`).
- **Queries** live in `backend/sql/queries/`. After you change one, regenerate
  with sqlc. Never edit `internal/db/*.sql.go` directly.
- **Seed data** comes from `cmd/seed`, which uses gofakeit to create users of
  every role, courses, lessons, enrollments, assignments and submissions. That
  way the app feels lived-in from the first login.

---

## Running it day to day

```bash
make up       # build, start, migrate, seed (creates .env files if missing)
make logs     # follow everything
make ps       # what's running
make down     # stop
make clean    # stop and wipe Postgres/Redis data
```

The backend container runs **air**, so Go changes on disk rebuild and restart
it automatically. The library container mounts its folder too. The frontend is
a built image, so frontend changes need a rebuild.

Ports at a glance: frontend **5173**, backend **8080**, library **5000**,
Postgres **5433**, Redis **6380**, MailHog **1025** (SMTP) / **8025** (UI),
Loki **3100**, Grafana **3001**.
