from __future__ import annotations

import os
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True, slots=True)
class Settings:
    database_path: Path | None
    database_url: str | None
    ai_provider: str
    ai_base_url: str | None
    ai_api_key: str | None
    ai_model: str
    cors_allowed_origins: tuple[str, ...]


def load_settings() -> Settings:
    database_url = _get_optional_env("DATABASE_URL")
    database_path = None
    if database_url is None:
        database_path = Path(os.getenv("SERVER_API_DB_PATH", "server-api.sqlite3")).expanduser()

    ai_base_url = _get_optional_env("AI_BASE_URL")
    ai_api_key = _get_optional_env("AI_API_KEY")
    ai_model = os.getenv("AI_MODEL", "scratch-ai-coach")
    ai_provider = os.getenv("AI_PROVIDER", "fallback").strip().lower() or "fallback"
    cors_allowed_origins = _split_csv_env("CORS_ALLOWED_ORIGINS")

    if ai_base_url and ai_provider == "fallback":
        ai_provider = "http"

    return Settings(
        database_path=database_path,
        database_url=database_url,
        ai_provider=ai_provider,
        ai_base_url=ai_base_url,
        ai_api_key=ai_api_key,
        ai_model=ai_model,
        cors_allowed_origins=cors_allowed_origins,
    )


def _get_optional_env(name: str) -> str | None:
    value = os.getenv(name)
    if value is None:
        return None
    normalized = value.strip()
    return normalized or None


def _split_csv_env(name: str) -> tuple[str, ...]:
    raw_value = os.getenv(name, "")
    return tuple(part.strip() for part in raw_value.split(",") if part.strip())
