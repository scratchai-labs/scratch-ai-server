package memory

import (
	"database/sql"
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

func TestStoreInitSchemaMigratesLegacySQLiteClassroomColumnsBeforeIndexes(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "legacy.sqlite3")
	legacyDB, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = legacyDB.Close()
	})

	legacyStatements := []string{
		`CREATE TABLE teachers (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'teacher',
			status TEXT NOT NULL DEFAULT 'active',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE classrooms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE students (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
			username TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE assignments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			goal TEXT NOT NULL,
			description TEXT NOT NULL,
			status TEXT NOT NULL,
			file_name TEXT NOT NULL,
			sb3_file_path TEXT NOT NULL,
			analysis_status TEXT NOT NULL,
			analysis_error_message TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
	}

	for _, statement := range legacyStatements {
		_, err = legacyDB.Exec(statement)
		require.NoError(t, err)
	}
	require.NoError(t, legacyDB.Close())

	store, err := NewStore(config.Config{
		DatabasePath:  dbPath,
		SB3StorageDir: t.TempDir(),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = store.sql.db.Close()
	})

	require.True(t, sqliteColumnExists(t, store.sql.db, "students", "classroom_id"))
	require.True(t, sqliteColumnExists(t, store.sql.db, "assignments", "classroom_id"))
	require.Equal(t, []int{1, 2, 3, 4, 5}, sqliteMigrationVersions(t, store.sql.db))
}

func TestStoreInitSchemaRecordsMigrationsForFreshSQLiteDatabase(t *testing.T) {
	store := newOpenStore(t)
	t.Cleanup(func() {
		_ = store.sql.db.Close()
	})

	require.Equal(t, []int{1, 2, 3, 4, 5}, sqliteMigrationVersions(t, store.sql.db))
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

func sqliteColumnExists(t *testing.T, db *sql.DB, tableName string, columnName string) bool {
	t.Helper()

	rows, err := db.Query("PRAGMA table_info(" + tableName + ")")
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		require.NoError(t, rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &pk))
		if name == columnName {
			return true
		}
	}

	require.NoError(t, rows.Err())
	return false
}

func sqliteMigrationVersions(t *testing.T, db *sql.DB) []int {
	t.Helper()

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version ASC")
	require.NoError(t, err)
	defer rows.Close()

	versions := make([]int, 0)
	for rows.Next() {
		var version int
		require.NoError(t, rows.Scan(&version))
		versions = append(versions, version)
	}

	require.NoError(t, rows.Err())
	return versions
}
