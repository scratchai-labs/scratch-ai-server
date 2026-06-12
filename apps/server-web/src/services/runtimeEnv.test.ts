import { describe, expect, it } from 'vitest'
import {
  resolveTeacherApiRuntime,
  validateTeacherApiRuntimeEnv,
} from './runtimeEnv'

describe('resolveTeacherApiRuntime', () => {
  it('accepts real mode with an explicit API base url in production', () => {
    expect(() =>
      validateTeacherApiRuntimeEnv(
        {
          DEV: false,
          PROD: true,
          VITE_SERVER_WEB_API_MODE: 'real',
          VITE_SERVER_WEB_API_BASE_URL: 'https://api.example',
        },
        true,
      ),
    ).not.toThrow()
  })

  it('requires real mode in production', () => {
    expect(() =>
      resolveTeacherApiRuntime(
        {
          DEV: false,
          PROD: true,
          VITE_SERVER_WEB_API_MODE: 'mock',
          VITE_SERVER_WEB_API_BASE_URL: 'https://api.example',
        },
        true,
      ),
    ).toThrow('VITE_SERVER_WEB_API_MODE=real')
  })

  it('requires API base url in production real mode', () => {
    expect(() =>
      resolveTeacherApiRuntime(
        {
          DEV: false,
          PROD: true,
          VITE_SERVER_WEB_API_MODE: 'real',
          VITE_SERVER_WEB_API_BASE_URL: '',
        },
        true,
      ),
    ).toThrow('VITE_SERVER_WEB_API_BASE_URL')
  })

  it('defaults to mock mode during local development', () => {
    expect(
      resolveTeacherApiRuntime(
        {
          DEV: true,
          PROD: false,
          VITE_SERVER_WEB_API_MODE: undefined,
          VITE_SERVER_WEB_API_BASE_URL: undefined,
        },
        false,
      ),
    ).toEqual({
      mode: 'mock',
      baseUrl: '',
      showMockLoginHint: true,
    })
  })
})
