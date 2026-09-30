package service

import (
	"context"
	"math"

	"terrafertil/backend-go/internal/domain"
)

type ProjectService struct{}

func NewProjectService() *ProjectService {
	return &ProjectService{}
}

func (s *ProjectService) GetAll(ctx context.Context) ([]domain.Project, error) {
	return GetSeedProjects(), nil
}

func GetSeedProjects() []domain.Project {
	return []domain.Project{
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

// CalculateFutureValue calculates compound returns with bounds validation
func CalculateFutureValue(principal, annualRate float64, years int) (finalAmount float64, profit float64) {
	if principal <= 0 || annualRate <= 0 || years <= 0 {
		return 0, 0
	}
	rate := annualRate / 100.0
	finalAmount = principal * math.Pow(1+rate, float64(years))
	profit = finalAmount - principal
	return finalAmount, profit
}
