import logging
import time
from contextlib import asynccontextmanager
from typing import Literal
from fastapi import FastAPI, HTTPException, Request, status
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from scoring.calculateTerraScore import StageType, RiskLevel
from scoring.engine import RiskInput, TerraScoringEngine
from scoring.rag import RagStore

# Structured JSON logging setup
logging.basicConfig(
    level=logging.INFO,
    format='{"time": "%(asctime)s", "level": "%(levelname)s", "logger": "%(name)s", "message": "%(message)s"}'
)
logger = logging.getLogger("ai_module")

class RiskRequest(BaseModel):
    ltv: float = Field(..., ge=0.0, le=100.0, description="Loan-To-Value ratio (0-100)")
    debtRatio: float = Field(..., ge=0.0, le=100.0, description="Debt ratio (0-100)")
    locationScore: float = Field(..., ge=0.0, le=100.0, description="Location quality index (0-100)")
    sponsorTrackRecord: float = Field(..., ge=0.0, le=100.0, description="Sponsor track record (0-100)")
    projectStage: StageType

class RiskResponse(BaseModel):
    riskScore: float
    riskLevel: RiskLevel
    confidence: float
    recommendations: list[str]

rag_store = RagStore.from_env()
engine = TerraScoringEngine(rag_store)

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup: initialize connection pool and tables
    await rag_store.initialize()
    logger.info("AI scoring service initialized successfully")
    yield
    # Shutdown: clean up connections
    await rag_store.close()
    logger.info("AI scoring service connection pools closed")

app = FastAPI(title="TerraFertil AI Module", version="1.0.0", lifespan=lifespan)

@app.middleware("http")
async def tracing_middleware(request: Request, call_next):
    trace_id = request.headers.get("X-Request-ID")
    if not trace_id:
        trace_id = f"ai-{int(time.time() * 1000)}"
    start_time = time.perf_counter()
    
    response = await call_next(request)
    
    elapsed_ms = round((time.perf_counter() - start_time) * 1000, 2)
    response.headers["X-Request-ID"] = trace_id
    logger.info(
        f'{{"trace_id": "{trace_id}", "method": "{request.method}", "path": "{request.url.path}", '
        f'"status": {response.status_code}, "latency_ms": {elapsed_ms}}}'
    )
    return response

@app.get("/health")
@app.get("/health/liveness")
async def health() -> dict[str, str]:
    return {"status": "ok"}

@app.get("/health/readiness")
async def readiness() -> dict[str, str]:
    return {"status": "ready"}

@app.post("/predict-risk", response_model=RiskResponse, status_code=status.HTTP_200_OK)
async def predict_risk(req: RiskRequest, request: Request) -> RiskResponse:
    trace_id = request.headers.get("X-Request-ID")
    try:
        result = await engine.score(
            RiskInput(
                ltv=req.ltv,
                debt_ratio=req.debtRatio,
                location_score=req.locationScore,
                sponsor_track_record=req.sponsorTrackRecord,
                project_stage=req.projectStage,
            ),
            trace_id=trace_id,
        )
        
        return RiskResponse(
            riskScore=result.risk_score,
            riskLevel=result.risk_level,
            confidence=result.confidence,
            recommendations=result.recommendations,
        )
    except Exception as ex:
        logger.error(f"Error computing risk score: {ex}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Error processing risk assessment"
        )