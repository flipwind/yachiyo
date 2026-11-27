package qq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"yachiyo/yachiyo-util/logger"
)

var apiLog = logger.New("Yachiyo.QQ.API")

const (
	defaultBaseURL = "https://api.bot.qq.com"
)

type APIClient struct {
	baseURL             string
	tokens              *TokenProvider
	httpClient          *http.Client
	defaultMarkdown     bool
	fallbackToPlainText bool
}

type APIClientOptions struct {
	Sandbox             bool
	APIBaseURL          string
	DefaultMarkdown     bool
	FallbackToPlainText bool
	HTTPClient          *http.Client
}

func NewAPIClient(tokens *TokenProvider, opts APIClientOptions) *APIClient {
	base := strings.TrimRight(strings.TrimSpace(opts.APIBaseURL), "/")
	if base == "" {
		base = defaultBaseURL
		if opts.Sandbox {
			apiLog.Warn("qq.sandbox=true uses unified REST domain %s (legacy api.sgroup.qq.com is deprecated; set qq.api_base_url to override)", defaultBaseURL)
		}
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 8 * time.Second}
	}
	return &APIClient{
		baseURL:             base,
		tokens:              tokens,
		httpClient:          hc,
		defaultMarkdown:     opts.DefaultMarkdown,
		fallbackToPlainText: opts.FallbackToPlainText,
	}
}

func (a *APIClient) GetGatewayURL(ctx context.Context) (string, error) {
	raw, err := a.doJSON(ctx, http.MethodGet, "/gateway", nil)
	if err != nil {
		return "", err
	}
	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return "", fmt.Errorf("decode /gateway response: %w", err)
	}
	if resp.URL == "" {
		return "", errors.New("empty gateway url returned by QQ API")
	}
	return resp.URL, nil
}

type SendOptions struct {
	MsgID        string
	MsgSeq       int
	RefMessageID string
}

// SendGroupMessage sends a message to a QQ group using Markdown by default.
// When markdown fails due to permission/format rejection and fallback is enabled,
// it retries once as plain text (msg_type=0) and reports fallbackPlainText=true.
func (a *APIClient) SendGroupMessage(ctx context.Context, groupOpenID, text string, opts SendOptions) (SendMessageResponse, bool, error) {
	if groupOpenID == "" {
		return SendMessageResponse{}, false, errors.New("group_openid cannot be empty")
	}
	path := fmt.Sprintf("/v2/groups/%s/messages", url.PathEscape(groupOpenID))
	return a.sendMessage(ctx, path, "group "+groupOpenID, text, opts)
}

// SendC2CMessage sends a private (C2C) message to a QQ user using Markdown by default,
// with the same automatic plain-text fallback as SendGroupMessage.
func (a *APIClient) SendC2CMessage(ctx context.Context, userOpenID, text string, opts SendOptions) (SendMessageResponse, bool, error) {
	if userOpenID == "" {
		return SendMessageResponse{}, false, errors.New("user_openid cannot be empty")
	}
	path := fmt.Sprintf("/v2/users/%s/messages", url.PathEscape(userOpenID))
	return a.sendMessage(ctx, path, "user "+userOpenID, text, opts)
}

func (a *APIClient) sendMessage(ctx context.Context, path, targetLabel, text string, opts SendOptions) (SendMessageResponse, bool, error) {
	buildReq := func(useMarkdown bool) SendGroupMessageRequest {
		req := SendGroupMessageRequest{
			MsgID:  opts.MsgID,
			MsgSeq: opts.MsgSeq,
		}
		if opts.RefMessageID != "" {
			req.MessageReference = &MessageReference{MessageID: opts.RefMessageID}
		}
		if useMarkdown {
			req.MsgType = MsgTypeMarkdown
			req.Markdown = &MarkdownPayload{Content: text}
		} else {
			req.MsgType = MsgTypeText
			req.Content = text
		}
		return req
	}

	statusCode, raw, err := a.doJSONWithStatus(ctx, http.MethodPost, path, buildReq(a.defaultMarkdown))
	if err != nil && a.defaultMarkdown && a.fallbackToPlainText && shouldFallbackMarkdown(err) {
		apiLog.Warn("Markdown send failed for %s (%v), falling back to plain text (msg_type=0)", targetLabel, err)
		statusCode, raw, err = a.doJSONWithStatus(ctx, http.MethodPost, path, buildReq(false))
		if err != nil {
			return SendMessageResponse{}, true, err
		}
		out, decErr := decodeSendResponse(statusCode, raw)
		if decErr != nil {
			return SendMessageResponse{}, true, decErr
		}
		return out, true, nil
	}
	if err != nil {
		return SendMessageResponse{}, false, err
	}

	out, decErr := decodeSendResponse(statusCode, raw)
	if decErr != nil {
		return SendMessageResponse{}, false, decErr
	}
	return out, false, nil
}

func decodeSendResponse(statusCode int, raw json.RawMessage) (SendMessageResponse, error) {
	var out SendMessageResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return SendMessageResponse{}, fmt.Errorf("decode send message response (http=%d): %w (body=%s)", statusCode, err, string(raw))
	}
	if statusCode == http.StatusAccepted {
		out.AsyncAccepted = true
		return out, nil
	}
	if strings.TrimSpace(out.ID) == "" {
		return SendMessageResponse{}, fmt.Errorf("qq api returned http=%d without a valid message id (body=%s)", statusCode, string(raw))
	}
	return out, nil
}

func shouldFallbackMarkdown(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	if apiErr.IsAuthError() || apiErr.IsRateLimit() {
		return false
	}
	switch apiErr.ErrCode {
	case 40034127, 40054005:
		return true
	}
	switch apiErr.Code {
	case 11292, 11293, 50056:
		return true
	}
	lower := strings.ToLower(apiErr.Message)
	return strings.Contains(lower, "markdown") || strings.Contains(lower, "msg_type")
}

func (a *APIClient) RecallGroupMessage(ctx context.Context, groupOpenID, messageID string) error {
	if groupOpenID == "" || messageID == "" {
		return errors.New("group_openid and message_id are required")
	}
	path := fmt.Sprintf("/v2/groups/%s/messages/%s", url.PathEscape(groupOpenID), url.PathEscape(messageID))
	_, err := a.doJSON(ctx, http.MethodDelete, path, nil)
	return err
}

func (a *APIClient) GetGroupInfo(ctx context.Context, groupOpenID string) (GroupInfo, error) {
	path := fmt.Sprintf("/v2/groups/%s/info", url.PathEscape(groupOpenID))
	raw, err := a.doJSON(ctx, http.MethodGet, path, nil)
	if err != nil {
		return GroupInfo{}, err
	}
	var info GroupInfo
	_ = json.Unmarshal(raw, &info)
	if info.GroupOpenID == "" {
		info.GroupOpenID = groupOpenID
	}
	info.Raw = raw
	return info, nil
}

func (a *APIClient) GetGroupBotState(ctx context.Context, groupOpenID string) (json.RawMessage, error) {
	path := fmt.Sprintf("/v2/groups/%s/bot_state", url.PathEscape(groupOpenID))
	return a.doJSON(ctx, http.MethodGet, path, nil)
}

func (a *APIClient) GetGroupRobots(ctx context.Context, groupOpenID string) (json.RawMessage, error) {
	return a.GetGroupBotState(ctx, groupOpenID)
}

func (a *APIClient) GetGroupMember(ctx context.Context, groupOpenID, memberOpenID string) (GroupMemberInfo, error) {
	path := fmt.Sprintf("/v2/groups/%s/members/%s", url.PathEscape(groupOpenID), url.PathEscape(memberOpenID))
	raw, err := a.doJSON(ctx, http.MethodGet, path, nil)
	if err != nil {
		return GroupMemberInfo{}, err
	}
	var info GroupMemberInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return GroupMemberInfo{}, err
	}
	if info.MemberOpenID == "" {
		info.MemberOpenID = memberOpenID
	}
	return info, nil
}

func (a *APIClient) doJSON(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	_, raw, err := a.doJSONWithStatus(ctx, method, path, body)
	return raw, err
}

func (a *APIClient) doJSONWithStatus(ctx context.Context, method, path string, body any) (int, json.RawMessage, error) {
	token, err := a.tokens.GetToken(ctx)
	if err != nil {
		return 0, nil, err
	}

	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, a.baseURL+path, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "QQBot "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	traceID := strings.TrimSpace(resp.Header.Get("X-Tps-Trace-Id"))
	if traceID == "" {
		traceID = strings.TrimSpace(resp.Header.Get("X-Trace-Id"))
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			TraceID:    traceID,
			RawBody:    string(respBody),
		}
		_ = json.Unmarshal(respBody, apiErr)
		return resp.StatusCode, nil, apiErr
	}
	if len(respBody) == 0 {
		return resp.StatusCode, json.RawMessage("{}"), nil
	}
	return resp.StatusCode, json.RawMessage(respBody), nil
}
