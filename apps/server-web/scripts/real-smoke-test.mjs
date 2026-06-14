import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { mkdtemp, mkdir, rm, writeFile } from 'node:fs/promises'
import net from 'node:net'
import os from 'node:os'
import path from 'node:path'
import process from 'node:process'
import { setTimeout as delay } from 'node:timers/promises'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

import {
  buildRealSmokeApiEnv,
  buildRealSmokeWebEnv,
  createSampleSb3Archive,
} from './real-smoke-support.mjs'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const appDir = path.resolve(scriptDir, '..')
const repoDir = path.resolve(appDir, '..', '..')
const apiDir = path.join(repoDir, 'apps', 'server-api')
const host = '127.0.0.1'
const childRefs = {
  api: null,
  web: null,
}
const tempDirRef = {
  current: '',
}

try {
  await run()
} catch (error) {
  console.error(error instanceof Error ? error.message : error)
  process.exitCode = 1
} finally {
  await stopProcess(childRefs.web)
  await stopProcess(childRefs.api)

  if (tempDirRef.current) {
    await rm(tempDirRef.current, { recursive: true, force: true })
  }
}

async function run() {
  const tag = Date.now().toString(36)
  const tempDir = await mkdtemp(
    path.join(os.tmpdir(), 'scratch-ai-teacher-real-smoke-'),
  )
  tempDirRef.current = tempDir

  const teacher = {
    username: `teacher-smoke-${tag}`,
    password: 'Teach12345!',
  }
  const student = {
    username: `student-smoke-${tag}`,
    displayName: '小测同学',
    initialPassword: 'Start12345!',
    resetPassword: 'Reset12345!',
  }
  const release = {
    title: `课堂冒烟任务-${tag}`,
    goal: '让 Cat 角色移动起来',
    description: 'real-mode 教师链路冒烟验证',
  }

  const apiPort = await findFreePort()
  const webPort = await findFreePort()
  const apiBaseUrl = `http://${host}:${apiPort}`
  const webBaseUrl = `http://${host}:${webPort}/`
  const webOrigin = `http://${host}:${webPort}`

  await mkdir(path.join(tempDir, 'sb3-storage'), { recursive: true })
  const sb3FilePath = await writeSampleSb3(tempDir)

  console.log(`[real-smoke] temp dir: ${tempDir}`)
  console.log(`[real-smoke] starting api at ${apiBaseUrl}`)
  childRefs.api = spawn('go', ['run', './cmd/server-api'], {
    cwd: apiDir,
    env: buildRealSmokeApiEnv({
      inheritedEnv: process.env,
      apiPort,
      webOrigin,
      tempDir,
    }),
    stdio: 'inherit',
    shell: process.platform === 'win32',
  })

  await waitForServer(`${apiBaseUrl}/health`)

  const teacherSession = await registerTeacher(apiBaseUrl, teacher)

  console.log(`[real-smoke] starting web at ${webBaseUrl}`)
  childRefs.web = spawn(
    viteCommand(),
    ['--host', host, '--port', String(webPort), '--strictPort'],
    {
      cwd: appDir,
      env: buildRealSmokeWebEnv({
        inheritedEnv: process.env,
        apiBaseUrl,
      }),
      stdio: 'inherit',
    },
  )

  await waitForServer(webBaseUrl)

  const hint = await runBrowserSmoke({
    apiBaseUrl,
    webBaseUrl,
    teacher,
    teacherToken: teacherSession.token,
    student,
    release,
    sb3FilePath,
    artifactDir: tempDir,
  })

  console.log(
    `[real-smoke] passed with hint provider "${hint.providerName}" and hint "${hint.hintText}"`,
  )
}

async function runBrowserSmoke({
  apiBaseUrl,
  webBaseUrl,
  teacher,
  teacherToken,
  student,
  release,
  sb3FilePath,
  artifactDir,
}) {
  const browser = await chromium.launch({ headless: true })
  let page = null

  try {
    const context = await browser.newContext({
      viewport: {
        width: 1440,
        height: 960,
      },
    })
    page = await context.newPage()

    const pageErrors = []
    const failedRequests = []
    const failedResponses = []

    page.on('pageerror', (error) => {
      pageErrors.push(error.message)
    })
    page.on('requestfailed', (request) => {
      const failure = request.failure()?.errorText ?? 'request failed'
      if (failure === 'net::ERR_ABORTED') {
        return
      }
      failedRequests.push(`${failure} ${request.url()}`)
    })
    page.on('response', (response) => {
      if (
        response.url().startsWith(apiBaseUrl)
        && response.status() >= 400
      ) {
        failedResponses.push(
          `${response.status()} ${response.request().method()} ${response.url()}`,
        )
      }
    })

    await page.goto(webBaseUrl, {
      waitUntil: 'domcontentloaded',
    })

    await page.waitForURL((url) => url.pathname === '/login')
    await page.getByLabel('账号').fill(teacher.username)
    await page.getByLabel('密码').fill(teacher.password)
    await Promise.all([
      page.waitForURL((url) => url.pathname === '/dashboard'),
      page.getByRole('button', { name: '登录' }).click(),
    ])
    await waitForBodyIncludes(page, [
      `欢迎 ${teacher.username}`,
      '在册学生',
      '发布单',
      '课堂最新状态',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/students'),
      page.getByRole('link', { name: '学生管理' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '新建学生',
      '学生列表',
      '创建学生',
    ])

    await page.locator('input[name="student-username"]').fill(student.username)
    await page.locator('input[name="student-display-name"]').fill(student.displayName)
    await page.locator('input[name="student-password"]').fill(student.initialPassword)
    await page.getByRole('button', { name: '创建学生' }).click()
    await waitForBodyIncludes(page, [`已创建学生账号 ${student.username}`])
    await waitForBodyIncludes(page, [student.username, student.displayName])

    const studentRow = page.locator('tbody tr').filter({ hasText: student.username }).first()
    await studentRow.locator('input[placeholder="输入新密码"]').fill(student.resetPassword)
    await studentRow.getByRole('button', { name: '重置密码' }).click()
    await waitForBodyIncludes(page, [`已重置 ${student.username} 的密码`])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/releases'),
      page.getByRole('link', { name: '发布单管理' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '上传参考 sb3',
      '发布单列表',
      '上传并创建发布单',
    ])

    await page.locator('input[name="release-title"]').fill(release.title)
    await page.locator('input[name="release-goal"]').fill(release.goal)
    await page.locator('textarea[name="release-description"]').fill(release.description)
    await page.locator('input[name="release-sb3"]').setInputFiles(sb3FilePath)
    await page.getByRole('button', { name: '上传并创建发布单' }).click()
    await waitForBodyIncludes(page, [`已上传发布单 ${release.title}`])

    const releaseState = await waitForReleaseReady(apiBaseUrl, teacherToken, release.title)
    const releaseCard = page.locator('.release-card').filter({ hasText: release.title }).first()
    await assertVisible(releaseCard, `发布单卡片 ${release.title}`)

    await releaseCard.getByRole('button', { name: '查看详情' }).click()
    await waitForBodyIncludes(page, [
      release.title,
      release.goal,
      release.description,
      '分析完成',
      'Stage',
      'Cat',
      '开始',
      '分数',
      '步骤列表',
      'pen',
    ], 15000)

    const assignLabel = page.locator('label').filter({ hasText: student.username }).first()
    await assignLabel.locator('input[type="checkbox"]').check()
    await page.getByRole('button', { name: '保存分配' }).click()
    await waitForBodyIncludes(page, ['已分配 1 名学生', `${student.displayName}（${student.username}）`])

    await page.getByRole('button', { name: '发布发布单' }).click()
    await waitForBodyIncludes(page, [`已发布 ${release.title}`])
    await waitForBodyIncludes(page, ['已发布'])

    await Promise.all([
      page.waitForURL(
        (url) => url.pathname === `/releases/${releaseState.releaseId}/live`,
      ),
      releaseCard.getByRole('link', { name: '查看实时看板' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '实时看板',
      release.title,
      student.displayName,
    ])

    const studentToken = await loginStudent(apiBaseUrl, {
      username: student.username,
      password: student.resetPassword,
    })
    await waitForStudentAssignment(apiBaseUrl, studentToken, release.title)
    await reportStudentProgress(apiBaseUrl, studentToken, releaseState.releaseId)
    const hint = await requestStudentHint(apiBaseUrl, studentToken, releaseState.releaseId)

    await waitForBodyIncludes(page, [
      student.displayName,
      '让 Cat 角色移动起来',
      '已经把事件积木接上了',
      hint.hintText,
    ], 15000)

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/releases'),
      page.getByRole('link', { name: '发布单管理' }).click(),
    ])
    await assertVisible(releaseCard, `发布单卡片 ${release.title}`)
    await releaseCard.getByRole('button', { name: '查看详情' }).click()
    await waitForBodyIncludes(page, [release.title, '分析完成'])

    await page.getByRole('button', { name: '归档发布单' }).click()
    await waitForBodyIncludes(page, [`已归档 ${release.title}`, '已归档'])

    assertNoBrowserFailures(pageErrors, failedRequests, failedResponses)
    return hint
  } catch (error) {
    if (page) {
      const screenshotPath = path.join(artifactDir, 'teacher-real-smoke-failure.png')
      await page.screenshot({
        path: screenshotPath,
        fullPage: true,
      }).catch(() => {})
      console.error(`[real-smoke] failure screenshot: ${screenshotPath}`)
    }
    throw error
  } finally {
    await browser.close()
  }
}

async function registerTeacher(apiBaseUrl, teacher) {
  return requestJson(`${apiBaseUrl}/api/teacher/register`, {
    method: 'POST',
    body: teacher,
  })
}

async function loginStudent(apiBaseUrl, student) {
  const session = await requestJson(`${apiBaseUrl}/api/student/login`, {
    method: 'POST',
    body: {
      ...student,
      clientType: 'desktop',
    },
  })

  return String(session.token)
}

async function waitForReleaseReady(apiBaseUrl, teacherToken, releaseTitle) {
  return waitForCondition(
    async () => {
      const payload = await requestJson(`${apiBaseUrl}/api/teacher/assignments`, {
        token: teacherToken,
      })
      const releaseRecord =
        payload.items?.find((item) => String(item.title) === releaseTitle) ?? null

      if (!releaseRecord) {
        return false
      }

      const releaseId = String(releaseRecord.id)
      const analysis = await requestJson(
        `${apiBaseUrl}/api/teacher/assignments/${releaseId}/analysis`,
        {
          token: teacherToken,
        },
      )

      if (String(analysis.analysisStatus) !== 'ready') {
        return false
      }

      return {
        releaseId,
        analysis,
      }
    },
    {
      description: `发布单 ${releaseTitle} 分析完成`,
      timeoutMs: 15000,
      intervalMs: 250,
    },
  )
}

async function waitForStudentAssignment(apiBaseUrl, studentToken, releaseTitle) {
  return waitForCondition(
    async () => {
      const payload = await requestJson(`${apiBaseUrl}/api/student/assignments`, {
        token: studentToken,
      })

      return payload.items?.some((item) => String(item.title) === releaseTitle)
    },
    {
      description: `学生任务 ${releaseTitle} 已发布可见`,
      timeoutMs: 8000,
      intervalMs: 250,
    },
  )
}

async function reportStudentProgress(apiBaseUrl, studentToken, releaseId) {
  return requestJson(
    `${apiBaseUrl}/api/student/assignments/${releaseId}/progress`,
    {
      method: 'POST',
      token: studentToken,
      body: {
        currentTarget: '让 Cat 角色移动起来',
        stepSummary: '已经把事件积木接上了',
        localProjectHash: 'real-smoke-hash',
        reportedAt: new Date().toISOString(),
        snapshot: {
          currentRoleName: 'Cat',
          roles: [
            {
              roleName: 'Stage',
              roleType: 'stage',
              blocks: ['当绿旗被点击', '广播消息 开始'],
            },
            {
              roleName: 'Cat',
              roleType: 'sprite',
              blocks: ['当接收到 开始', '移动 10 步'],
            },
          ],
        },
      },
    },
  )
}

async function requestStudentHint(apiBaseUrl, studentToken, releaseId) {
  return requestJson(
    `${apiBaseUrl}/api/student/assignments/${releaseId}/hints`,
    {
      method: 'POST',
      token: studentToken,
    },
  )
}

async function writeSampleSb3(tempDir) {
  const archive = createSampleSb3Archive()
  const targetPath = path.join(tempDir, archive.fileName)
  await writeFile(targetPath, archive.buffer)
  return targetPath
}

async function requestJson(
  url,
  {
    method = 'GET',
    token,
    body,
  } = {},
) {
  const response = await fetch(url, {
    method,
    headers: {
      ...(body === undefined ? {} : { 'Content-Type': 'application/json' }),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  const raw = await response.text()
  const payload = raw ? JSON.parse(raw) : null

  if (!response.ok) {
    const detail =
      typeof payload?.error === 'string'
        ? payload.error
        : raw || `unexpected status ${response.status}`
    throw new Error(`${method} ${url} failed (${response.status}): ${detail}`)
  }

  return payload
}

function assertNoBrowserFailures(pageErrors, failedRequests, failedResponses) {
  if (!pageErrors.length && !failedRequests.length && !failedResponses.length) {
    return
  }

  const problems = [
    ...pageErrors.map((message) => `pageerror ${message}`),
    ...failedRequests.map((message) => `requestfailed ${message}`),
    ...failedResponses.map((message) => `badresponse ${message}`),
  ]

  throw new Error(`Smoke test found browser failures:\n${problems.join('\n')}`)
}

async function assertVisible(locator, description) {
  try {
    await locator.waitFor({
      state: 'visible',
      timeout: 10000,
    })
  } catch (error) {
    throw new Error(`${description} 未出现: ${error instanceof Error ? error.message : error}`)
  }
}

async function waitForServer(baseUrl, timeoutMs = 30000) {
  await waitForCondition(
    async () => {
      const response = await fetch(baseUrl)
      return response.ok
    },
    {
      description: `${baseUrl} 可访问`,
      timeoutMs,
      intervalMs: 250,
    },
  )
}

async function waitForCondition(
  check,
  {
    description,
    timeoutMs = 10000,
    intervalMs = 200,
  },
) {
  const deadline = Date.now() + timeoutMs
  let lastError = null

  while (Date.now() < deadline) {
    try {
      const result = await check()
      if (result) {
        return result
      }
    } catch (error) {
      lastError = error
    }

    await delay(intervalMs)
  }

  throw new Error(
    `Timed out waiting for ${description}${lastError ? `: ${String(lastError)}` : ''}`,
  )
}

async function waitForBodyIncludes(page, snippets, timeoutMs = 5000) {
  const deadline = Date.now() + timeoutMs
  let bodyText = ''

  while (Date.now() < deadline) {
    bodyText = await page.locator('body').innerText()
    if (snippets.every((snippet) => bodyText.includes(snippet))) {
      return
    }
    await delay(200)
  }

  assert.fail(
    `Timed out waiting for page text: ${snippets.join(', ')}\nCurrent body:\n${bodyText}`,
  )
}

function viteCommand() {
  return path.join(
    appDir,
    'node_modules',
    '.bin',
    process.platform === 'win32' ? 'vite.cmd' : 'vite',
  )
}

async function stopProcess(child) {
  if (!child || child.killed) {
    return
  }

  child.kill('SIGTERM')

  await Promise.race([
    new Promise((resolve) => child.once('exit', resolve)),
    delay(3000).then(() => {
      if (!child.killed) {
        child.kill('SIGKILL')
      }
    }),
  ])
}

async function findFreePort() {
  return new Promise((resolve, reject) => {
    const server = net.createServer()

    server.listen(0, host, () => {
      const address = server.address()
      if (!address || typeof address === 'string') {
        reject(new Error('failed to allocate a dynamic port'))
        return
      }

      server.close((error) => {
        if (error) {
          reject(error)
          return
        }
        resolve(address.port)
      })
    })

    server.once('error', reject)
  })
}
