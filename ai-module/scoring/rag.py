import asyncio
import logging
import os
import time
from datetime import datetime, timezone
from typing import Any
import aiosqlite
import httpx
from scoring.calculateTerraScore import STAGE_RISK_FACTORS, STAGE_CONFIDENCE_FACTORS

logger = logging.getLogger("scoring.rag")

class RagStore:
    def __init__(self, db_path: str = "./ai_module.db", backend_api_url: str = "http://localhost:8080"):
        self.db_path = db_path
        self.backend_api_url = backend_api_url.rstrip("/")
        self._http_client: httpx.AsyncClient | None = None
        self._cached_avg_return: float | None = None
        self._last_cache_time: float = 0.0
        self._cache_lock = asyncio.Lock()
        self._ensure_schema_sync()

    def _ensure_schema_sync(self) -> None:
        import sqlite3
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS ai_risk_evaluations (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    project_stage TEXT NOT NULL,
                    ltv REAL NOT NULL,
                    debt_ratio REAL NOT NULL,
                    location_score REAL NOT NULL,
                    sponsor_track_record REAL NOT NULL,
                    risk_score REAL NOT NULL,
                    risk_level TEXT NOT NULL,
                    confidence REAL NOT NULL,
                    created_at TEXT NOT NULL
                );
            """)

    @classmethod
    def from_env(cls) -> "RagStore":
        return cls(
            db_path=os.getenv("AI_DATABASE_PATH", "./ai_module.db"),
            backend_api_url=os.getenv("BACKEND_API_URL", "http://localhost:8080")
        )

    async def initialize(self) -> None:
        """Initializes database schema and HTTP client with connection pooling."""
        self._http_client = httpx.AsyncClient(
            timeout=httpx.Timeout(connect=1.0, read=2.0, write=1.0, pool=5.0),
            limits=httpx.Limits(max_keepalive_connections=20, max_connections=50),
        )
        async with aiosqlite.connect(self.db_path) as db:
            await db.execute("""
                CREATE TABLE IF NOT EXISTS ai_risk_evaluations (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    project_stage TEXT NOT NULL,
                    ltv REAL NOT NULL,
                    debt_ratio REAL NOT NULL,
                    location_score REAL NOT NULL,
                    sponsor_track_record REAL NOT NULL,
                    risk_score REAL NOT NULL,
                    risk_level TEXT NOT NULL,
                    confidence REAL NOT NULL,
                    created_at TEXT NOT NULL
                );
            """)
            await db.commit()

    async def close(self) -> None:
        if self._http_client:
            await self._http_client.aclose()

    async def save_evaluation(self, **kwargs: Any) -> None:
        created_at = datetime.now(timezone.utc).isoformat()
        try:
            async with aiosqlite.connect(self.db_path) as db:
                await db.execute("""
                    INSERT INTO ai_risk_evaluations 
                    (project_stage, ltv, debt_ratio, location_score, sponsor_track_record, risk_score, risk_level, confidence, created_at)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
                """, (
                    kwargs['project_stage'],
                    kwargs['ltv'],
                    kwargs['debt_ratio'],
                    kwargs['location_score'],
                    kwargs['sponsor_track_record'],
                    kwargs['risk_score'],
                    kwargs['risk_level'],
                    kwargs['confidence'],
                    created_at,
                ))
                await db.commit()
        except Exception as ex:
            logger.error(f"Failed to persist risk evaluation to SQLite: {ex}", exc_info=True)

    async def get_risk_context(self, project_stage: str, trace_id: str | None = None) -> dict[str, Any]:
        backend_recommendations = await self._recommendations_from_backend(project_stage, trace_id)
        
        return {
            "stage_adjustment": STAGE_RISK_FACTORS,
            "confidence_base": STAGE_CONFIDENCE_FACTORS,
            "default_recommendations": ["Current profile looks balanced for initial screening."],
            "backend_recommendations": backend_recommendations,
        }

    async def _recommendations_from_backend(self, project_stage: str, trace_id: str | None) -> list[str]:
        """Fetches project average returns with TTL caching to decouple microservice calls."""
        now = time.time()
        # TTL Cache de 5 minutos
        if self._cached_avg_return is None or (now - self._last_cache_time) > 300:
            async with self._cache_lock:
                if self._cached_avg_return is None or (now - self._last_cache_time) > 300:
                    try:
                        headers = {"X-Request-ID": trace_id} if trace_id else {}
                        client = self._http_client or httpx.AsyncClient(timeout=2.0)
                        resp = await client.get(f"{self.backend_api_url}/api/projects", headers=headers)
                        if resp.status_code == 200:
                            payload = resp.json()
                            items = payload.get("items", [])
                            returns = [float(i["expectedReturn"]) for i in items if "expectedReturn" in i]
                            if returns:
                                self._cached_avg_return = sum(returns) / len(returns)
                                self._last_cache_time = now
                    except Exception as ex:
                        logger.warning(f"Could not refresh project stats from backend: {ex}. Using existing cache or defaults.")

        recs: list[str] = []
        if self._cached_avg_return is not None:
            if self._cached_avg_return < 11.0:
                recs.append("Backend project dataset shows lower average returns; consider conservative stress tests.")
            elif self._cached_avg_return > 13.0:
                recs.append("Backend projects show high expected returns; validate assumptions with downside scenarios.")

        if project_stage == "land":
            recs.append("For land-stage projects, request stricter milestone-based disbursement from the fiduciary model.")

        return recs