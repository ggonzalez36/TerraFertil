from dataclasses import dataclass
from typing import Literal
from scoring.calculateTerraScore import calculate_terra_score
from scoring.rag import RagStore

@dataclass
class RiskInput:
    ltv: float
    debt_ratio: float
    location_score: float
    sponsor_track_record: float
    project_stage: Literal["land", "pre-sale", "construction", "stabilized"]

@dataclass
class EngineOutput:
    risk_score: float
    risk_level: Literal["low", "medium", "high"]
    confidence: float
    recommendations: list[str]

class TerraScoringEngine:
    def __init__(self, rag_store: RagStore) -> None:
        self.rag_store = rag_store

    def score(self, risk_input: RiskInput) -> EngineOutput:
        # Obtenemos el contexto (ajustes y confianza) desde el RAG/DB
        rag_context = self.rag_store.get_risk_context(risk_input.project_stage)
        
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

        # Lógica de recomendaciones dinámicas basada en umbrales
        if risk_input.debt_ratio > 65:
            recommendations.append("Lower debt ratio before opening the round.")
        if risk_input.ltv > 75:
            recommendations.append("Increase equity buffer to reduce LTV risk.")
        if risk_input.location_score < 50:
            recommendations.append("Run deeper location due diligence.")
        if risk_input.sponsor_track_record < 55:
            recommendations.append("Request stronger sponsor guarantees.")

        # Añadimos recomendaciones que vienen del backend/mercado
        recommendations.extend(rag_context["backend_recommendations"])
        
        # Eliminamos duplicados manteniendo el orden
        unique_recommendations = list(dict.fromkeys(recommendations))

        # Guardamos la evaluación en SQLite para historial
        self.rag_store.save_evaluation(
            project_stage=risk_input.project_stage,
            ltv=risk_input.ltv,
            debt_ratio=risk_input.debt_ratio,
            location_score=risk_input.location_score,
            sponsor_track_record=risk_input.sponsor_track_record,
            risk_score=score_result["risk_score"],
            risk_level=score_result["risk_level"],
            confidence=score_result["confidence"],
        )

        return EngineOutput(
            risk_score=score_result["risk_score"],
            risk_level=score_result["risk_level"],
            confidence=score_result["confidence"],
            recommendations=unique_recommendations,
        )