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
      page.waitForURL((url) => url.pathname === '/classes'),
      page.getByRole('button', { name: '登录' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '班级管理',
      '新建班级',
      '班级列表',
    ])

    await page.locator('input').first().fill('real smoke 班级')
    await page.getByRole('button', { name: '创建班级' }).click()
    await waitForBodyIncludes(page, ['real smoke 班级'])

    const classCard = page.locator('.release-card').filter({ hasText: 'real smoke 班级' }).first()
    await assertVisible(classCard, '新建班级卡片')
    await Promise.all([
      page.waitForURL((url) => url.pathname.startsWith('/classes/')),
      classCard.getByRole('link', { name: '进入班级' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '学生管理',
      '项目管理',
      '批量导入学生',
      '创建项目',
    ])

    const studentInputs = page.locator('input')
    await studentInputs.nth(0).fill(student.username)
    await studentInputs.nth(1).fill(student.displayName)
    await studentInputs.nth(2).fill(student.initialPassword)
    await page.getByRole('button', { name: '创建学生' }).click()
    await waitForBodyIncludes(page, [student.username, student.displayName])

    await page.locator('input[type="file"]').setInputFiles(sb3FilePath)
    await studentInputs.nth(4).fill(release.title)
    await studentInputs.nth(5).fill(release.goal)
    await studentInputs.nth(6).fill(release.description)
    await page.getByRole('button', { name: '创建项目' }).click()
    await waitForBodyIncludes(page, [release.title, 'pending'])

    const releaseState = await waitForReleaseReady(apiBaseUrl, teacherToken, release.title)
    await page.reload({ waitUntil: 'domcontentloaded' })
    await waitForBodyIncludes(page, [release.title, '查看项目详情'])

    const projectCard = page.locator('.release-card').filter({ hasText: release.title }).first()
    await assertVisible(projectCard, `项目卡片 ${release.title}`)
    await Promise.all([
      page.waitForURL((url) => url.pathname === `/projects/${releaseState.releaseId}`),
      projectCard.getByRole('link', { name: '查看项目详情' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '项目概览',
      '学生当前进度与提示',
      release.title,
    ], 15000)

    const studentToken = await loginStudent(apiBaseUrl, {
      username: student.username,
      password: student.initialPassword,
    })
    await assignAndPublish(apiBaseUrl, teacherToken, releaseState.releaseId, student.username)
    await waitForStudentAssignment(apiBaseUrl, studentToken, release.title)
    await reportStudentProgress(apiBaseUrl, studentToken, releaseState.releaseId)
    const hint = await requestStudentHint(apiBaseUrl, studentToken, releaseState.releaseId)
    await page.reload({ waitUntil: 'domcontentloaded' })

    await waitForBodyIncludes(page, [
      student.displayName,
      '让 Cat 角色移动起来',
      '已经把事件积木接上了',
      hint.hintText,
    ], 15000)

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

async function assignAndPublish(apiBaseUrl, teacherToken, releaseId, studentUsername) {
  const studentsPayload = await requestJson(`${apiBaseUrl}/api/teacher/students`, {
    token: teacherToken,
  })
  const studentRecord =
    studentsPayload.items?.find((item) => String(item.username) === studentUsername) ?? null

  if (!studentRecord) {
    throw new Error(`student ${studentUsername} not found for assignment`)
  }

  await requestJson(`${apiBaseUrl}/api/teacher/assignments/${releaseId}/assign-students`, {
    method: 'POST',
    token: teacherToken,
    body: {
      studentIds: [Number(studentRecord.id)],
    },
  })

  await requestJson(`${apiBaseUrl}/api/teacher/assignments/${releaseId}/publish`, {
    method: 'POST',
    token: teacherToken,
  })
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
