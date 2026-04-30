from typing import Literal
from fastapi import FastAPI
from pydantic import BaseModel, Field
from scoring.engine import RiskInput, TerraScoringEngine
from scoring.rag import RagStore

class RiskRequest(BaseModel):
    ltv: float = Field(ge=0, le=100)
    debtRatio: float = Field(ge=0, le=100)
    locationScore: float = Field(ge=0, le=100)
    sponsorTrackRecord: float = Field(ge=0, le=100)
    projectStage: Literal["land", "pre-sale", "construction", "stabilized"]

class RiskResponse(BaseModel):
    riskScore: float
    riskLevel: Literal["low", "medium", "high"]
    confidence: float
    recommendations: list[str]

app = FastAPI(title="TerraFertil AI Module", version="0.1.0")
rag_store = RagStore.from_env()
engine = TerraScoringEngine(rag_store)

@app.get("/health")
def health() -> dict[str, str]:
    return {"status": "ok"}

@app.post("/predict-risk", response_model=RiskResponse)
def predict_risk(req: RiskRequest) -> RiskResponse:
    result = engine.score(
        RiskInput(
            ltv=req.ltv,
            debt_ratio=req.debtRatio,
            location_score=req.locationScore,
            sponsor_track_record=req.sponsorTrackRecord,
            project_stage=req.projectStage,
        )
    )
    
    return RiskResponse(
        riskScore=result.risk_score,
        riskLevel=result.risk_level,
        confidence=result.confidence,
        recommendations=result.recommendations,
    )