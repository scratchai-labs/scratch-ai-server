# Scratch AI Coach

`Scratch AI Coach` is the server-side repository for the classroom product line. The current mainline is now `Go API + Teacher Web`, while the existing `Python FastAPI` code stays only as a transition prototype.

Cross-repo docs, architecture notes, and planning live in [`scratch-ai-docs`](https://github.com/scratchai-labs/scratch-ai-docs/blob/main/README.en.md).

## Current Scope

- This repository only maintains the **server teaching edition**
- The core is the server API
- Teachers manage students and assignments through a web console
- Students log in from the client and only receive hints there
- All AI processing stays on the server side

## Target Capabilities

- Teacher register and login
- Batch creation of student accounts and passwords
- Reference `sb3` upload and analysis
- Student client login and progress reporting
- Server-side DeepSeek hint generation
- Live teacher progress dashboard

## Development Notes

The root commands now target the Go API and the teacher web app directly:

```bash
git clone git@github.com:scratchai-labs/scratch-ai-server.git
cd scratch-ai-server
npm ci
npm run server:api:test
npm run server:web:test
npm run server:dev
```

Current database behavior:

- local development uses `SQLite` by default
- the server switches to `Postgres` when `DATABASE_URL` is provided
- raw `sb3` files are stored under `SB3_STORAGE_DIR`

Current integration status:

- the teacher web app has been verified with a real browser click-through
- the real API flow has been validated for login, students, assignments, live dashboard, and logout

The detailed Chinese development spec for the next phase lives in [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md).

## Documentation

- Chinese overview: [`README.zh-CN.md`](README.zh-CN.md)
- Project structure: [`docs/project-structure.zh-CN.md`](docs/project-structure.zh-CN.md)
- Architecture: [`docs/architecture.zh-CN.md`](docs/architecture.zh-CN.md)
- Server development spec: [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)

## License

This project is licensed under [`AGPL-3.0`](LICENSE).
