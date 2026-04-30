package service

type Project struct {
 ID              string  `json:"id"`
 Name            string  `json:"name"`
 City            string  `json:"city"`
 Country         string  `json:"country"`
 ExpectedReturn  float64 `json:"expectedReturn"`
 MinimumInvestment float64 `json:"minimumInvestment"`
 FundedPercent   int     `json:"fundedPercent"`
}

type SimulateInvestmentRequest struct {
 Principal  float64 `json:"principal"`
 AnnualRate float64 `json:"annualRate"`
 Years      int     `json:"years"`
}

type SimulateInvestmentResponse struct {
 Principal   float64 `json:"principal"`
 AnnualRate  float64 `json:"annualRate"`
 Years       int     `json:"years"`
 FinalAmount float64 `json:"finalAmount"`
 Profit      float64 `json:"profit"`
}

type AIRiskRequest struct {
 LTV               float64 `json:"ltv"`
 DebtRatio         float64 `json:"debtRatio"`
 LocationScore     float64 `json:"locationScore"`
 SponsorTrackRecord float64 `json:"sponsorTrackRecord"`
 ProjectStage      string  `json:"projectStage"`
}

type AIRiskResponse struct {
 RiskScore       float64  `json:"riskScore"`
 RiskLevel       string   `json:"riskLevel"`
 Confidence      float64  `json:"confidence"`
 Recommendations []string `json:"recommendations"`
}

type OnboardingRequest struct {
 Email  string `json:"email"`
 Source string `json:"source"`
}

type OnboardingResponse struct {
 ID        int64  `json:"id"`
 Email     string `json:"email"`
 Source    string `json:"source"`
 CreatedAt string `json:"createdAt"`
 Status    string `json:"status"`
}
