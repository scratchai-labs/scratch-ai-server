package memory

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

func TestStoreCreateAssignmentReturnsWriteError(t *testing.T) {
	store := newClosedStore(t)

	_, err := store.CreateAssignment(1, CreateAssignmentInput{
		ClassroomID: 1,
		Title:       "Broken Assignment",
		Goal:        "verify db failure",
		Description: "db closed",
		FileName:    "broken.sb3",
		SB3FilePath: "/tmp/broken.sb3",
		SB3Data:     []byte("abc"),
	})

	require.Error(t, err)
}

func TestStoreCreateStudentUsesDefaultClassroomWhenClassroomIDMissing(t *testing.T) {
	store := newOpenStore(t)
	teacherRecord, err := store.CreateTeacher("teacher-default-student", "hash")
	require.NoError(t, err)

	studentRecord, err := store.CreateStudent(teacherRecord.ID, CreateStudentInput{
		Username:     "student-default-classroom",
		DisplayName:  "默认学生",
		PasswordHash: "password-hash",
	})
	require.NoError(t, err)
	require.NotZero(t, studentRecord.ClassroomID)

	classroomRecord, ok := store.GetClassroomByTeacher(teacherRecord.ID, studentRecord.ClassroomID)
	require.True(t, ok)
	require.Equal(t, "默认班级", classroomRecord.Name)
}

func TestStoreCreateAssignmentUsesDefaultClassroomWhenClassroomIDMissing(t *testing.T) {
	store := newOpenStore(t)
	teacherRecord, err := store.CreateTeacher("teacher-default-assignment", "hash")
	require.NoError(t, err)

	assignmentRecord, err := store.CreateAssignment(teacherRecord.ID, CreateAssignmentInput{
		Title:       "默认任务",
		Goal:        "默认班级兼容",
		Description: "legacy create path",
		FileName:    "default.sb3",
		SB3FilePath: "/tmp/default.sb3",
		SB3Data:     []byte("abc"),
	})
	require.NoError(t, err)
	require.NotZero(t, assignmentRecord.ClassroomID)

	classroomRecord, ok := store.GetClassroomByTeacher(teacherRecord.ID, assignmentRecord.ClassroomID)
	require.True(t, ok)
	require.Equal(t, "默认班级", classroomRecord.Name)
}

func TestStoreSaveTeacherTokenReturnsWriteError(t *testing.T) {
	store := newClosedStore(t)

	err := store.SaveTeacherToken("teacher-token", 1)

	require.Error(t, err)
}

func TestStoreSetAssignmentAnalysisReadyReturnsWriteError(t *testing.T) {
	store := newOpenStore(t)
	teacherRecord, err := store.CreateTeacher("teacher-store-test", "hash")
	require.NoError(t, err)
	classroomRecord, err := store.CreateClassroom(teacherRecord.ID, CreateClassroomInput{
		Name: "默认班级",
	})
	require.NoError(t, err)
	assignmentRecord, err := store.CreateAssignment(teacherRecord.ID, CreateAssignmentInput{
		ClassroomID: classroomRecord.ID,
		Title:       "Ready Assignment",
		Goal:        "verify analysis persistence",
		Description: "db close after create",
		FileName:    "ready.sb3",
		SB3FilePath: "/tmp/ready.sb3",
		SB3Data:     []byte("abc"),
	})
	require.NoError(t, err)
	require.NoError(t, store.sql.db.Close())

	err = store.SetAssignmentAnalysisReady(assignmentRecord.ID, AssignmentAnalysis{
		RoleNames: []string{"Cat"},
	})

	require.Error(t, err)
}

func TestStoreCreateHintReturnsWriteError(t *testing.T) {
	store := newClosedStore(t)

	_, err := store.CreateHint(CreateHintInput{
		AssignmentID:     1,
		StudentID:        2,
		ProgressReportID: 3,
		PromptInput:      map[string]any{"assignmentTitle": "A"},
		HintText:         "do the next step",
		ProviderName:     "fallback",
	})

	require.Error(t, err)
}

func newOpenStore(t *testing.T) *Store {
	t.Helper()

	store, err := NewStore(config.Config{
		DatabasePath:  filepath.Join(t.TempDir(), "server-api.sqlite3"),
		SB3StorageDir: t.TempDir(),
	})
	require.NoError(t, err)
	return store
}

func newClosedStore(t *testing.T) *Store {
	t.Helper()

	store := newOpenStore(t)
	require.NoError(t, store.sql.db.Close())
	return store
}
