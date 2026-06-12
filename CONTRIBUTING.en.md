# Contributing

Thanks for your interest in `Scratch AI Coach`.

This repository now only maintains one clear goal: the `apps/server-api` + `apps/server-web` server-side monorepo. Please align with the following rules before opening a pull request.

## Before You Start

- Read [`README.en.md`](README.en.md)
- Review the repo layout in [`docs/project-structure.en.md`](docs/project-structure.en.md)
- Review the server development notes in [`docs/server-development.zh-CN.md`](docs/server-development.zh-CN.md)
- Open an issue first for large changes, new features, or roadmap shifts

## Environment

- Node.js `22`
- Go `1.26`
- npm workspaces
- Windows, macOS, or Linux for development

Bootstrap:

```bash
npm ci
npm run test
```

## Good Contribution Targets

- Documentation improvements
- Bug fixes with reproduction steps, API details, and logs
- Regression tests for API contracts, front-end/back-end integration, and the `server-web` smoke test
- Feature proposals grounded in real classroom scenarios and success criteria

## Before You Submit

- Prefer test-first changes whenever practical
- Update docs when commands, layout, or release behavior changes
- Run the tests that match your change scope; when unsure, run `npm run test`
- Do not commit `node_modules/`, `dist/`, `.venv/`, `.pytest_cache/`, temporary screenshots, or local debug output

## Commit Messages

The repository currently prefers concise Chinese commit logs with a simple type prefix:

- `feat:` new feature
- `fix:` bug fix
- `improve:` documentation, structure, or engineering cleanup

Suggested format:

```text
improve: organize the server-side repository baseline
Problem: the repo still had old workspace references
Approach: align README, support docs, cleanup scripts, and ignore rules
```

## Pull Request Expectations

- explain the motivation
- explain the impact scope
- list the verification commands you ran
- include notes or screenshots when API or Web entrypoints changed
- clearly separate “implemented now” from “future roadmap” when touching strategy docs

## Changes We Usually Do Not Want in One Shot

- adding a new runtime, backend language, or heavy infrastructure without discussion
- combining product refactors, doc rewrites, and release overhauls in one PR
- changing API contracts or local startup flows without verification
- merging unimplemented desktop or single-machine ideas into the current mainline

## Conduct and Security

- Follow [`CODE_OF_CONDUCT.en.md`](CODE_OF_CONDUCT.en.md)
- Report vulnerabilities through the private path described in [`SECURITY.en.md`](SECURITY.en.md)
- Use [`SUPPORT.en.md`](SUPPORT.en.md) for usage questions and general discussion
