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
const previewProcessRef = {
  current: null,
}

try {
  await run()
} catch (error) {
  console.error(error instanceof Error ? error.message : error)
  process.exitCode = 1
} finally {
  await stopProcess(previewProcessRef.current)
}

async function run() {
  const port = await findFreePort(4173)
  const baseUrl = `http://${host}:${port}/`

  await runCommand('build', ['run', 'build'])

  const previewProcess = spawn(viteCommand(), ['preview', '--host', host, '--port', String(port), '--strictPort'], {
    cwd: appDir,
    env: buildEnv(),
    stdio: 'inherit',
  })

  previewProcessRef.current = previewProcess

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
      page.waitForURL((url) => url.pathname === '/dashboard'),
      page.getByRole('button', { name: '登录' }).click(),
    ])

    await waitForBodyIncludes(page, [
      '欢迎 王老师',
      '在册学生',
      '3',
      '1 / 2',
      '55%',
      'Ada',
      '72%',
      '补上广播消息后再测试一次',
      '第一期发布单',
      '已发布',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/students'),
      page.getByRole('link', { name: '学生管理' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '学生列表',
      'Ada',
      'Alan',
      'Mia',
      '四年级一班',
      '四年级二班',
      '72%',
      '38%',
      '55%',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/releases'),
      page.getByRole('link', { name: '发布单管理' }).click(),
    ])
    await waitForBodyIncludes(page, [
      '发布单列表',
      '第一期发布单',
      '第二期发布单',
      '已发布',
      '草稿',
      '24',
      '18',
    ])

    await Promise.all([
      page.waitForURL((url) => url.pathname === '/releases/rel-1/live'),
      page.getByRole('link', { name: '查看实时看板' }).first().click(),
    ])
    await waitForBodyIncludes(page, [
      '实时看板',
      '第一期发布单',
      'Ada',
      'Alan',
      '42%',
      '33%',
      '先把绿旗事件连起来',
      '先整理重复执行的脚本块',
    ])
    await waitForBodyIncludes(
      page,
      [
        '轮询中',
        '68%',
        '51%',
        '现在补上角色切换逻辑',
        '把等待和广播组合起来',
      ],
      9000,
    )

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

async function runCommand(label, args) {
  await new Promise((resolve, reject) => {
    const child = spawn(npmCommand(), args, {
      cwd: appDir,
      env: buildEnv(),
      stdio: 'inherit',
    })

    child.on('exit', (code, signal) => {
      if (code === 0) {
        resolve(undefined)
        return
      }

      reject(
        new Error(
          `${label} command failed with ${signal ? `signal ${signal}` : `code ${code}`}`,
        ),
      )
    })
    child.on('error', reject)
  })
}

function buildEnv() {
  return {
    ...process.env,
    VITE_SERVER_WEB_API_MODE: 'mock',
  }
}

function npmCommand() {
  return process.platform === 'win32' ? 'npm.cmd' : 'npm'
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
