package service_test

import (
	"math"
	"testing"

	"terrafertil/backend-go/internal/service"
)

func TestCalculateFutureValue(t *testing.T) {
	tests := []struct {
		name        string
		principal   float64
		annualRate  float64
		years       int
		wantFinal   float64
		wantProfit  float64
	}{
		{
			name:       "valid compounding 1 year",
			principal:  1000,
			annualRate: 10,
			years:      1,
			wantFinal:  1100,
			wantProfit: 100,
		},
		{
			name:       "zero principal",
			principal:  0,
			annualRate: 10,
			years:      1,
			wantFinal:  0,
			wantProfit: 0,
		},
		{
			name:       "negative rate",
			principal:  1000,
			annualRate: -5,
			years:      1,
			wantFinal:  0,
			wantProfit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFinal, gotProfit := service.CalculateFutureValue(tt.principal, tt.annualRate, tt.years)
			if math.Abs(gotFinal-tt.wantFinal) > 0.01 {
				t.Errorf("got final %v, want %v", gotFinal, tt.wantFinal)
			}
			if math.Abs(gotProfit-tt.wantProfit) > 0.01 {
				t.Errorf("got profit %v, want %v", gotProfit, tt.wantProfit)
			}
		})
	}
}
