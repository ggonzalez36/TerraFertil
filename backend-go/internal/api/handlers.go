package api

import (
 "bytes"
 "encoding/json"
 "fmt"
 "io"
 "log"
 "net/mail"
 "net/http"
 "strings"
 "time"

 "terrafertil/backend-go/internal/service"
 "terrafertil/backend-go/internal/store"
)

type Server struct {
 aiServiceURL string
 httpClient   *http.Client
 waitlist     *store.WaitlistStore
}

func NewServer(aiServiceURL string, waitlist *store.WaitlistStore) *Server {
 return &Server{
  aiServiceURL: aiServiceURL,
  httpClient: &http.Client{
   Timeout: 5 * time.Second,
  },
  waitlist: waitlist,
 }
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
 mux.HandleFunc("GET /health", s.handleHealth)
 mux.HandleFunc("GET /api/projects", s.handleProjects)
 mux.HandleFunc("POST /api/investment/simulate", s.handleSimulateInvestment)
 mux.HandleFunc("POST /api/ai/risk-score", s.handleRiskScore)
 mux.HandleFunc("POST /api/onboarding", s.handleOnboarding)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
 respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleProjects(w http.ResponseWriter, _ *http.Request) {
 respondJSON(w, http.StatusOK, map[string]any{"items": service.GetSeedProjects()})
}

func (s *Server) handleSimulateInvestment(w http.ResponseWriter, r *http.Request) {
 var req service.SimulateInvestmentRequest
 if err := decodeJSON(r.Body, &req); err != nil {
  respondError(w, http.StatusBadRequest, "invalid request body")
  return
 }

 if req.Principal <= 0 || req.AnnualRate <= 0 || req.Years <= 0 {
  respondError(w, http.StatusBadRequest, "principal, annualRate and years must be greater than 0")
  return
 }

 finalAmount, profit := service.CalculateFutureValue(req.Principal, req.AnnualRate, req.Years)
 respondJSON(w, http.StatusOK, service.SimulateInvestmentResponse{
  Principal:   req.Principal,
  AnnualRate:  req.AnnualRate,
  Years:       req.Years,
  FinalAmount: finalAmount,
  Profit:      profit,
 })
}

func (s *Server) handleRiskScore(w http.ResponseWriter, r *http.Request) {
 var req service.AIRiskRequest
 if err := decodeJSON(r.Body, &req); err != nil {
  respondError(w, http.StatusBadRequest, "invalid request body")
  return
 }

 payload, err := json.Marshal(req)
 if err != nil {
  respondError(w, http.StatusInternalServerError, "could not encode request")
  return
 }

 upstreamURL := fmt.Sprintf("%s/predict-risk", s.aiServiceURL)
 upstreamResp, err := s.httpClient.Post(upstreamURL, "application/json", bytes.NewReader(payload))
 if err != nil {
  log.Printf("ai service error: %v", err)
  respondError(w, http.StatusBadGateway, "ai service unavailable")
  return
 }
 defer upstreamResp.Body.Close()

 if upstreamResp.StatusCode >= 400 {
  body, _ := io.ReadAll(upstreamResp.Body)
  respondError(w, http.StatusBadGateway, fmt.Sprintf("ai service returned error: %s", string(body)))
  return
 }

 var aiResp service.AIRiskResponse
 if err := json.NewDecoder(upstreamResp.Body).Decode(&aiResp); err != nil {
  respondError(w, http.StatusBadGateway, "invalid ai service response")
  return
 }

 respondJSON(w, http.StatusOK, aiResp)
}

func (s *Server) handleOnboarding(w http.ResponseWriter, r *http.Request) {
 if s.waitlist == nil {
  respondError(w, http.StatusInternalServerError, "waitlist store not configured")
  return
 }

 var req service.OnboardingRequest
 if err := decodeJSON(r.Body, &req); err != nil {
  respondError(w, http.StatusBadRequest, "invalid request body")
  return
 }

 req.Email = strings.TrimSpace(strings.ToLower(req.Email))
 if req.Source == "" {
  req.Source = "landing"
 }

 if req.Email == "" {
  respondError(w, http.StatusBadRequest, "email is required")
  return
 }

 if _, err := mail.ParseAddress(req.Email); err != nil {
  respondError(w, http.StatusBadRequest, "invalid email format")
  return
 }

 id, createdAt, err := s.waitlist.CreateSignup(req.Email, req.Source)
 if err != nil {
  if err == store.ErrDuplicateEmail {
   respondError(w, http.StatusConflict, "email already registered")
   return
  }

  log.Printf("waitlist create error: %v", err)
  respondError(w, http.StatusInternalServerError, "could not save onboarding")
  return
 }

 respondJSON(w, http.StatusCreated, service.OnboardingResponse{
  ID:        id,
  Email:     req.Email,
  Source:    req.Source,
  CreatedAt: createdAt.Format(time.RFC3339),
  Status:    "created",
 })
}

func decodeJSON(body io.Reader, dst any) error {
 dec := json.NewDecoder(body)
 dec.DisallowUnknownFields()
 return dec.Decode(dst)
}

func respondError(w http.ResponseWriter, code int, message string) {
 respondJSON(w, code, map[string]string{"error": message})
}

func respondJSON(w http.ResponseWriter, code int, data any) {
 w.Header().Set("Content-Type", "application/json")
 w.WriteHeader(code)
 _ = json.NewEncoder(w).Encode(data)
}
