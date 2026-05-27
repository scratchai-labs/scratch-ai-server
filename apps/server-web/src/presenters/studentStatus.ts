export interface StudentStatusLike {
  progress: number
  status?: string
}

export function studentStatusTone(status?: string) {
  if (status === 'active') {
    return 'info'
  }

  if (status === 'assigned') {
    return 'warning'
  }

  return 'muted'
}

export function studentStatusLabel(student: StudentStatusLike) {
  if (student.progress > 0) {
    return `${student.progress}%`
  }

  if (student.status === 'active') {
    return '已上报'
  }

  if (student.status === 'assigned') {
    return '已分配'
  }

  return '等待中'
}
