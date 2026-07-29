# DVAPI

**Damn Vulnerable API (DVAPI)** is an intentionally vulnerable API application designed for learning, practicing, and testing API security concepts.

The project provides a realistic development environment where developers and security researchers can explore common API vulnerabilities in a safe and controlled setup.


## Quick start

You need Docker and Make.

```bash
make up
```

That starts everything, sets up the database, and loads sample data. When it’s ready:

| What        | Open / connect at         |
| ----------- | ------------------------- |
| App (UI)    | http://localhost:5173     |
| API         | http://localhost:8080     |
| Email inbox | http://localhost:8025     |

Password-reset emails and similar go to the **email inbox** page above (nothing is sent to real addresses).

Stop everything with:

```bash
make down
```

## What’s running

| Piece        | Role                                      | Port on your machine |
| ------------ | ----------------------------------------- | -------------------- |
| Frontend     | Vue web app                               | 5173                 |
| Backend      | Go API                                    | 8080                 |
| Library      | Small Flask helper service                | 5000                 |
| PostgreSQL   | Database                                  | 5433                 |
| Redis        | Cache                                     | 6380                 |
| MailHog      | Fake mail server + web inbox              | 1025 (mail), 8025 (UI) |

## Common commands

```bash
make up          # start everything (build if needed), migrate, seed
make down        # stop and remove containers
make ps          # what’s running?
make logs        # follow all logs
make clean       # stop and wipe database/cache data
```

`make dev` does the same thing as `make up`.

### Database

```bash
make migrate-up       # apply new schema changes
make migrate-down     # undo the last change
make migrate-status   # see migration state
```

### Sample data

The stack must already be up (`make up`).

```bash
make seed                           # load all default sample data
make seed ARGS="users"              # only users
make seed ARGS="courses enrollments"
make seed-status                    # what’s already loaded?
```

Default seed sets: users, courses, enrollments, assignments, submissions, announcements.

### Other useful bits

```bash
make rebuild          # rebuild images from scratch
make prune            # remove containers, data volumes, and local images
make backend-shell    # shell inside the backend container
make db-shell         # open a database prompt
```

Run `make help` for the full list.

-----

## Vulnerability Surface

DVAPI intentionally includes vulnerabilities across multiple attack surfaces and protocols to simulate realistic API security challenges. 

### Attack Surfaces

- REST API
- GraphQL
- SMTP
- Database Layer
- Operating System (OS) Layer

### Vulnerability Categories

The project covers vulnerabilities from a variety of security domains, including:

- OWASP API Security Top 10
- Injection vulnerabilities (across multiple protocols and technologies)
- Authentication and authorization flaws
- Information disclosure
- Input validation failures

> **Note:** Specific vulnerabilities and exploitation paths are intentionally undocumented. Discovering and exploiting them is part of the learning experience.