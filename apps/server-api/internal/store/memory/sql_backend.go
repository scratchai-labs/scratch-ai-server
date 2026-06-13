package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/scratchai-labs/scratch-ai-server/apps/server-api/internal/config"
)

type sqlBackend struct {
	db      *sql.DB
	dialect string
}

func newSQLBackend(cfg config.Config) (*sqlBackend, error) {
	var (
		db      *sql.DB
		dialect string
		err     error
	)

	if strings.TrimSpace(cfg.DatabaseURL) != "" {
		dialect = "postgres"
		db, err = sql.Open("pgx", cfg.DatabaseURL)
	} else {
		dialect = "sqlite"
		if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0o755); err != nil {
			return nil, err
		}
		db, err = sql.Open("sqlite", cfg.DatabasePath)
		if err == nil {
			db.SetMaxOpenConns(1)
		}
	}
	if err != nil {
		return nil, err
	}

	backend := &sqlBackend{
		db:      db,
		dialect: dialect,
	}
	if err := backend.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return backend, nil
}

func (b *sqlBackend) initSchema() error {
	statements := sqliteSchemaStatements
	if b.dialect == "postgres" {
		statements = postgresSchemaStatements
	}

	for _, statement := range statements {
		if _, err := b.db.Exec(statement); err != nil {
			return err
		}
	}
	if err := b.ensureTeacherColumns(); err != nil {
		return err
	}
	return b.ensureAssignmentAnalysisColumns()
}

func (b *sqlBackend) CreateTeacher(username string, passwordHash string) (Teacher, error) {
	return b.CreateTeacherWithRole(username, passwordHash, "teacher", "active")
}

func (b *sqlBackend) CreateTeacherWithRole(username string, passwordHash string, role string, status string) (Teacher, error) {
	if _, ok := b.FindTeacherByUsername(username); ok {
		return Teacher{}, ErrTeacherConflict
	}

	now := nowUTC()
	id, err := b.insertReturningID(
		"INSERT INTO teachers (username, password_hash, role, status, created_at) VALUES (?, ?, ?, ?, ?)",
		username,
		passwordHash,
		role,
		status,
		now,
	)
	if err != nil {
		return Teacher{}, err
	}

	return Teacher{
		ID:           id,
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		Status:       status,
		CreatedAt:    parseTime(now),
	}, nil
}

func (b *sqlBackend) EnsureTeacher(username string, passwordHash string, role string, status string) (Teacher, error) {
	if existing, ok := b.FindTeacherByUsername(username); ok {
		_, err := b.db.Exec(
			b.rebind("UPDATE teachers SET password_hash = ?, role = ?, status = ? WHERE id = ?"),
			passwordHash,
			role,
			status,
			existing.ID,
		)
		if err != nil {
			return Teacher{}, err
		}
		existing.PasswordHash = passwordHash
		existing.Role = role
		existing.Status = status
		return existing, nil
	}

	return b.CreateTeacherWithRole(username, passwordHash, role, status)
}

func (b *sqlBackend) FindTeacherByUsername(username string) (Teacher, bool) {
	row := b.db.QueryRow(
		b.rebind("SELECT id, username, password_hash, role, status, created_at FROM teachers WHERE username = ?"),
		username,
	)
	record, err := scanTeacher(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Teacher{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) SaveTeacherToken(token string, teacherID int64) error {
	_, err := b.db.Exec(
		b.rebind("INSERT INTO teacher_sessions (teacher_id, token, expires_at, created_at) VALUES (?, ?, ?, ?)"),
		teacherID,
		token,
		nil,
		nowUTC(),
	)
	return err
}

func (b *sqlBackend) DeleteTeacherToken(token string) error {
	_, err := b.db.Exec(b.rebind("DELETE FROM teacher_sessions WHERE token = ?"), token)
	return err
}

func (b *sqlBackend) FindTeacherByToken(token string) (Teacher, bool) {
	row := b.db.QueryRow(
		b.rebind(`
			SELECT t.id, t.username, t.password_hash, t.role, t.status, t.created_at
			FROM teacher_sessions s
			JOIN teachers t ON t.id = s.teacher_id
			WHERE s.token = ?
		`),
		token,
	)
	record, err := scanTeacher(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Teacher{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) ListTeachers() []Teacher {
	rows, err := b.db.Query(
		b.rebind("SELECT id, username, password_hash, role, status, created_at FROM teachers ORDER BY id ASC"),
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	records := make([]Teacher, 0)
	for rows.Next() {
		record, scanErr := scanTeacher(rows)
		if scanErr == nil {
			records = append(records, record)
		}
	}
	return records
}

func (b *sqlBackend) GetTeacherByID(teacherID int64) (Teacher, bool) {
	row := b.db.QueryRow(
		b.rebind("SELECT id, username, password_hash, role, status, created_at FROM teachers WHERE id = ?"),
		teacherID,
	)
	record, err := scanTeacher(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Teacher{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) UpdateTeacherPassword(teacherID int64, passwordHash string) (Teacher, error) {
	if _, ok := b.GetTeacherByID(teacherID); !ok {
		return Teacher{}, ErrTeacherNotFound
	}

	_, err := b.db.Exec(
		b.rebind("UPDATE teachers SET password_hash = ? WHERE id = ?"),
		passwordHash,
		teacherID,
	)
	if err != nil {
		return Teacher{}, err
	}

	teacher, _ := b.GetTeacherByID(teacherID)
	return teacher, nil
}

func (b *sqlBackend) UpdateTeacherStatus(teacherID int64, status string) (Teacher, error) {
	if _, ok := b.GetTeacherByID(teacherID); !ok {
		return Teacher{}, ErrTeacherNotFound
	}

	_, err := b.db.Exec(
		b.rebind("UPDATE teachers SET status = ? WHERE id = ?"),
		status,
		teacherID,
	)
	if err != nil {
		return Teacher{}, err
	}

	teacher, _ := b.GetTeacherByID(teacherID)
	return teacher, nil
}

func (b *sqlBackend) CreateStudent(teacherID int64, input CreateStudentInput) (Student, error) {
	if _, ok := b.FindStudentByUsername(input.Username); ok {
		return Student{}, ErrStudentConflict
	}

	now := nowUTC()
	id, err := b.insertReturningID(
		"INSERT INTO students (teacher_id, username, display_name, password_hash, status, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		teacherID,
		input.Username,
		input.DisplayName,
		input.PasswordHash,
		"active",
		now,
	)
	if err != nil {
		return Student{}, err
	}

	return Student{
		ID:           id,
		TeacherID:    teacherID,
		Username:     input.Username,
		DisplayName:  input.DisplayName,
		PasswordHash: input.PasswordHash,
		Status:       "active",
		CreatedAt:    parseTime(now),
	}, nil
}

func (b *sqlBackend) ListStudentsByTeacher(teacherID int64) []Student {
	rows, err := b.db.Query(
		b.rebind("SELECT id, teacher_id, username, display_name, password_hash, status, created_at FROM students WHERE teacher_id = ? ORDER BY id ASC"),
		teacherID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	records := make([]Student, 0)
	for rows.Next() {
		record, scanErr := scanStudent(rows)
		if scanErr == nil {
			records = append(records, record)
		}
	}
	return records
}

func (b *sqlBackend) FindStudentByUsername(username string) (Student, bool) {
	row := b.db.QueryRow(
		b.rebind("SELECT id, teacher_id, username, display_name, password_hash, status, created_at FROM students WHERE username = ?"),
		username,
	)
	record, err := scanStudent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Student{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) SaveStudentToken(token string, studentID int64) error {
	_, err := b.db.Exec(
		b.rebind("INSERT INTO student_sessions (student_id, token, client_type, expires_at, created_at) VALUES (?, ?, ?, ?, ?)"),
		studentID,
		token,
		"desktop",
		nil,
		nowUTC(),
	)
	return err
}

func (b *sqlBackend) DeleteStudentToken(token string) error {
	_, err := b.db.Exec(b.rebind("DELETE FROM student_sessions WHERE token = ?"), token)
	return err
}

func (b *sqlBackend) FindStudentByToken(token string) (Student, bool) {
	row := b.db.QueryRow(
		b.rebind(`
			SELECT s.id, s.teacher_id, s.username, s.display_name, s.password_hash, s.status, s.created_at
			FROM student_sessions ss
			JOIN students s ON s.id = ss.student_id
			WHERE ss.token = ?
		`),
		token,
	)
	record, err := scanStudent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Student{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) GetStudentByTeacher(teacherID int64, studentID int64) (Student, bool) {
	row := b.db.QueryRow(
		b.rebind("SELECT id, teacher_id, username, display_name, password_hash, status, created_at FROM students WHERE teacher_id = ? AND id = ?"),
		teacherID,
		studentID,
	)
	record, err := scanStudent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Student{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) UpdateStudentPassword(teacherID int64, studentID int64, passwordHash string) (Student, error) {
	if _, ok := b.GetStudentByTeacher(teacherID, studentID); !ok {
		return Student{}, ErrStudentNotFound
	}

	_, err := b.db.Exec(
		b.rebind("UPDATE students SET password_hash = ? WHERE teacher_id = ? AND id = ?"),
		passwordHash,
		teacherID,
		studentID,
	)
	if err != nil {
		return Student{}, err
	}

	record, ok := b.GetStudentByTeacher(teacherID, studentID)
	if !ok {
		return Student{}, ErrStudentNotFound
	}
	return record, nil
}

func (b *sqlBackend) CreateAssignment(teacherID int64, input CreateAssignmentInput) (Assignment, error) {
	now := nowUTC()
	id, err := b.insertReturningID(
		"INSERT INTO assignments (teacher_id, title, goal, description, status, file_name, sb3_file_path, analysis_status, analysis_error_message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		teacherID,
		input.Title,
		input.Goal,
		input.Description,
		"draft",
		input.FileName,
		input.SB3FilePath,
		"pending",
		"",
		now,
		now,
	)
	if err != nil {
		return Assignment{}, err
	}

	record, ok := b.getAssignmentByFilters("a.teacher_id = ? AND a.id = ?", teacherID, id)
	if !ok {
		return Assignment{}, sql.ErrNoRows
	}
	return record, nil
}

func (b *sqlBackend) GetAssignmentByTeacher(teacherID int64, assignmentID int64) (Assignment, bool) {
	return b.getAssignmentByFilters("a.teacher_id = ? AND a.id = ?", teacherID, assignmentID)
}

func (b *sqlBackend) ListAssignmentsByTeacher(teacherID int64) []Assignment {
	return b.listAssignmentsByFilters("a.teacher_id = ?", teacherID)
}

func (b *sqlBackend) ListAssignmentsPendingAnalysis() []Assignment {
	return b.listAssignmentsByFilters("a.analysis_status IN (?, ?)", "pending", "processing")
}

func (b *sqlBackend) SetAssignmentAnalysisProcessing(assignmentID int64) error {
	_, err := b.db.Exec(
		b.rebind("UPDATE assignments SET analysis_status = ?, updated_at = ? WHERE id = ?"),
		"processing",
		nowUTC(),
		assignmentID,
	)
	return err
}

func (b *sqlBackend) SetAssignmentAnalysisReady(assignmentID int64, analysis AssignmentAnalysis) error {
	now := nowUTC()
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		b.rebind(`
			INSERT INTO assignment_analysis (
				assignment_id,
				role_names_json,
				script_counts_json,
				block_counts_json,
				category_counts_json,
				broadcast_messages_json,
				variable_names_json,
				list_names_json,
				extensions_json,
				teaching_points_json,
				created_at,
				updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT (assignment_id) DO UPDATE SET
					role_names_json = excluded.role_names_json,
					script_counts_json = excluded.script_counts_json,
					block_counts_json = excluded.block_counts_json,
					category_counts_json = excluded.category_counts_json,
				broadcast_messages_json = excluded.broadcast_messages_json,
				variable_names_json = excluded.variable_names_json,
				list_names_json = excluded.list_names_json,
				extensions_json = excluded.extensions_json,
				teaching_points_json = excluded.teaching_points_json,
				updated_at = excluded.updated_at
		`),
		assignmentID,
		mustJSON(analysis.RoleNames),
		mustJSON(analysis.ScriptCounts),
		mustJSON(analysis.BlockCounts),
		mustJSON(analysis.CategoryCounts),
		mustJSON(analysis.BroadcastMessages),
		mustJSON(analysis.VariableNames),
		mustJSON(analysis.ListNames),
		mustJSON(analysis.Extensions),
		mustJSON(analysis.TeachingPoints),
		now,
		now,
	); err != nil {
		return err
	}

	if _, err := tx.Exec(
		b.rebind("UPDATE assignments SET analysis_status = ?, analysis_error_message = ?, updated_at = ? WHERE id = ?"),
		"ready",
		"",
		now,
		assignmentID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (b *sqlBackend) SetAssignmentAnalysisFailed(assignmentID int64, message string) error {
	_, err := b.db.Exec(
		b.rebind("UPDATE assignments SET analysis_status = ?, analysis_error_message = ?, updated_at = ? WHERE id = ?"),
		"failed",
		message,
		nowUTC(),
		assignmentID,
	)
	return err
}

func (b *sqlBackend) AssignStudents(teacherID int64, assignmentID int64, studentIDs []int64) error {
	if _, ok := b.GetAssignmentByTeacher(teacherID, assignmentID); !ok {
		return ErrAssignmentNotFound
	}

	for _, studentID := range studentIDs {
		if _, ok := b.GetStudentByTeacher(teacherID, studentID); !ok {
			return ErrStudentNotFound
		}
		_, err := b.db.Exec(
			b.rebind(`
				INSERT INTO assignment_students (assignment_id, student_id, assigned_at)
				VALUES (?, ?, ?)
				ON CONFLICT (assignment_id, student_id) DO NOTHING
			`),
			assignmentID,
			studentID,
			nowUTC(),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *sqlBackend) PublishAssignment(teacherID int64, assignmentID int64) (Assignment, error) {
	record, ok := b.GetAssignmentByTeacher(teacherID, assignmentID)
	if !ok {
		return Assignment{}, ErrAssignmentNotFound
	}
	if record.AnalysisStatus != "ready" {
		return Assignment{}, ErrAssignmentNotReady
	}

	_, err := b.db.Exec(
		b.rebind("UPDATE assignments SET status = ?, updated_at = ? WHERE teacher_id = ? AND id = ?"),
		"published",
		nowUTC(),
		teacherID,
		assignmentID,
	)
	if err != nil {
		return Assignment{}, err
	}

	record, _ = b.GetAssignmentByTeacher(teacherID, assignmentID)
	return record, nil
}

func (b *sqlBackend) ArchiveAssignment(teacherID int64, assignmentID int64) (Assignment, error) {
	if _, ok := b.GetAssignmentByTeacher(teacherID, assignmentID); !ok {
		return Assignment{}, ErrAssignmentNotFound
	}

	_, err := b.db.Exec(
		b.rebind("UPDATE assignments SET status = ?, updated_at = ? WHERE teacher_id = ? AND id = ?"),
		"archived",
		nowUTC(),
		teacherID,
		assignmentID,
	)
	if err != nil {
		return Assignment{}, err
	}

	record, _ := b.GetAssignmentByTeacher(teacherID, assignmentID)
	return record, nil
}

func (b *sqlBackend) GetAssignmentForStudent(studentID int64, assignmentID int64) (Assignment, bool) {
	return b.getAssignmentByFilters("a.id = ? AND EXISTS (SELECT 1 FROM assignment_students rel WHERE rel.assignment_id = a.id AND rel.student_id = ?)", assignmentID, studentID)
}

func (b *sqlBackend) ListAssignmentsByStudent(studentID int64) []Assignment {
	return b.listAssignmentsByFilters("a.status = ? AND EXISTS (SELECT 1 FROM assignment_students rel WHERE rel.assignment_id = a.id AND rel.student_id = ?)", "published", studentID)
}

func (b *sqlBackend) ListAssignedAssignmentsByStudent(studentID int64) []Assignment {
	return b.listAssignmentsByFilters("EXISTS (SELECT 1 FROM assignment_students rel WHERE rel.assignment_id = a.id AND rel.student_id = ?)", studentID)
}

func (b *sqlBackend) CreateProgress(input CreateProgressInput) (ProgressReport, error) {
	now := nowUTC()
	reportedAt := input.ReportedAt
	if reportedAt == "" {
		reportedAt = now
	}
	id, err := b.insertReturningID(
		"INSERT INTO progress_reports (assignment_id, student_id, current_target, step_summary, local_project_hash, reported_at, snapshot_json, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		input.AssignmentID,
		input.StudentID,
		input.CurrentTarget,
		input.StepSummary,
		input.LocalProjectHash,
		reportedAt,
		mustJSON(input.Snapshot),
		now,
	)
	if err != nil {
		return ProgressReport{}, err
	}

	record, ok := b.getLatestProgressBy("id = ?", id)
	if !ok {
		return ProgressReport{}, sql.ErrNoRows
	}
	return record, nil
}

func (b *sqlBackend) LatestProgress(studentID int64, assignmentID int64) (ProgressReport, bool) {
	return b.getLatestProgressBy("student_id = ? AND assignment_id = ?", studentID, assignmentID)
}

func (b *sqlBackend) CreateHint(input CreateHintInput) (HintRecord, error) {
	now := nowUTC()
	id, err := b.insertReturningID(
		"INSERT INTO hint_records (assignment_id, student_id, progress_report_id, prompt_input_json, hint_text, provider_name, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		input.AssignmentID,
		input.StudentID,
		input.ProgressReportID,
		mustJSON(input.PromptInput),
		input.HintText,
		input.ProviderName,
		now,
	)
	if err != nil {
		return HintRecord{}, err
	}

	record, ok := b.getLatestHintBy("id = ?", id)
	if !ok {
		return HintRecord{}, sql.ErrNoRows
	}
	return record, nil
}

func (b *sqlBackend) Ping(ctx context.Context) error {
	return b.db.PingContext(ctx)
}

func (b *sqlBackend) LatestHint(studentID int64, assignmentID int64) (HintRecord, bool) {
	return b.getLatestHintBy("student_id = ? AND assignment_id = ?", studentID, assignmentID)
}

func (b *sqlBackend) ListAssignedStudents(assignmentID int64) []Student {
	rows, err := b.db.Query(
		b.rebind(`
			SELECT s.id, s.teacher_id, s.username, s.display_name, s.password_hash, s.status, s.created_at
			FROM assignment_students rel
			JOIN students s ON s.id = rel.student_id
			WHERE rel.assignment_id = ?
			ORDER BY rel.assigned_at ASC, s.id ASC
		`),
		assignmentID,
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	records := make([]Student, 0)
	for rows.Next() {
		record, scanErr := scanStudent(rows)
		if scanErr == nil {
			records = append(records, record)
		}
	}
	return records
}

func (b *sqlBackend) getLatestProgressBy(whereClause string, args ...any) (ProgressReport, bool) {
	query := `
		SELECT id, assignment_id, student_id, current_target, step_summary, local_project_hash, reported_at, snapshot_json, created_at
		FROM progress_reports
		WHERE ` + whereClause + `
		ORDER BY id DESC
		LIMIT 1
	`

	row := b.db.QueryRow(b.rebind(query), args...)
	record, err := scanProgress(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ProgressReport{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) getLatestHintBy(whereClause string, args ...any) (HintRecord, bool) {
	query := `
		SELECT id, assignment_id, student_id, progress_report_id, prompt_input_json, hint_text, provider_name, created_at
		FROM hint_records
		WHERE ` + whereClause + `
		ORDER BY id DESC
		LIMIT 1
	`

	row := b.db.QueryRow(b.rebind(query), args...)
	record, err := scanHint(row)
	if errors.Is(err, sql.ErrNoRows) {
		return HintRecord{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) getAssignmentByFilters(whereClause string, args ...any) (Assignment, bool) {
	query := assignmentSelectSQL + " WHERE " + whereClause + " LIMIT 1"
	row := b.db.QueryRow(b.rebind(query), args...)
	record, err := scanAssignment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Assignment{}, false
	}
	return record, err == nil
}

func (b *sqlBackend) listAssignmentsByFilters(whereClause string, args ...any) []Assignment {
	query := assignmentSelectSQL + " WHERE " + whereClause + " ORDER BY a.id ASC"
	rows, err := b.db.Query(b.rebind(query), args...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	records := make([]Assignment, 0)
	for rows.Next() {
		record, scanErr := scanAssignment(rows)
		if scanErr == nil {
			records = append(records, record)
		}
	}
	return records
}

func (b *sqlBackend) insertReturningID(query string, args ...any) (int64, error) {
	if b.dialect == "postgres" {
		row := b.db.QueryRow(b.rebind(query+" RETURNING id"), args...)
		var id int64
		if err := row.Scan(&id); err != nil {
			return 0, err
		}
		return id, nil
	}

	result, err := b.db.Exec(b.rebind(query), args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (b *sqlBackend) rebind(query string) string {
	if b.dialect != "postgres" {
		return query
	}

	var builder strings.Builder
	index := 1
	for _, char := range query {
		if char == '?' {
			builder.WriteString(fmt.Sprintf("$%d", index))
			index++
			continue
		}
		builder.WriteRune(char)
	}
	return builder.String()
}

func scanTeacher(scanner interface{ Scan(...any) error }) (Teacher, error) {
	var (
		record    Teacher
		createdAt string
	)
	err := scanner.Scan(&record.ID, &record.Username, &record.PasswordHash, &record.Role, &record.Status, &createdAt)
	if err != nil {
		return Teacher{}, err
	}
	record.CreatedAt = parseTime(createdAt)
	return record, nil
}

func scanStudent(scanner interface{ Scan(...any) error }) (Student, error) {
	var (
		record    Student
		createdAt string
	)
	err := scanner.Scan(&record.ID, &record.TeacherID, &record.Username, &record.DisplayName, &record.PasswordHash, &record.Status, &createdAt)
	if err != nil {
		return Student{}, err
	}
	record.CreatedAt = parseTime(createdAt)
	return record, nil
}

func scanAssignment(scanner interface{ Scan(...any) error }) (Assignment, error) {
	var (
		record                Assignment
		createdAt             string
		updatedAt             string
		roleNamesJSON         sql.NullString
		scriptCountsJSON      sql.NullString
		blockCountsJSON       sql.NullString
		categoryCountsJSON    sql.NullString
		broadcastMessagesJSON sql.NullString
		variableNamesJSON     sql.NullString
		listNamesJSON         sql.NullString
		extensionsJSON        sql.NullString
		teachingPointsJSON    sql.NullString
	)
	err := scanner.Scan(
		&record.ID,
		&record.TeacherID,
		&record.Title,
		&record.Goal,
		&record.Description,
		&record.Status,
		&record.FileName,
		&record.SB3FilePath,
		&record.AnalysisStatus,
		&record.AnalysisErrorMessage,
		&createdAt,
		&updatedAt,
		&roleNamesJSON,
		&scriptCountsJSON,
		&blockCountsJSON,
		&categoryCountsJSON,
		&broadcastMessagesJSON,
		&variableNamesJSON,
		&listNamesJSON,
		&extensionsJSON,
		&teachingPointsJSON,
	)
	if err != nil {
		return Assignment{}, err
	}

	record.CreatedAt = parseTime(createdAt)
	record.UpdatedAt = parseTime(updatedAt)
	record.Analysis = AssignmentAnalysis{
		RoleNames:         decodeStrings(roleNamesJSON.String),
		ScriptCounts:      decodeIntMap(scriptCountsJSON.String),
		BlockCounts:       decodeIntMap(blockCountsJSON.String),
		CategoryCounts:    decodeIntMap(categoryCountsJSON.String),
		BroadcastMessages: decodeStrings(broadcastMessagesJSON.String),
		VariableNames:     decodeStrings(variableNamesJSON.String),
		ListNames:         decodeStrings(listNamesJSON.String),
		Extensions:        decodeStrings(extensionsJSON.String),
		TeachingPoints:    decodeStrings(teachingPointsJSON.String),
	}
	return record, nil
}

func scanProgress(scanner interface{ Scan(...any) error }) (ProgressReport, error) {
	var (
		record       ProgressReport
		snapshotJSON string
		createdAt    string
	)
	err := scanner.Scan(
		&record.ID,
		&record.AssignmentID,
		&record.StudentID,
		&record.CurrentTarget,
		&record.StepSummary,
		&record.LocalProjectHash,
		&record.ReportedAt,
		&snapshotJSON,
		&createdAt,
	)
	if err != nil {
		return ProgressReport{}, err
	}

	record.Snapshot = decodeObject(snapshotJSON)
	record.CreatedAt = parseTime(createdAt)
	return record, nil
}

func scanHint(scanner interface{ Scan(...any) error }) (HintRecord, error) {
	var (
		record          HintRecord
		promptInputJSON string
		createdAt       string
	)
	err := scanner.Scan(
		&record.ID,
		&record.AssignmentID,
		&record.StudentID,
		&record.ProgressReportID,
		&promptInputJSON,
		&record.HintText,
		&record.ProviderName,
		&createdAt,
	)
	if err != nil {
		return HintRecord{}, err
	}

	record.PromptInput = decodeObject(promptInputJSON)
	record.CreatedAt = parseTime(createdAt)
	return record, nil
}

func decodeStrings(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var items []string
	_ = json.Unmarshal([]byte(raw), &items)
	if items == nil {
		return []string{}
	}
	return items
}

func decodeIntMap(raw string) map[string]int {
	if strings.TrimSpace(raw) == "" {
		return map[string]int{}
	}
	var items map[string]int
	_ = json.Unmarshal([]byte(raw), &items)
	if items == nil {
		return map[string]int{}
	}
	return items
}

func decodeObject(raw string) map[string]any {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}
	}
	var payload map[string]any
	_ = json.Unmarshal([]byte(raw), &payload)
	if payload == nil {
		return map[string]any{}
	}
	return payload
}

func mustJSON(value any) string {
	payload, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(payload)
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func parseTime(raw string) time.Time {
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

const assignmentSelectSQL = `
	SELECT
		a.id,
		a.teacher_id,
		a.title,
		a.goal,
		a.description,
		a.status,
		a.file_name,
		a.sb3_file_path,
		a.analysis_status,
		a.analysis_error_message,
		a.created_at,
		a.updated_at,
		aa.role_names_json,
		aa.script_counts_json,
		aa.block_counts_json,
		aa.category_counts_json,
		aa.broadcast_messages_json,
		aa.variable_names_json,
		aa.list_names_json,
		aa.extensions_json,
		aa.teaching_points_json
	FROM assignments a
	LEFT JOIN assignment_analysis aa ON aa.assignment_id = a.id
`

var sqliteSchemaStatements = []string{
	"PRAGMA foreign_keys = ON",
	`
	CREATE TABLE IF NOT EXISTS teachers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'teacher',
		status TEXT NOT NULL DEFAULT 'active',
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS teacher_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE,
		expires_at TEXT,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
		username TEXT NOT NULL UNIQUE,
		display_name TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS student_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE,
		client_type TEXT NOT NULL,
		expires_at TEXT,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS assignments (
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
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS assignment_students (
		assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
		student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		assigned_at TEXT NOT NULL,
		PRIMARY KEY (assignment_id, student_id)
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS assignment_analysis (
		assignment_id INTEGER PRIMARY KEY REFERENCES assignments(id) ON DELETE CASCADE,
		role_names_json TEXT NOT NULL,
		script_counts_json TEXT NOT NULL DEFAULT '{}',
		block_counts_json TEXT NOT NULL,
		category_counts_json TEXT NOT NULL,
		broadcast_messages_json TEXT NOT NULL DEFAULT '[]',
		variable_names_json TEXT NOT NULL DEFAULT '[]',
		list_names_json TEXT NOT NULL DEFAULT '[]',
		extensions_json TEXT NOT NULL DEFAULT '[]',
		teaching_points_json TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS progress_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
		student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		current_target TEXT NOT NULL,
		step_summary TEXT NOT NULL,
		local_project_hash TEXT NOT NULL,
		reported_at TEXT NOT NULL,
		snapshot_json TEXT NOT NULL,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS hint_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
		student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		progress_report_id INTEGER NOT NULL REFERENCES progress_reports(id) ON DELETE CASCADE,
		prompt_input_json TEXT NOT NULL,
		hint_text TEXT NOT NULL,
		provider_name TEXT NOT NULL,
		created_at TEXT NOT NULL
	)
	`,
	"CREATE INDEX IF NOT EXISTS idx_students_teacher_id ON students(teacher_id)",
	"CREATE INDEX IF NOT EXISTS idx_teacher_sessions_token ON teacher_sessions(token)",
	"CREATE INDEX IF NOT EXISTS idx_student_sessions_token ON student_sessions(token)",
	"CREATE INDEX IF NOT EXISTS idx_assignments_teacher_id ON assignments(teacher_id)",
	"CREATE INDEX IF NOT EXISTS idx_assignment_students_student_id ON assignment_students(student_id)",
	"CREATE INDEX IF NOT EXISTS idx_progress_student_assignment_id ON progress_reports(student_id, assignment_id, id DESC)",
	"CREATE INDEX IF NOT EXISTS idx_hint_student_assignment_id ON hint_records(student_id, assignment_id, id DESC)",
}

var postgresSchemaStatements = []string{
	`
	CREATE TABLE IF NOT EXISTS teachers (
		id BIGSERIAL PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL DEFAULT 'teacher',
		status TEXT NOT NULL DEFAULT 'active',
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS teacher_sessions (
		id BIGSERIAL PRIMARY KEY,
		teacher_id BIGINT NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE,
		expires_at TEXT,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS students (
		id BIGSERIAL PRIMARY KEY,
		teacher_id BIGINT NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
		username TEXT NOT NULL UNIQUE,
		display_name TEXT NOT NULL,
		password_hash TEXT NOT NULL,
		status TEXT NOT NULL,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS student_sessions (
		id BIGSERIAL PRIMARY KEY,
		student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE,
		client_type TEXT NOT NULL,
		expires_at TEXT,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS assignments (
		id BIGSERIAL PRIMARY KEY,
		teacher_id BIGINT NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
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
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS assignment_students (
		assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
		student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		assigned_at TEXT NOT NULL,
		PRIMARY KEY (assignment_id, student_id)
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS assignment_analysis (
		assignment_id BIGINT PRIMARY KEY REFERENCES assignments(id) ON DELETE CASCADE,
		role_names_json TEXT NOT NULL,
		script_counts_json TEXT NOT NULL DEFAULT '{}',
		block_counts_json TEXT NOT NULL,
		category_counts_json TEXT NOT NULL,
		broadcast_messages_json TEXT NOT NULL DEFAULT '[]',
		variable_names_json TEXT NOT NULL DEFAULT '[]',
		list_names_json TEXT NOT NULL DEFAULT '[]',
		extensions_json TEXT NOT NULL DEFAULT '[]',
		teaching_points_json TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS progress_reports (
		id BIGSERIAL PRIMARY KEY,
		assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
		student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		current_target TEXT NOT NULL,
		step_summary TEXT NOT NULL,
		local_project_hash TEXT NOT NULL,
		reported_at TEXT NOT NULL,
		snapshot_json TEXT NOT NULL,
		created_at TEXT NOT NULL
	)
	`,
	`
	CREATE TABLE IF NOT EXISTS hint_records (
		id BIGSERIAL PRIMARY KEY,
		assignment_id BIGINT NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
		student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
		progress_report_id BIGINT NOT NULL REFERENCES progress_reports(id) ON DELETE CASCADE,
		prompt_input_json TEXT NOT NULL,
		hint_text TEXT NOT NULL,
		provider_name TEXT NOT NULL,
		created_at TEXT NOT NULL
	)
	`,
	"CREATE INDEX IF NOT EXISTS idx_students_teacher_id ON students(teacher_id)",
	"CREATE INDEX IF NOT EXISTS idx_teacher_sessions_token ON teacher_sessions(token)",
	"CREATE INDEX IF NOT EXISTS idx_student_sessions_token ON student_sessions(token)",
	"CREATE INDEX IF NOT EXISTS idx_assignments_teacher_id ON assignments(teacher_id)",
	"CREATE INDEX IF NOT EXISTS idx_assignment_students_student_id ON assignment_students(student_id)",
	"CREATE INDEX IF NOT EXISTS idx_progress_student_assignment_id ON progress_reports(student_id, assignment_id, id DESC)",
	"CREATE INDEX IF NOT EXISTS idx_hint_student_assignment_id ON hint_records(student_id, assignment_id, id DESC)",
}

func (b *sqlBackend) ensureAssignmentAnalysisColumns() error {
	requiredColumns := map[string]string{
		"script_counts_json":      "script_counts_json TEXT NOT NULL DEFAULT '{}'",
		"broadcast_messages_json": "broadcast_messages_json TEXT NOT NULL DEFAULT '[]'",
		"variable_names_json":     "variable_names_json TEXT NOT NULL DEFAULT '[]'",
		"list_names_json":         "list_names_json TEXT NOT NULL DEFAULT '[]'",
		"extensions_json":         "extensions_json TEXT NOT NULL DEFAULT '[]'",
	}

	for columnName, columnDefinition := range requiredColumns {
		exists, err := b.columnExists("assignment_analysis", columnName)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if _, err := b.db.Exec(fmt.Sprintf("ALTER TABLE assignment_analysis ADD COLUMN %s", columnDefinition)); err != nil {
			return err
		}
	}

	return nil
}

func (b *sqlBackend) ensureTeacherColumns() error {
	requiredColumns := map[string]string{
		"role":   "role TEXT NOT NULL DEFAULT 'teacher'",
		"status": "status TEXT NOT NULL DEFAULT 'active'",
	}

	for columnName, columnDefinition := range requiredColumns {
		exists, err := b.columnExists("teachers", columnName)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		if _, err := b.db.Exec(fmt.Sprintf("ALTER TABLE teachers ADD COLUMN %s", columnDefinition)); err != nil {
			return err
		}
	}

	return nil
}

func (b *sqlBackend) columnExists(tableName string, columnName string) (bool, error) {
	if b.dialect == "postgres" {
		row := b.db.QueryRow(
			`
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = current_schema()
					AND table_name = $1
					AND column_name = $2
			`,
			tableName,
			columnName,
		)
		var marker int
		err := row.Scan(&marker)
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return err == nil, err
	}

	rows, err := b.db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid        int
			name       string
			columnType string
			notNull    int
			defaultVal sql.NullString
			primaryKey int
		)
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultVal, &primaryKey); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	return false, rows.Err()
}
