SQLITE_SCHEMA_STATEMENTS = (
    "PRAGMA foreign_keys = ON",
    """
    CREATE TABLE IF NOT EXISTS teachers (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS students (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        display_name TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS auth_tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        role TEXT NOT NULL,
        user_id INTEGER NOT NULL,
        token TEXT NOT NULL UNIQUE,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS releases (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        teacher_id INTEGER NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
        title TEXT NOT NULL,
        sb3_url TEXT NOT NULL,
        goal TEXT NOT NULL,
        status TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS release_assignments (
        release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
        student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        PRIMARY KEY (release_id, student_id)
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS progress_updates (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
        current_target TEXT NOT NULL,
        step_summary TEXT NOT NULL,
        snapshot_json TEXT NOT NULL,
        updated_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS ai_prompts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        student_id INTEGER NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
        current_target TEXT NOT NULL,
        step_summary TEXT NOT NULL,
        snapshot_json TEXT NOT NULL,
        prompt TEXT NOT NULL,
        provider_name TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    "CREATE INDEX IF NOT EXISTS idx_students_teacher_id ON students(teacher_id)",
    "CREATE INDEX IF NOT EXISTS idx_releases_teacher_id ON releases(teacher_id)",
    "CREATE INDEX IF NOT EXISTS idx_progress_updates_student_release_updated_at ON progress_updates(student_id, release_id, updated_at DESC)",
    "CREATE INDEX IF NOT EXISTS idx_ai_prompts_student_release_created_at ON ai_prompts(student_id, release_id, created_at DESC)",
)

POSTGRES_SCHEMA_STATEMENTS = (
    """
    CREATE TABLE IF NOT EXISTS teachers (
        id BIGSERIAL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS students (
        id BIGSERIAL PRIMARY KEY,
        teacher_id BIGINT NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
        username TEXT NOT NULL UNIQUE,
        password_hash TEXT NOT NULL,
        display_name TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS auth_tokens (
        id BIGSERIAL PRIMARY KEY,
        role TEXT NOT NULL,
        user_id BIGINT NOT NULL,
        token TEXT NOT NULL UNIQUE,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS releases (
        id BIGSERIAL PRIMARY KEY,
        teacher_id BIGINT NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
        title TEXT NOT NULL,
        sb3_url TEXT NOT NULL,
        goal TEXT NOT NULL,
        status TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS release_assignments (
        release_id BIGINT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
        student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        PRIMARY KEY (release_id, student_id)
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS progress_updates (
        id BIGSERIAL PRIMARY KEY,
        student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        release_id BIGINT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
        current_target TEXT NOT NULL,
        step_summary TEXT NOT NULL,
        snapshot_json TEXT NOT NULL,
        updated_at TEXT NOT NULL
    )
    """,
    """
    CREATE TABLE IF NOT EXISTS ai_prompts (
        id BIGSERIAL PRIMARY KEY,
        student_id BIGINT NOT NULL REFERENCES students(id) ON DELETE CASCADE,
        release_id BIGINT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
        current_target TEXT NOT NULL,
        step_summary TEXT NOT NULL,
        snapshot_json TEXT NOT NULL,
        prompt TEXT NOT NULL,
        provider_name TEXT NOT NULL,
        created_at TEXT NOT NULL
    )
    """,
    "CREATE INDEX IF NOT EXISTS idx_students_teacher_id ON students(teacher_id)",
    "CREATE INDEX IF NOT EXISTS idx_releases_teacher_id ON releases(teacher_id)",
    "CREATE INDEX IF NOT EXISTS idx_progress_updates_student_release_updated_at ON progress_updates(student_id, release_id, updated_at DESC)",
    "CREATE INDEX IF NOT EXISTS idx_ai_prompts_student_release_created_at ON ai_prompts(student_id, release_id, created_at DESC)",
)


def schema_statements_for(dialect: str) -> tuple[str, ...]:
    if dialect == "postgres":
        return POSTGRES_SCHEMA_STATEMENTS
    return SQLITE_SCHEMA_STATEMENTS
