from fastapi.testclient import TestClient

from app.core.config import Settings, load_settings
from app.core.db import Database
from app.main import create_app


def test_settings_prefer_database_url_and_parse_cors_origins(monkeypatch):
    monkeypatch.setenv("DATABASE_URL", "postgresql://scratch:secret@db.example.com:5432/scratchai")
    monkeypatch.delenv("SERVER_API_DB_PATH", raising=False)
    monkeypatch.setenv(
        "CORS_ALLOWED_ORIGINS",
        "https://scratch-ai.vercel.app, https://preview-scratch-ai.vercel.app ",
    )

    settings = load_settings()

    assert settings.database_url == "postgresql://scratch:secret@db.example.com:5432/scratchai"
    assert settings.database_path is None
    assert settings.cors_allowed_origins == (
        "https://scratch-ai.vercel.app",
        "https://preview-scratch-ai.vercel.app",
    )


def test_create_app_enables_cors_preflight_for_configured_origin(tmp_path):
    settings = Settings(
        database_path=tmp_path / "server-api.sqlite3",
        database_url=None,
        ai_provider="fallback",
        ai_base_url=None,
        ai_api_key=None,
        ai_model="scratch-ai-coach",
        cors_allowed_origins=("https://scratch-ai.vercel.app",),
    )
    client = TestClient(create_app(settings))

    response = client.options(
        "/api/teacher/login",
        headers={
            "Origin": "https://scratch-ai.vercel.app",
            "Access-Control-Request-Method": "POST",
        },
    )

    assert response.status_code == 200
    assert response.headers["access-control-allow-origin"] == "https://scratch-ai.vercel.app"
    assert "POST" in response.headers["access-control-allow-methods"]


def test_postgres_database_uses_postgres_schema_and_sql_placeholders(monkeypatch):
    import app.core.db as db_module

    class FakeCursor:
        def __init__(self, row=None):
            self._row = row

        def fetchone(self):
            return self._row

        def fetchall(self):
            return []

    class FakeConnection:
        def __init__(self):
            self.executed: list[tuple[str, tuple[object, ...]]] = []
            self.committed = False
            self.closed = False

        def execute(self, sql, params=()):
            normalized_params = tuple(params)
            self.executed.append((sql, normalized_params))
            if "RETURNING id" in sql:
                return FakeCursor({"id": 41})
            if sql.startswith("SELECT id, username FROM teachers"):
                return FakeCursor({"id": 41, "username": "teacher1"})
            return FakeCursor()

        def commit(self):
            self.committed = True

        def close(self):
            self.closed = True

    class FakePsycopg:
        def __init__(self):
            self.connections: list[FakeConnection] = []
            self.last_connect: tuple[str, object] | None = None

        def connect(self, url, row_factory=None):
            self.last_connect = (url, row_factory)
            connection = FakeConnection()
            self.connections.append(connection)
            return connection

    fake_psycopg = FakePsycopg()
    fake_dict_row = object()
    monkeypatch.setattr(db_module, "psycopg", fake_psycopg)
    monkeypatch.setattr(db_module, "dict_row", fake_dict_row)

    db = Database(database_url="postgresql://scratch:secret@db.example.com:5432/scratchai")
    db.init_schema()
    teacher_id = db.insert(
        "INSERT INTO teachers (username, password_hash, created_at) VALUES (?, ?, ?)",
        ("teacher1", "hash", "2026-05-08T00:00:00+00:00"),
    )
    teacher = db.fetch_one("SELECT id, username FROM teachers WHERE id = ?", (teacher_id,))

    assert fake_psycopg.last_connect == (
        "postgresql://scratch:secret@db.example.com:5432/scratchai",
        fake_dict_row,
    )
    assert any(
        "CREATE TABLE IF NOT EXISTS teachers" in sql and "BIGSERIAL PRIMARY KEY" in sql
        for sql, _params in fake_psycopg.connections[0].executed
    )
    assert all("PRAGMA" not in sql for sql, _params in fake_psycopg.connections[0].executed)
    assert fake_psycopg.connections[1].executed[0][0].endswith("RETURNING id")
    assert fake_psycopg.connections[1].executed[0][0].count("%s") == 3
    assert fake_psycopg.connections[2].executed[0] == (
        "SELECT id, username FROM teachers WHERE id = %s",
        (41,),
    )
    assert teacher_id == 41
    assert teacher == {"id": 41, "username": "teacher1"}
