import { describe, expect, it } from 'vitest'
import {
  buildStudentBatchCreateInputs,
  buildStudentBatchTemplateCsv,
  parseStudentBatchPaste,
} from './studentBatchImport'

describe('studentBatchImport', () => {
  it('builds an excel-friendly csv template', () => {
    const csv = buildStudentBatchTemplateCsv()

    expect(csv.startsWith('\uFEFF')).toBe(true)
    expect(csv).toContain('姓名,账号（可选）,初始密码（可选）')
  })

  it('parses pasted spreadsheet rows with an optional header', () => {
    expect(
      parseStudentBatchPaste(
        '姓名\t账号\t初始密码\n小明\tstudent-01\tabc12345\n小红\t\t',
      ),
    ).toEqual([
      {
        displayName: '小明',
        username: 'student-01',
        initialPassword: 'abc12345',
      },
      {
        displayName: '小红',
        username: '',
        initialPassword: '',
      },
    ])
  })

  it('fills missing usernames and passwords while avoiding existing usernames', () => {
    expect(
      buildStudentBatchCreateInputs({
        pastedText: '姓名\n小明\n小红',
        defaultPassword: 'abc12345',
        existingUsernames: ['student-01', 'student-03'],
      }),
    ).toEqual([
      {
        username: 'student-02',
        displayName: '小明',
        initialPassword: 'abc12345',
      },
      {
        username: 'student-04',
        displayName: '小红',
        initialPassword: 'abc12345',
      },
    ])
  })
})
