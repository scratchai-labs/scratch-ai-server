import { describe, expect, it } from 'vitest'
import {
  buildStudentBatchCreateInputs,
  parseStudentBatchPaste,
  studentBatchTemplate,
} from './studentBatchImport'

describe('studentBatchImport', () => {
  it('exposes the xlsx template download metadata', () => {
    expect(studentBatchTemplate.href).toBe('/student-batch-template.xlsx')
    expect(studentBatchTemplate.downloadName).toBe('学生批量导入模板.xlsx')
    expect(studentBatchTemplate.firstDataRow).toBe(8)
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
