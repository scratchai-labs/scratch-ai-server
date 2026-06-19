import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { setTimeout as delay } from 'node:timers/promises'
import net from 'node:net'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const appDir = path.resolve(scriptDir, '..')
const host = '127.0.0.1'
const serverProcessRef = {
  current: null,
}

try {
  await run()
} catch (error) {
  console.error(error instanceof Error ? error.message : error)
  process.exitCode = 1
} finally {
  await stopProcess(serverProcessRef.current)
}

async function run() {
  const port = await findFreePort(4173)
  const baseUrl = `http://${host}:${port}/`

  const serverProcess = spawn(
    viteCommand(),
    ['--host', host, '--port', String(port), '--strictPort'],
    {
      cwd: appDir,
      env: buildEnv(),
      stdio: 'inherit',
    },
  )

  serverProcessRef.current = serverProcess

  await waitForServer(baseUrl)
  await runBrowserSmoke(baseUrl)
}

async function runBrowserSmoke(baseUrl) {
  const browser = await chromium.launch({
    headless: true,
  })
  try {
    const context = await browser.newContext({
      viewport: {
        width: 1440,
        height: 960,
      },
    })
    const page = await context.newPage()
    const pageErrors = []
    const failedRequests = []

    page.on('pageerror', (error) => {
      pageErrors.push(error.message)
    })
    page.on('requestfailed', (request) => {
      failedRequests.push(
        `${request.failure()?.errorText ?? 'request failed'} ${request.url()}`,
      )
    })

    await page.goto(baseUrl, {
      waitUntil: 'domcontentloaded',
    })

    await page.waitForURL((url) => url.pathname === '/login')
    await page.getByLabel('账号').fill('teacher')
    await page.getByLabel('密码').fill('teach123')
    await Promise.all([
      page.waitForURL((url) => url.pathname === '/classes'),
      page.getByRole('button', { name: '登录' }).click(),
    ])

    await waitForBodyIncludes(page, [
      '班级管理',
      '新建班级',
      '班级列表',
      '四年级一班',
      '四年级二班',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/dashboard'),
      page.getByRole('link', { name: '实时总览' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '欢迎 王老师',
      '在册学生',
      '最新学生状态',
      'Ada',
      '72%',
      '最新发布单',
      '第一期发布单',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/classes'),
      page.getByRole('link', { name: '班级管理' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '班级列表',
      '四年级一班',
      '四年级二班',
      '2 名学生 · 1 个项目',
      '1 名学生 · 1 个项目',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/classes/class-1'),
      page.getByRole('link', { name: '进入班级' }).first().click(),
    ])
    await waitForBodyIncludes(page, [
      '学生管理',
      '项目管理',
      'Ada',
      '迷宫项目',
      '查看项目详情',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/projects/rel-1'),
      page.getByRole('link', { name: '查看项目详情' }).first().click(),
    ])
    await waitForBodyIncludes(page, [
      '迷宫项目',
      '项目概览',
      '学生当前进度与提示',
      'Ada',
      '先把绿旗事件连起来',
    ])

    if (pageErrors.length) {
      throw new Error(`Smoke test found page errors:\n${pageErrors.join('\n')}`)
    }

    if (failedRequests.length) {
      throw new Error(`Smoke test found failed requests:\n${failedRequests.join('\n')}`)
    }

    console.log('Mock smoke test passed.')
  } finally {
    await browser.close()
  }
}

function buildEnv() {
  return {
    ...process.env,
    VITE_SERVER_WEB_API_MODE: 'mock',
  }
}

function viteCommand() {
  return path.join(
    appDir,
    'node_modules',
    '.bin',
    process.platform === 'win32' ? 'vite.cmd' : 'vite',
  )
}

async function waitForServer(baseUrl, timeoutMs = 20000) {
  const deadline = Date.now() + timeoutMs
  let lastError = null

  while (Date.now() < deadline) {
    try {
      const response = await fetch(baseUrl)
      if (response.ok) {
        return
      }
      lastError = new Error(`unexpected status ${response.status}`)
    } catch (error) {
      lastError = error
    }

    await delay(250)
  }

  throw new Error(
    `Preview server did not become ready at ${baseUrl}: ${String(lastError)}`,
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

async function findFreePort(preferredPort) {
  const attempt = await checkPort(preferredPort)
  if (attempt) {
    return preferredPort
  }

  return await new Promise((resolve, reject) => {
    const server = net.createServer()

    server.listen(0, host, () => {
      const address = server.address()
      if (!address || typeof address === 'string') {
        reject(new Error('failed to allocate a dynamic port'))
        return
      }

      const { port } = address
      server.close((error) => {
        if (error) {
          reject(error)
          return
        }
        resolve(port)
      })
    })

    server.on('error', reject)
  })
}

async function checkPort(port) {
  return await new Promise((resolve) => {
    const server = net.createServer()

    server.once('error', () => resolve(false))
    server.listen(port, host, () => {
      server.close(() => resolve(true))
    })
  })
}
