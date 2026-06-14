import path from 'node:path'
import { describe, expect, it } from 'vitest'

import {
  buildRealSmokeApiEnv,
  buildRealSmokeWebEnv,
  createSampleSb3Archive,
} from './real-smoke-support.mjs'

describe('real smoke support', () => {
  it('builds isolated api and web env for local real-mode smoke', () => {
    const tempDir = '/tmp/teacher-real-smoke'
    const apiEnv = buildRealSmokeApiEnv({
      inheritedEnv: {
        PATH: '/usr/bin',
        DEEPSEEK_API_KEY: 'test-key',
        DEEPSEEK_MODEL: 'deepseek-test',
      },
      apiPort: 8001,
      webOrigin: 'http://127.0.0.1:4175',
      tempDir,
    })

    expect(apiEnv).toMatchObject({
      PATH: '/usr/bin',
      PORT: '8001',
      GIN_MODE: 'debug',
      SERVER_API_DB_PATH: path.join(tempDir, 'server-api.sqlite3'),
      SB3_STORAGE_DIR: path.join(tempDir, 'sb3-storage'),
      CORS_ALLOWED_ORIGINS: 'http://127.0.0.1:4175',
      DEEPSEEK_API_KEY: 'test-key',
      DEEPSEEK_MODEL: 'deepseek-test',
    })

    const webEnv = buildRealSmokeWebEnv({
      inheritedEnv: {
        PATH: '/usr/bin',
      },
      apiBaseUrl: 'http://127.0.0.1:8001',
    })

    expect(webEnv).toMatchObject({
      PATH: '/usr/bin',
      VITE_SERVER_WEB_API_MODE: 'real',
      VITE_SERVER_WEB_API_BASE_URL: 'http://127.0.0.1:8001',
    })
  })

  it('creates a sample sb3 archive with an embedded Scratch project', () => {
    const archive = createSampleSb3Archive()

    expect(archive.fileName).toBe('teacher-real-smoke.sb3')
    expect(archive.contentType).toBe('application/zip')
    expect(archive.buffer.subarray(0, 4).toString('binary')).toBe('PK\u0003\u0004')
    expect(archive.buffer.includes(Buffer.from('project.json'))).toBe(true)
    expect(archive.buffer.includes(Buffer.from('"Cat"'))).toBe(true)
    expect(archive.buffer.includes(Buffer.from('motion_movesteps'))).toBe(true)
  })
})
