package service

import "math"

func GetSeedProjects() []Project {
	return []Project{
		{
			ID:                "proj-001",
			Name:              "Residencial Los Álamos",
			City:              "Quito",
			Country:           "Ecuador",
			ExpectedReturn:    12.5,
			MinimumInvestment: 500.0,
			FundedPercent:     45,
		},
		{
			ID:                "proj-002",
			Name:              "Torre Bicentenario",
			City:              "Bogotá",
			Country:           "Colombia",
			ExpectedReturn:    14.0,
			MinimumInvestment: 1000.0,
			FundedPercent:     80,
		},
	}
}

func CalculateFutureValue(principal, annualRate float64, years int) (finalAmount float64, profit float64) {
	rate := annualRate / 100.0
	finalAmount = principal * math.Pow(1+rate, float64(years))
	profit = finalAmount - principal
	return finalAmount, profit
}
