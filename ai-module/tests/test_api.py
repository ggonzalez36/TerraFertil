import unittest
from fastapi.testclient import TestClient
from main import app

class TestAPIEndpoints(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.client = TestClient(app)

    @classmethod
    def tearDownClass(cls):
        cls.client.close()

    def test_health_endpoints(self):
        resp = self.client.get("/health")
        self.assertEqual(resp.status_code, 200)
        self.assertEqual(resp.json(), {"status": "ok"})

        liveness = self.client.get("/health/liveness")
        self.assertEqual(liveness.status_code, 200)

        readiness = self.client.get("/health/readiness")
        self.assertEqual(readiness.status_code, 200)

    def test_predict_risk_success(self):
        payload = {
            "ltv": 55.0,
            "debtRatio": 40.0,
            "locationScore": 85.0,
            "sponsorTrackRecord": 80.0,
            "projectStage": "construction"
        }
        resp = self.client.post("/predict-risk", json=payload, headers={"X-Request-ID": "test-trace-123"})
        self.assertEqual(resp.status_code, 200)
        data = resp.json()
        self.assertIn("riskScore", data)
        self.assertIn("riskLevel", data)
        self.assertIn("confidence", data)
        self.assertIn("recommendations", data)
        self.assertEqual(resp.headers.get("X-Request-ID"), "test-trace-123")

    def test_predict_risk_validation_error(self):
        payload = {
            "ltv": 150.0, # Invalid > 100
            "debtRatio": 40.0,
            "locationScore": 85.0,
            "sponsorTrackRecord": 80.0,
            "projectStage": "construction"
        }
        resp = self.client.post("/predict-risk", json=payload)
        self.assertEqual(resp.status_code, 422)

if __name__ == "__main__":
    unittest.main()
