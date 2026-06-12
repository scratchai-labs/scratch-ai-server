export interface ResponseLike {
  ok: boolean
  status: number
  json: () => Promise<unknown>
}

export type FetchLike = (
  input: RequestInfo | URL,
  init?: RequestInit,
) => Promise<ResponseLike>

export class HttpError extends Error {
  readonly status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'HttpError'
    this.status = status
  }
}

export function buildApiUrl(baseUrl: string | undefined, path: string): string {
  const normalizedBase = baseUrl?.trim()

  if (!normalizedBase) {
    return path
  }

  const baseWithSlash = normalizedBase.endsWith('/')
    ? normalizedBase
    : `${normalizedBase}/`

  return new URL(path.replace(/^\//, ''), baseWithSlash).toString()
}

export async function requestJson<T>(
  fetchImpl: FetchLike,
  url: string,
  init?: RequestInit,
  options: {
    onUnauthorized?: () => void | Promise<void>
  } = {},
): Promise<T> {
  const response = await fetchImpl(url, init)

  if (!response.ok) {
    const body = await safeJson(response)
    if (response.status === 401) {
      await options.onUnauthorized?.()
    }
    throw new HttpError(resolveErrorMessage(response.status, body), response.status)
  }

  return (await response.json()) as T
}

async function safeJson(response: ResponseLike): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    return null
  }
}

function resolveErrorMessage(status: number, body: unknown): string {
  if (typeof body === 'string' && body.trim()) {
    return body
  }

  if (body && typeof body === 'object') {
    const record = body as Record<string, unknown>
    if (typeof record.message === 'string' && record.message.trim()) {
      return record.message
    }
    if (typeof record.error === 'string' && record.error.trim()) {
      return record.error
    }
  }

  return `请求失败（${status}）`
}
