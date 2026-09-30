package domain

import (
	"context"
	"time"
)

type WaitlistRepository interface {
	Save(ctx context.Context, email, source string) (int64, time.Time, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

type ProjectRepository interface {
	GetAll(ctx context.Context) ([]Project, error)
}

type AIRiskClient interface {
	AssessRisk(ctx context.Context, cmd AIRiskCommand) (*AIRiskAssessment, error)
}
