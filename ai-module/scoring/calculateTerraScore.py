from typing import Literal, Union

def calculate_terra_score(
    *,
    ltv: float,
    debt_ratio: float,
    location_score: float,
    sponsor_track_record: float,
    project_stage: Literal["land", "pre-sale", "construction", "stabilized"],
    stage_adjustment: dict[str, float],
    confidence_base: dict[str, float],
) -> dict[str, Union[float, str]]:
    
    # Algoritmo de scoring de TerraFértil
    risk = (
        35.0
        + (debt_ratio * 0.35)
        + (ltv * 0.30)
        - (location_score * 0.20)
        - (sponsor_track_record * 0.15)
        + stage_adjustment[project_stage]
    )

    # Normalización del riesgo
    risk = max(0.0, min(100.0, risk))

    if risk <= 33:
        level: Literal["low", "medium", "high"] = "low"
    elif risk <= 66:
        level = "medium"
    else:
        level = "high"

    confidence = confidence_base.get(project_stage, 0.74)

    return {
        "risk_score": round(risk, 2),
        "risk_level": level,
        "confidence": confidence,
    }
