from __future__ import annotations

import sqlite3
from contextlib import contextmanager
from pathlib import Path
from typing import Any, Iterator

try:
    import psycopg
    from psycopg.rows import dict_row
except ImportError:  # pragma: no cover - exercised indirectly in deploy env
    psycopg = None
    dict_row = None

from .schema import schema_statements_for


class Database:
    def __init__(self, database_path: Path | None = None, database_url: str | None = None) -> None:
        if database_path is None and database_url is None:
            raise ValueError("database_path or database_url is required")

        self.path = Path(database_path) if database_path is not None else None
        self.url = database_url
        self.dialect = "postgres" if database_url else "sqlite"

        if self.path is not None:
            self.path.parent.mkdir(parents=True, exist_ok=True)

    @contextmanager
    def connect(self) -> Iterator[sqlite3.Connection | Any]:
        if self.dialect == "postgres":
            if psycopg is None:
                raise RuntimeError("psycopg is required when DATABASE_URL is configured")
            connection = psycopg.connect(self.url, row_factory=dict_row)
        else:
            connection = sqlite3.connect(self.path)
            connection.row_factory = sqlite3.Row
            connection.execute("PRAGMA foreign_keys = ON")
        try:
            yield connection
            connection.commit()
        finally:
            connection.close()

    def init_schema(self) -> None:
        with self.connect() as connection:
            for statement in schema_statements_for(self.dialect):
                connection.execute(statement)

    def fetch_one(self, sql: str, params: tuple[Any, ...] = ()) -> sqlite3.Row | None:
        with self.connect() as connection:
            return connection.execute(self._sql(sql), params).fetchone()

    def fetch_all(self, sql: str, params: tuple[Any, ...] = ()) -> list[sqlite3.Row]:
        with self.connect() as connection:
            return connection.execute(self._sql(sql), params).fetchall()

    def insert(self, sql: str, params: tuple[Any, ...] = ()) -> int:
        with self.connect() as connection:
            cursor = connection.execute(self._insert_sql(sql), params)
            if self.dialect == "postgres":
                row = cursor.fetchone()
                if row is None or "id" not in row:
                    raise RuntimeError("postgres insert did not return an id")
                return int(row["id"])
            return int(cursor.lastrowid)

    def execute(self, sql: str, params: tuple[Any, ...] = ()) -> int:
        with self.connect() as connection:
            cursor = connection.execute(self._sql(sql), params)
            return int(getattr(cursor, "rowcount", 0) or 0)

    def _insert_sql(self, sql: str) -> str:
        normalized_sql = self._sql(sql)
        if self.dialect == "postgres" and "RETURNING" not in normalized_sql.upper():
            return f"{normalized_sql} RETURNING id"
        return normalized_sql

    def _sql(self, sql: str) -> str:
        if self.dialect != "postgres":
            return sql
        return sql.replace("?", "%s")
