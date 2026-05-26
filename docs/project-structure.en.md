# Project Structure

This repository is the split-out server monorepo. The formal server track is now `Go API + Teacher Web`, while the older `Python FastAPI` code remains only as a transition prototype.

## Top-Level Directories

- `apps/server-api`
  - the Go backend
  - handles teacher/student auth, releases, progress reporting, AI hints, and dashboard APIs
- `apps/server-web`
  - the Vue teacher dashboard
  - handles teacher login, student management, release management, and the live classroom view
- `docs`
  - open source entry docs and engineering notes

## Product Boundary

- the maintained product is the server teaching edition of `Scratch AI Coach`
- the server track currently uses `Go + Vue`
- the local client and networked desktop client live in separate repositories

## Current Implementation Notes

- `apps/server-api` uses `Gin`
- local development uses `SQLite` by default
- the API switches to `Postgres` when `DATABASE_URL` is configured
- the teacher web app can run against mock data or the real API

## Where to Read Next

- first visit: [`../README.en.md`](../README.en.md)
- contribution workflow: [`../CONTRIBUTING.en.md`](../CONTRIBUTING.en.md)
- module responsibilities: [`./architecture.zh-CN.md`](./architecture.zh-CN.md)
- maintenance rules: [`./maintenance.zh-CN.md`](./maintenance.zh-CN.md)
