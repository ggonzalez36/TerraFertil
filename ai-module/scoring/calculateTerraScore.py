from dataclasses import dataclass
from typing import Literal

StageType = Literal["land", "pre-sale", "construction", "stabilized"]
RiskLevel = Literal["low", "medium", "high"]

STAGE_RISK_FACTORS: dict[StageType, float] = {
    "land": 15.0,
    "pre-sale": 8.0,
    "construction": 4.0,
    "stabilized": -5.0,
}

STAGE_CONFIDENCE_FACTORS: dict[StageType, float] = {
    "land": 0.68,
    "pre-sale": 0.72,
    "construction": 0.74,
    "stabilized": 0.82,
}

@dataclass(frozen=True)
class RiskInputData:
    ltv: float
    debt_ratio: float
    location_score: float
    sponsor_track_record: float
    project_stage: StageType

@dataclass(frozen=True)
class RiskCalculationResult:
    risk_score: float
    risk_level: RiskLevel
    confidence: float

def calculate_terra_score(
    *,
    ltv: float,
    debt_ratio: float,
    location_score: float,
    sponsor_track_record: float,
    project_stage: StageType,
    stage_adjustment: dict[str, float] | None = None,
    confidence_base: dict[str, float] | None = None,
) -> dict[str, float | str]:
    """Pure domain scoring calculation with boundary protection."""
    adj = (stage_adjustment or STAGE_RISK_FACTORS).get(project_stage, 0.0)
    
    risk = (
        35.0
        + (debt_ratio * 0.35)
        + (ltv * 0.30)
        - (location_score * 0.20)
        - (sponsor_track_record * 0.15)
        + adj
    )

    bounded_risk = max(0.0, min(100.0, risk))

    if bounded_risk <= 33.0:
        level: RiskLevel = "low"
    elif bounded_risk <= 66.0:
        level = "medium"
    else:
        level = "high"

    conf = (confidence_base or STAGE_CONFIDENCE_FACTORS).get(project_stage, 0.74)

    return {
        "risk_score": round(bounded_risk, 2),
        "risk_level": level,
        "confidence": conf,
    }
