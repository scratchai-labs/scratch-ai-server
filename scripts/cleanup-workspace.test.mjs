import test from "node:test";
import assert from "node:assert/strict";
import os from "node:os";
import path from "node:path";
import { mkdtemp, mkdir, readFile, rm, writeFile } from "node:fs/promises";

import { cleanupWorkspace } from "./cleanup-workspace.mjs";

async function createWorkspaceFixture() {
  const repoRoot = await mkdtemp(path.join(os.tmpdir(), "scratchai-cleanup-"));

  await mkdir(path.join(repoRoot, "node_modules"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-web", "node_modules"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-web", "dist"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-web", "coverage"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-web", "src"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-api", ".venv"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-api", ".pytest_cache"), { recursive: true });
  await mkdir(path.join(repoRoot, "apps", "server-api", ".data"), { recursive: true });
  await mkdir(path.join(repoRoot, "docs", "assets", "screenshots"), { recursive: true });
  await mkdir(path.join(repoRoot, "tmp-demo"), { recursive: true });

  await writeFile(path.join(repoRoot, "apps", "server-web", "node_modules", ".keep"), "keep");
  await writeFile(path.join(repoRoot, "apps", "server-web", "dist", "bundle.js"), "bundle");
  await writeFile(path.join(repoRoot, "apps", "server-web", "coverage", "lcov.info"), "coverage");
  await writeFile(path.join(repoRoot, "apps", "server-web", "src", "main.ts"), "keep me");
  await writeFile(path.join(repoRoot, "apps", "server-api", ".coverage"), "coverage");
  await writeFile(path.join(repoRoot, "apps", "server-api", ".pytest_cache", "cache.db"), "cache");
  await writeFile(path.join(repoRoot, "apps", "server-api", ".data", "server-api.sqlite3"), "db");
  await writeFile(path.join(repoRoot, "apps", "server-api", ".venv", ".keep"), "venv");
  await writeFile(path.join(repoRoot, "docs", "assets", "screenshots", "shot.png"), "png");
  await writeFile(path.join(repoRoot, "tmp-demo", "tmp.txt"), "tmp");

  return repoRoot;
}

test("cleanupWorkspace dry-run reports generated artifacts without deleting them", async () => {
  const repoRoot = await createWorkspaceFixture();
  const logs = [];

  try {
    const result = await cleanupWorkspace({
      repoRoot,
      dryRun: true,
      log: entry => logs.push(entry)
    });

    assert.equal(result.failedPaths.length, 0);
    assert.ok(result.removedPaths.includes("node_modules"));
    assert.ok(result.removedPaths.includes("apps/server-web/node_modules"));
    assert.ok(result.removedPaths.includes("apps/server-web/dist"));
    assert.ok(result.removedPaths.includes("apps/server-web/coverage"));
    assert.ok(result.removedPaths.includes("apps/server-api/.venv"));
    assert.ok(result.removedPaths.includes("apps/server-api/.pytest_cache"));
    assert.ok(result.removedPaths.includes("apps/server-api/.coverage"));
    assert.ok(result.removedPaths.includes("apps/server-api/.data"));
    assert.ok(result.removedPaths.includes("tmp-demo"));
    assert.ok(result.removedPaths.includes("docs/assets/screenshots/shot.png"));

    const keptSource = await readFile(path.join(repoRoot, "apps", "server-web", "src", "main.ts"), "utf8");
    assert.equal(keptSource, "keep me");
    assert.ok(logs.some(entry => entry.includes("[dry-run] remove node_modules")));
  } finally {
    await rm(repoRoot, { recursive: true, force: true });
  }
});

test("cleanupWorkspace removes generated artifacts and preserves tracked source files", async () => {
  const repoRoot = await createWorkspaceFixture();

  try {
    const result = await cleanupWorkspace({ repoRoot });

    assert.equal(result.failedPaths.length, 0);

    await assert.rejects(() =>
      readFile(path.join(repoRoot, "apps", "server-web", "node_modules", ".keep"))
    );
    await assert.rejects(() => readFile(path.join(repoRoot, "apps", "server-api", ".coverage")));
    await assert.rejects(() => readFile(path.join(repoRoot, "docs", "assets", "screenshots", "shot.png")));
    await assert.rejects(() => readFile(path.join(repoRoot, "tmp-demo", "tmp.txt")));

    const keptSource = await readFile(path.join(repoRoot, "apps", "server-web", "src", "main.ts"), "utf8");

    assert.equal(keptSource, "keep me");
  } finally {
    await rm(repoRoot, { recursive: true, force: true });
  }
});

test("cleanupWorkspace skips missing paths instead of reporting fake removals", async () => {
  const repoRoot = await mkdtemp(path.join(os.tmpdir(), "scratchai-cleanup-empty-"));
  const logs = [];

  try {
    const result = await cleanupWorkspace({
      repoRoot,
      dryRun: true,
      log: entry => logs.push(entry)
    });

    assert.deepEqual(result.removedPaths, []);
    assert.equal(result.failedPaths.length, 0);
    assert.deepEqual(logs, ["Dry run finished."]);
  } finally {
    await rm(repoRoot, { recursive: true, force: true });
  }
});
