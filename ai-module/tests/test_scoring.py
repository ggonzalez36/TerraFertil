import unittest
from scoring.calculateTerraScore import calculate_terra_score

class TestCalculateTerraScore(unittest.TestCase):
    def test_calculate_terra_score_low_risk(self):
        result = calculate_terra_score(
            ltv=20.0,
            debt_ratio=15.0,
            location_score=95.0,
            sponsor_track_record=95.0,
            project_stage="stabilized",
        )
        self.assertEqual(result["risk_level"], "low")
        self.assertLess(float(result["risk_score"]), 33.0)
        self.assertEqual(result["confidence"], 0.82)

    def test_calculate_terra_score_high_risk(self):
        result = calculate_terra_score(
            ltv=90.0,
            debt_ratio=85.0,
            location_score=20.0,
            sponsor_track_record=20.0,
            project_stage="land",
        )
        self.assertEqual(result["risk_level"], "high")
        self.assertGreater(float(result["risk_score"]), 66.0)
        self.assertEqual(result["confidence"], 0.68)

    def test_calculate_terra_score_clamping(self):
        extreme_high = calculate_terra_score(
            ltv=100.0,
            debt_ratio=100.0,
            location_score=0.0,
            sponsor_track_record=0.0,
            project_stage="land",
        )
        self.assertLessEqual(float(extreme_high["risk_score"]), 100.0)
        self.assertGreaterEqual(float(extreme_high["risk_score"]), 0.0)

if __name__ == "__main__":
    unittest.main()
