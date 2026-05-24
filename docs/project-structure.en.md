# Project Structure

This repository is the split-out server monorepo. The JavaScript side keeps the teacher dashboard in an npm workspace, while the teaching server API is maintained as a parallel Python app.

## Top-Level Directories

- `apps/server-api`
  - the Python FastAPI backend
  - handles teacher/student auth, releases, progress reporting, AI hints, and dashboard APIs
- `apps/server-web`
  - the Vue teacher dashboard
  - handles teacher login, student management, release management, and the live classroom view
- `docs`
  - open source entry docs and engineering notes

## Product Boundary

- the maintained product is the server teaching edition of `Scratch AI Coach`
- the server track currently uses `Python FastAPI + Vue`
- the local client and networked desktop client live in separate repositories

## Where to Read Next

- first visit: [`../README.en.md`](../README.en.md)
- contribution workflow: [`../CONTRIBUTING.en.md`](../CONTRIBUTING.en.md)
- module responsibilities: [`./architecture.zh-CN.md`](./architecture.zh-CN.md)
- maintenance rules: [`./maintenance.zh-CN.md`](./maintenance.zh-CN.md)
