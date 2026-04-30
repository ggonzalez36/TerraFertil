import json
import os
import sqlite3
from datetime import datetime, timezone
from typing import Literal, Union
from urllib import error, request

class RagStore:
    def __init__(self, db_path: str = "./ai_module.db", backend_api_url: str = "http://localhost:8080"):
        self.db_path = db_path
        self.backend_api_url = backend_api_url.rstrip("/")
        self._ensure_schema()

    @classmethod
    def from_env(cls) -> "RagStore":
        return cls(
            db_path=os.getenv("AI_DATABASE_PATH", "./ai_module.db"),
            backend_api_url=os.getenv("BACKEND_API_URL", "http://localhost:8080")
        )

    def _connect(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self.db_path)
        conn.row_factory = sqlite3.Row
        return conn

    def _ensure_schema(self):
        with self._connect() as conn:
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

    def save_evaluation(self, **kwargs):
        created_at = datetime.now(timezone.utc).isoformat()
        with self._connect() as conn:
            conn.execute("""
                INSERT INTO ai_risk_evaluations 
                (project_stage, ltv, debt_ratio, location_score, sponsor_track_record, risk_score, risk_level, confidence, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (kwargs['project_stage'], kwargs['ltv'], kwargs['debt_ratio'], kwargs['location_score'], 
                  kwargs['sponsor_track_record'], kwargs['risk_score'], kwargs['risk_level'], kwargs['confidence'], created_at))

    def get_risk_context(self, project_stage: str) -> dict:
        # Configuración de negocio para el motor
        stage_adjustment = {"land": 15.0, "pre-sale": 8.0, "construction": 4.0, "stabilized": -5.0}
        confidence_base = {"land": 0.68, "pre-sale": 0.72, "construction": 0.74, "stabilized": 0.82}
        
        return {
            "stage_adjustment": stage_adjustment,
            "confidence_base": confidence_base,
            "default_recommendations": ["Current profile looks balanced for initial screening."],
            "backend_recommendations": self._recommendations_from_backend(project_stage)
        }

    def _recommendations_from_backend(self, project_stage: str) -> list[str]:
        endpoint = f"{self.backend_api_url}/api/projects"
        # Lógica de fetch al backend de Go para promediar retornos y generar alertas
        # (Basado en los cálculos de valid/avg_return que vi en tus fotos)
        try:
            req = request.Request(endpoint, method="GET")
            with request.urlopen(req, timeout=2.5) as resp:
                if resp.status >= 400: return []
                payload = json.loads(resp.read().decode("utf-8"))
            
            items = payload.get("items", [])
            if not items: return []
            
            returns = [float(i["expectedReturn"]) for i in items if "expectedReturn" in i]
            avg_return = sum(returns) / len(returns) if returns else 0
            
            recs = []
            if avg_return < 11.0:
                recs.append("Backend project dataset shows lower average returns; consider conservative stress tests.")
            elif avg_return > 13.0:
                recs.append("Backend projects show high expected returns; validate assumptions with downside scenarios.")
            
            if project_stage == "land":
                recs.append("For land-stage projects, request stricter milestone-based disbursement from the fiduciary model.")
            return recs
        except:
            return []