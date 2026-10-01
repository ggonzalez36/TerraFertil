package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"terrafertil/backend-go/internal/domain"
	"terrafertil/backend-go/internal/service"
)

type Server struct {
	aiClient     domain.AIRiskClient
	waitlistRepo domain.WaitlistRepository
	projectRepo  domain.ProjectRepository
}

func NewServer(aiClient domain.AIRiskClient, waitlist domain.WaitlistRepository, projects domain.ProjectRepository) *Server {
	return &Server{
		aiClient:     aiClient,
		waitlistRepo: waitlist,
		projectRepo:  projects,
	}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /api/projects", s.handleProjects)
	mux.HandleFunc("POST /api/investment/simulate", s.handleSimulateInvestment)
	mux.HandleFunc("POST /api/ai/risk-score", s.handleRiskScore)
	mux.HandleFunc("POST /api/onboarding", s.handleOnboarding)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)
}

type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitempty"`
	// Backwards compatibility with frontend clients
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleProjects(w http.ResponseWriter, r *http.Request) {
	if s.projectRepo == nil {
		respondJSON(w, http.StatusOK, map[string]any{"items": service.GetSeedProjects()})
		return
	}
	projects, err := s.projectRepo.GetAll(r.Context())
	if err != nil {
		respondProblem(w, r, http.StatusInternalServerError, "Internal Server Error", "Could not fetch projects")
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"items": projects})
}

func (s *Server) handleSimulateInvestment(w http.ResponseWriter, r *http.Request) {
	var req domain.SimulateInvestmentRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		respondProblem(w, r, http.StatusBadRequest, "Invalid Request Body", "Malformed JSON payload or unexpected fields")
		return
	}

	// Boundary and sanity checks
	if req.Principal <= 0 || req.Principal > 100_000_000 {
		respondProblem(w, r, http.StatusUnprocessableEntity, "Validation Error", "Principal must be greater than 0 and up to 100,000,000")
		return
	}
	if req.AnnualRate <= 0 || req.AnnualRate > 100 {
		respondProblem(w, r, http.StatusUnprocessableEntity, "Validation Error", "Annual rate must be between 0.1% and 100%")
		return
	}
	if req.Years <= 0 || req.Years > 50 {
		respondProblem(w, r, http.StatusUnprocessableEntity, "Validation Error", "Investment duration must be between 1 and 50 years")
		return
	}

	finalAmount, profit := service.CalculateFutureValue(req.Principal, req.AnnualRate, req.Years)
	respondJSON(w, http.StatusOK, domain.SimulateInvestmentResponse{
		Principal:   req.Principal,
		AnnualRate:  req.AnnualRate,
		Years:       req.Years,
		FinalAmount: finalAmount,
		Profit:      profit,
	})
}

func (s *Server) handleRiskScore(w http.ResponseWriter, r *http.Request) {
	if s.aiClient == nil {
		respondProblem(w, r, http.StatusServiceUnavailable, "Service Unavailable", "AI risk assessment client not configured")
		return
	}

	var req domain.AIRiskCommand
	if err := decodeJSON(r.Body, &req); err != nil {
		respondProblem(w, r, http.StatusBadRequest, "Invalid Request Body", "Malformed JSON or unexpected fields")
		return
	}

	// Input domain validation
	if req.LTV < 0 || req.LTV > 100 || req.DebtRatio < 0 || req.DebtRatio > 100 ||
		req.LocationScore < 0 || req.LocationScore > 100 || req.SponsorTrackRecord < 0 || req.SponsorTrackRecord > 100 {
		respondProblem(w, r, http.StatusUnprocessableEntity, "Validation Error", "LTV, DebtRatio, LocationScore, and SponsorTrackRecord must be between 0 and 100")
		return
	}

	assessment, err := s.aiClient.AssessRisk(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrServiceUnavailable) {
			respondProblem(w, r, http.StatusServiceUnavailable, "Service Degraded", "AI service is currently unavailable or circuit is open")
			return
		}
		respondProblem(w, r, http.StatusBadGateway, "AI Service Error", "Failed to communicate with AI evaluation service")
		return
	}

	respondJSON(w, http.StatusOK, assessment)
}

func (s *Server) handleOnboarding(w http.ResponseWriter, r *http.Request) {
	if s.waitlistRepo == nil {
		respondProblem(w, r, http.StatusInternalServerError, "Configuration Error", "Waitlist store not configured")
		return
	}

	var req domain.OnboardingRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		respondProblem(w, r, http.StatusBadRequest, "Invalid Request Body", "Malformed JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Source == "" {
		req.Source = "landing"
	}

	if req.Email == "" {
		respondProblem(w, r, http.StatusBadRequest, "Validation Error", "Email is required")
		return
	}

	if _, err := mail.ParseAddress(req.Email); err != nil {
		respondProblem(w, r, http.StatusBadRequest, "Validation Error", "Invalid email format")
		return
	}

	id, createdAt, err := s.waitlistRepo.Save(r.Context(), req.Email, req.Source)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateEmail) {
			respondProblem(w, r, http.StatusConflict, "Duplicate Registration", "Ese correo ya está registrado en la lista de espera.")
			return
		}

		Logger.Error("waitlist save failure", "error", err, "email", req.Email)
		respondProblem(w, r, http.StatusInternalServerError, "Persistence Error", "Could not complete onboarding registration")
		return
	}

	respondJSON(w, http.StatusCreated, domain.OnboardingResponse{
		ID:        id,
		Email:     req.Email,
		Source:    req.Source,
		CreatedAt: createdAt,
		Status:    "created",
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := decodeJSON(r.Body, &req); err != nil {
		respondProblem(w, r, http.StatusBadRequest, "Invalid Request Body", "Malformed JSON payload")
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		respondProblem(w, r, http.StatusBadRequest, "Validation Error", "Email and password are required")
		return
	}

	// Basic format validation
	if _, err := mail.ParseAddress(req.Email); err != nil {
		respondProblem(w, r, http.StatusBadRequest, "Validation Error", "Invalid email address format")
		return
	}

	// Mock production token with claim signature for MVP
	respondJSON(w, http.StatusOK, domain.LoginResponse{
		Token: "terrafertil-auth-jwt-" + req.Email + "-valid",
		Email: req.Email,
	})
}

func decodeJSON(r io.Reader, dst any) error {
	// 1MB limit prevents memory exhaustion attacks (DoS)
	dec := json.NewDecoder(io.LimitReader(r, 1<<20))
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func respondJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

func respondProblem(w http.ResponseWriter, r *http.Request, code int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ProblemDetails{
		Type:     "about:blank",
		Title:    title,
		Status:   code,
		Detail:   detail,
		Instance: r.URL.Path,
		Error:    detail,
		Message:  detail,
	})
}
