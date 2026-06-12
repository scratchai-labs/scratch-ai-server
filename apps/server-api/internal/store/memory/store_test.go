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
		Title:       "Broken Assignment",
		Goal:        "verify db failure",
		Description: "db closed",
		FileName:    "broken.sb3",
		SB3FilePath: "/tmp/broken.sb3",
		SB3Data:     []byte("abc"),
	})

	require.Error(t, err)
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
	assignmentRecord, err := store.CreateAssignment(teacherRecord.ID, CreateAssignmentInput{
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
