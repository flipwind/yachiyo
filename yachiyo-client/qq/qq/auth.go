package qq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const accessTokenURL = "https://bots.qq.com/app/getAppAccessToken"

type TokenProvider struct {
	appID      string
	appSecret  string
	httpClient *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func NewTokenProvider(appID, appSecret string, httpClient *http.Client) *TokenProvider {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &TokenProvider{
		appID:      strings.TrimSpace(appID),
		appSecret:  strings.TrimSpace(appSecret),
		httpClient: httpClient,
	}
}

func (p *TokenProvider) GetToken(ctx context.Context) (string, error) {
	p.mu.Lock()
	if p.token != "" && time.Now().Add(120*time.Second).Before(p.expiresAt) {
		tok := p.token
		p.mu.Unlock()
		return tok, nil
	}
	p.mu.Unlock()

	return p.Refresh(ctx)
}

func (p *TokenProvider) Refresh(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring lock.
	if p.token != "" && time.Now().Add(120*time.Second).Before(p.expiresAt) {
		return p.token, nil
	}

	if p.appID == "" || p.appSecret == "" {
		return "", errors.New("qq app_id and app_secret are required")
	}

	reqBody, _ := json.Marshal(map[string]string{
		"appId":        p.appID,
		"clientSecret": p.appSecret,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, accessTokenURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request access_token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read access_token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("access_token HTTP %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   any    `json:"expires_in"`
		Code        int    `json:"code"`
		Message     string `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode access_token response: %w", err)
	}
	if parsed.AccessToken == "" {
		return "", fmt.Errorf("empty access_token (code=%d, msg=%s)", parsed.Code, parsed.Message)
	}

	ttlSec := parseExpiresIn(parsed.ExpiresIn)
	if ttlSec <= 0 {
		ttlSec = 7200
	}

	p.token = parsed.AccessToken
	p.expiresAt = time.Now().Add(time.Duration(ttlSec) * time.Second)
	return p.token, nil
}

// StartAutoRefresh keeps the access token warm in the background so 5s tool calls
// never block waiting for a token refresh.
func (p *TokenProvider) StartAutoRefresh(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				refreshCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				_, _ = p.GetToken(refreshCtx)
				cancel()
			}
		}
	}()
}

func parseExpiresIn(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err == nil {
			return n
		}
	}
	return 0
}
