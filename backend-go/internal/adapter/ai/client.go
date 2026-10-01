package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"terrafertil/backend-go/internal/domain"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

type ResilientAIClient struct {
	baseURL      string
	httpClient   *http.Client
	state        CircuitState
	failCount    int
	maxFailures  int
	timeoutReset time.Duration
	lastFailure  time.Time
	mu           sync.RWMutex
}

func NewResilientAIClient(baseURL string, maxFailures int, resetTimeout time.Duration) *ResilientAIClient {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 25,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &ResilientAIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   4 * time.Second,
		},
		maxFailures:  maxFailures,
		timeoutReset: resetTimeout,
		state:        StateClosed,
	}
}

func (c *ResilientAIClient) AssessRisk(ctx context.Context, cmd domain.AIRiskCommand) (*domain.AIRiskAssessment, error) {
	if !c.allowRequest() {
		return nil, domain.ErrServiceUnavailable
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("serialize risk command: %w", err)
	}

	maxRetries := 2
	for attempt := 0; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/predict-risk", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		// Trace ID propagation
		if reqID, ok := ctx.Value("request_id").(string); ok && reqID != "" {
			req.Header.Set("X-Request-ID", reqID)
		}

		httpResp, err := c.httpClient.Do(req)
		if err == nil {
			if httpResp.StatusCode == http.StatusOK {
				var result domain.AIRiskAssessment
				decErr := json.NewDecoder(httpResp.Body).Decode(&result)
				_ = httpResp.Body.Close()
				if decErr == nil {
					c.recordSuccess()
					return &result, nil
				}
			} else {
				_, _ = io.Copy(io.Discard, httpResp.Body)
				_ = httpResp.Body.Close()
			}
		}

		if attempt < maxRetries {
			// Exponential backoff with jitter
			backoff := time.Duration(math.Pow(2, float64(attempt))*100)*time.Millisecond + time.Duration(rand.Intn(50))*time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	c.recordFailure()
	return nil, domain.ErrServiceUnavailable
}

func (c *ResilientAIClient) allowRequest() bool {
	c.mu.RLock()
	state := c.state
	lastFail := c.lastFailure
	c.mu.RUnlock()

	if state == StateOpen {
		if time.Since(lastFail) > c.timeoutReset {
			c.mu.Lock()
			c.state = StateHalfOpen
			c.mu.Unlock()
			return true
		}
		return false
	}
	return true
}

func (c *ResilientAIClient) recordSuccess() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failCount = 0
	c.state = StateClosed
}

func (c *ResilientAIClient) recordFailure() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failCount++
	c.lastFailure = time.Now()
	if c.failCount >= c.maxFailures {
		c.state = StateOpen
	}
}
