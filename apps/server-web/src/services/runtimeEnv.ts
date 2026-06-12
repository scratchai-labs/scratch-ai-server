export interface TeacherApiRuntime {
  mode: 'mock' | 'real'
  baseUrl: string
  showMockLoginHint: boolean
}

export interface TeacherApiRuntimeEnv {
  DEV: boolean
  PROD: boolean
  VITE_SERVER_WEB_API_MODE?: string
  VITE_SERVER_WEB_API_BASE_URL?: string
}

export function validateTeacherApiRuntimeEnv(
  env: TeacherApiRuntimeEnv,
  isProduction: boolean,
) {
  const mode = env.VITE_SERVER_WEB_API_MODE?.trim() === 'real' ? 'real' : 'mock'
  const baseUrl = env.VITE_SERVER_WEB_API_BASE_URL?.trim() ?? ''

  if (isProduction && mode !== 'real') {
    throw new Error('生产环境必须配置 VITE_SERVER_WEB_API_MODE=real')
  }
  if (isProduction && !baseUrl) {
    throw new Error('生产环境必须配置 VITE_SERVER_WEB_API_BASE_URL')
  }
}

export function resolveTeacherApiRuntime(
  env: TeacherApiRuntimeEnv,
  isProduction: boolean,
): TeacherApiRuntime {
  validateTeacherApiRuntimeEnv(env, isProduction)

  const mode = env.VITE_SERVER_WEB_API_MODE?.trim() === 'real' ? 'real' : 'mock'
  const baseUrl = env.VITE_SERVER_WEB_API_BASE_URL?.trim() ?? ''

  return {
    mode,
    baseUrl,
    showMockLoginHint: mode !== 'real',
  }
}
