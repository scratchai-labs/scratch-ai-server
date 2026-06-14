# Scratch AI Coach

`Scratch AI Coach` is the server-side repository for the classroom product line. The current mainline is now a `Gin`-based `Go` backend plus the Teacher Web app, and the older `Python FastAPI` server prototype has been removed from this repository.

Cross-repo docs, architecture notes, and planning live in [`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/README.en.md).

## Current Scope

- This repository only maintains the **server teaching edition**
- The core is the server API
- Admins can manage teacher and student accounts through the web console
- Teachers manage students and assignments through a web console
- Students log in from the client and only receive hints there
- All AI processing stays on the server side

## Target Capabilities

- Teacher register and login
- Admin create, disable, enable, and reset teacher accounts, and create, disable, enable, and reset student accounts for specific teachers
- Batch creation of student accounts and passwords
- Reference `sb3` upload and analysis
- Student client login and progress reporting
- Server-side DeepSeek hint generation
- Live teacher progress dashboard

## Current Access Flow

- The shared web login entry is `/login`
- Admins and teachers use the same login page; admins land on `/admin`, while teachers land on the teaching workspace
- The admin console is part of the same web app, but uses dedicated routes: `/admin`, `/admin/teachers`, and `/admin/students`
- There is no public web sign-up page for teachers right now; the first teacher must be created through `POST /api/teacher/register` or later from the admin console
- Use `staging` as the online test environment, with its own `DATABASE_URL` and `SB3_STORAGE_DIR`

## Development Notes

For backend work, use the Go commands directly:

```bash
git clone git@github.com:scratchai-labs/scratch-ai-server.git
cd scratch-ai-server
npm ci
cd apps/server-api
go test ./...
go run ./cmd/server-api
```

For the teacher web app on its own:

```bash
cd apps/server-web
npm run test
npm run dev
```

Use the root `npm run` commands only as monorepo shortcuts:

- `npm run server:api:dev`
- `npm run server:api:test`
- `npm run server:web:dev`
- `npm run server:web:test`
- `npm run server:dev`
- `npm run server:build`

Current database behavior:

- local development uses `SQLite` by default
- the server switches to `Postgres` when `DATABASE_URL` is provided
- raw `sb3` files are stored under `SB3_STORAGE_DIR`

Current integration status:

- the teacher web app has been verified with a real browser click-through
- the real API flow has been validated for login, students, assignments, live dashboard, and logout
- the admin console is now available, and the first admin can be bootstrapped with `ADMIN_BOOTSTRAP_USERNAME` / `ADMIN_BOOTSTRAP_PASSWORD` and land on `/admin`

The detailed Chinese development spec for the next phase lives in [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md).
The current deployment guide lives in [`docs/deployment.zh-CN.md`](docs/deployment.zh-CN.md).

## Documentation

- Chinese overview: [`README.zh-CN.md`](README.zh-CN.md)
- Project structure: [`docs/project-structure.zh-CN.md`](docs/project-structure.zh-CN.md)
- Architecture: [`docs/architecture.zh-CN.md`](docs/architecture.zh-CN.md)
- Deployment guide: [`docs/deployment.zh-CN.md`](docs/deployment.zh-CN.md)
- Server development spec: [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)

## License

This project is licensed under [`AGPL-3.0`](LICENSE).
