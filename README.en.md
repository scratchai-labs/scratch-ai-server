# Scratch AI Coach

`Scratch AI Coach` server is the split-out classroom backend repo. It focuses on the teaching workflow built with `FastAPI + Vue`: teacher auth, student accounts, releases, progress reporting, and AI hint APIs.

## Why This Project Exists

Scratch helped many people fall in love with computers for the first time. Since Scratch itself is open source, this project is being organized as a long-term open source repository too, so teachers, learners, and contributors can use it, review it, and evolve it in public.

## Current Scope

- This repository only maintains the **server teaching edition**
- The stack is `Python FastAPI + Vue`
- It contains `server-api` and `server-web`
- Chinese is the primary product language today, while the core open source docs are bilingual

## What It Does Today

- Handles teacher register/login
- Handles student account creation and login
- Manages `sb3` release assignments
- Stores student progress updates
- Generates server-side AI hints
- Exposes a live teacher dashboard

## Deployment Focus

This repository does not build desktop installers. Its release focus is:

- deploying `apps/server-api`
- building and deploying `apps/server-web`

See [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md) for the current deployment notes.

## Local Development

```bash
git clone git@github.com:scratchai-labs/scratch-ai-server.git
cd scratch-ai-server
npm ci
npm run test
```

Common commands:

```bash
npm run build
npm run test
npm run server:web:test
npm run server:api:test
npm run server:dev
```

Run the server stack locally:

```bash
npm run server:dev
```

## Documentation

- Project structure: [`docs/project-structure.en.md`](docs/project-structure.en.md)
- Server development doc (Chinese): [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)
- Roadmap: [`docs/roadmap.en.md`](docs/roadmap.en.md)
- Engineering docs index: [`docs/README.zh-CN.md`](docs/README.zh-CN.md)
- Server API: `apps/server-api`
- Teacher dashboard: `apps/server-web`

## Contributing

Contributions are welcome through issues, pull requests, docs improvements, and classroom feedback.

- Read [`CONTRIBUTING.en.md`](CONTRIBUTING.en.md) before submitting code
- Follow [`CODE_OF_CONDUCT.en.md`](CODE_OF_CONDUCT.en.md) in community spaces
- Do not report security issues publicly; see [`SECURITY.en.md`](SECURITY.en.md)
- Support and discussion guidance lives in [`SUPPORT.en.md`](SUPPORT.en.md)

## Future Direction

The direction from here includes:

- a fuller classroom workflow
- stronger deployment and migration discipline
- cleaner protocol boundaries with the networked desktop client

See [`docs/roadmap.en.md`](docs/roadmap.en.md) for the current direction.

## License

This project is licensed under [`AGPL-3.0`](LICENSE).
