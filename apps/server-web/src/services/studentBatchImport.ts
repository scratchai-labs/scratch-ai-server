import type { CreateTeacherStudentInput } from './teacherApi'

export interface StudentBatchPasteRow {
  displayName: string
  username: string
  initialPassword: string
}

export interface StudentBatchCreateInputOptions {
  pastedText: string
  defaultPassword: string
  existingUsernames: string[]
}

export function buildStudentBatchTemplateCsv(): string {
  return '\uFEFF姓名,账号（可选）,初始密码（可选）\n小明,,\n小红,,\n'
}

export function parseStudentBatchPaste(text: string): StudentBatchPasteRow[] {
  const rows: StudentBatchPasteRow[] = []
  const lines = text.replace(/\r\n/g, '\n').split('\n')

  for (const [index, rawLine] of lines.entries()) {
    const line = rawLine.trim()
    if (!line) {
      continue
    }

    const columns = splitBatchLine(line)
    if (index === 0 && isBatchHeaderRow(columns)) {
      continue
    }

    const displayName = (columns[0] ?? '').trim()
    if (!displayName) {
      continue
    }

    rows.push({
      displayName,
      username: (columns[1] ?? '').trim(),
      initialPassword: (columns[2] ?? '').trim(),
    })
  }

  return rows
}

export function buildStudentBatchCreateInputs(
  options: StudentBatchCreateInputOptions,
): CreateTeacherStudentInput[] {
  const rows = parseStudentBatchPaste(options.pastedText)
  if (!rows.length) {
    throw new Error('请先粘贴至少一行学生数据')
  }

  const defaultPassword = options.defaultPassword.trim()
  const usernameAllocator = createStudentUsernameAllocator(options.existingUsernames)

  return rows.map((row, index) => {
    const displayName = row.displayName.trim()
    if (!displayName) {
      throw new Error(`第 ${index + 1} 行缺少学生姓名`)
    }

    const username = row.username.trim()
      ? usernameAllocator.reserve(row.username.trim())
      : usernameAllocator.next()

    const initialPassword = row.initialPassword.trim() || defaultPassword
    if (!initialPassword) {
      throw new Error('请填写统一初始密码，或在表格中补充初始密码')
    }

    return {
      username,
      displayName,
      initialPassword,
    }
  })
}

function splitBatchLine(line: string): string[] {
  if (line.includes('\t')) {
    return line.split('\t')
  }

  if (line.includes('，')) {
    return line.split('，')
  }

  if (line.includes(',')) {
    return line.split(',')
  }

  return [line]
}

function isBatchHeaderRow(columns: string[]): boolean {
  const first = normalizeHeaderCell(columns[0] ?? '')
  const second = normalizeHeaderCell(columns[1] ?? '')
  const third = normalizeHeaderCell(columns[2] ?? '')

  return (
    ['姓名', '学生姓名', 'displayname', 'name'].includes(first)
    || ['账号', '用户名', 'username'].includes(second)
    || ['密码', '初始密码', 'password', 'initialpassword'].includes(third)
  )
}

function normalizeHeaderCell(value: string): string {
  return value.trim().toLowerCase().replace(/\s+/g, '')
}

function createStudentUsernameAllocator(existingUsernames: string[]) {
  const taken = new Set(existingUsernames.map((username) => username.trim()).filter(Boolean))
  const pattern = /^student-(\d+)$/i
  let width = 2

  for (const username of taken) {
    const match = username.match(pattern)
    if (!match) {
      continue
    }

    const numericPart = match[1] ?? '0'
    const parsed = Number.parseInt(numericPart, 10)
    if (Number.isNaN(parsed)) {
      continue
    }

    width = Math.max(width, numericPart.length)
  }

  let nextNumber = 1

  return {
    reserve(username: string): string {
      const trimmed = username.trim()
      if (!trimmed) {
        throw new Error('学生账号不能为空')
      }
      if (taken.has(trimmed)) {
        throw new Error(`学生账号 ${trimmed} 已存在`)
      }
      taken.add(trimmed)
      return trimmed
    },
    next(): string {
      while (true) {
        const candidate = `student-${String(nextNumber).padStart(width, '0')}`
        nextNumber += 1
        if (!taken.has(candidate)) {
          taken.add(candidate)
          return candidate
        }
      }
    },
  }
}
