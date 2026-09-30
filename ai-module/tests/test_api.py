import pytest
from fastapi.testclient import TestClient
from main import app

def test_health_endpoints():
    with TestClient(app) as client:
        resp = client.get("/health")
        assert resp.status_code == 200
        assert resp.json() == {"status": "ok"}

        liveness = client.get("/health/liveness")
        assert liveness.status_code == 200

        readiness = client.get("/health/readiness")
        assert readiness.status_code == 200

def test_predict_risk_success():
    with TestClient(app) as client:
        payload = {
            "ltv": 55.0,
            "debtRatio": 40.0,
            "locationScore": 85.0,
            "sponsorTrackRecord": 80.0,
            "projectStage": "construction"
        }
        resp = client.post("/predict-risk", json=payload, headers={"X-Request-ID": "test-trace-123"})
        assert resp.status_code == 200
        data = resp.json()
        assert "riskScore" in data
        assert "riskLevel" in data
        assert "confidence" in data
        assert "recommendations" in data
        assert resp.headers.get("X-Request-ID") == "test-trace-123"

def test_predict_risk_validation_error():
    with TestClient(app) as client:
        payload = {
            "ltv": 150.0, # Invalid > 100
            "debtRatio": 40.0,
            "locationScore": 85.0,
            "sponsorTrackRecord": 80.0,
            "projectStage": "construction"
        }
        resp = client.post("/predict-risk", json=payload)
        assert resp.status_code == 422 # Unprocessable Entity by Pydantic
