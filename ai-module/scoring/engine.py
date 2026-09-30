from dataclasses import dataclass
from typing import Literal
from scoring.calculateTerraScore import calculate_terra_score, StageType, RiskLevel
from scoring.rag import RagStore

@dataclass
class RiskInput:
    ltv: float
    debt_ratio: float
    location_score: float
    sponsor_track_record: float
    project_stage: StageType

@dataclass
class EngineOutput:
    risk_score: float
    risk_level: RiskLevel
    confidence: float
    recommendations: list[str]

class TerraScoringEngine:
    def __init__(self, rag_store: RagStore) -> None:
        self.rag_store = rag_store

    async def score(self, risk_input: RiskInput, trace_id: str | None = None) -> EngineOutput:
        # 1. Obtener contexto de riesgo sin bloquear el event loop
        rag_context = await self.rag_store.get_risk_context(risk_input.project_stage, trace_id=trace_id)

        # 2. Cálculo puro y determinista
        score_result = calculate_terra_score(
            ltv=risk_input.ltv,
            debt_ratio=risk_input.debt_ratio,
            location_score=risk_input.location_score,
            sponsor_track_record=risk_input.sponsor_track_record,
            project_stage=risk_input.project_stage,
            stage_adjustment=rag_context["stage_adjustment"],
            confidence_base=rag_context["confidence_base"],
        )

        recommendations = list(rag_context["default_recommendations"])

        # 3. Reglas de negocio sobre umbrales
        if risk_input.debt_ratio > 65:
            recommendations.append("Lower debt ratio before opening the round.")
        if risk_input.ltv > 75:
            recommendations.append("Increase equity buffer to reduce LTV risk.")
        if risk_input.location_score < 50:
            recommendations.append("Run deeper location due diligence.")
        if risk_input.sponsor_track_record < 55:
            recommendations.append("Request stronger sponsor guarantees.")

        recommendations.extend(rag_context["backend_recommendations"])
        unique_recommendations = list(dict.fromkeys(recommendations))

        # 4. Persistencia asíncrona en base de datos
        await self.rag_store.save_evaluation(
            project_stage=risk_input.project_stage,
            ltv=risk_input.ltv,
            debt_ratio=risk_input.debt_ratio,
            location_score=risk_input.location_score,
            sponsor_track_record=risk_input.sponsor_track_record,
            risk_score=float(score_result["risk_score"]),
            risk_level=str(score_result["risk_level"]),
            confidence=float(score_result["confidence"]),
        )

        return EngineOutput(
            risk_score=float(score_result["risk_score"]),
            risk_level=score_result["risk_level"], # type: ignore
            confidence=float(score_result["confidence"]),
            recommendations=unique_recommendations,
        )