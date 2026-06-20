export function toErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message.trim()) {
    if (isNetworkFetchError(error.message)) {
      return '无法连接到服务器，请检查 API 地址、CORS 白名单，或确认当前页面与 API 都已正确启用 HTTPS。'
    }
    return error.message
  }

  return fallback
}

function isNetworkFetchError(message: string): boolean {
  const normalized = message.trim().toLowerCase()
  return (
    normalized === 'failed to fetch' ||
    normalized === 'networkerror when attempting to fetch resource.'
  )
}
