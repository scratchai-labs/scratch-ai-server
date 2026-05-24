from __future__ import annotations

import os

from fastapi.middleware.cors import CORSMiddleware
from fastapi import FastAPI

from app.core.config import Settings, load_settings
from app.core.db import Database
from app.routes.ai import router as ai_router
from app.routes.auth import router as auth_router
from app.routes.health import router as health_router
from app.routes.progress import router as progress_router
from app.routes.releases import router as releases_router
from app.services.ai import create_prompt_provider


def create_app(settings: Settings | None = None) -> FastAPI:
    resolved_settings = settings or load_settings()
    db = Database(
        database_path=resolved_settings.database_path,
        database_url=resolved_settings.database_url,
    )
    db.init_schema()

    app = FastAPI(title="Scratch AI Server API")
    app.state.settings = resolved_settings
    app.state.db = db
    app.state.prompt_provider = create_prompt_provider(resolved_settings)

    if resolved_settings.cors_allowed_origins:
        app.add_middleware(
            CORSMiddleware,
            allow_origins=list(resolved_settings.cors_allowed_origins),
            allow_methods=["*"],
            allow_headers=["*"],
        )

    app.include_router(health_router)
    app.include_router(auth_router)
    app.include_router(releases_router)
    app.include_router(progress_router)
    app.include_router(ai_router)
    return app


app = create_app()


def main() -> None:
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=False)


if __name__ == "__main__":
    main()
