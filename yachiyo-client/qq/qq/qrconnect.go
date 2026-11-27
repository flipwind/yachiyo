package qq

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"rsc.io/qr"
)

// BindStatus mirrors @tencent-connect/qqbot-connector BindStatus enum.
type BindStatus int

const (
	BindStatusNone      BindStatus = 0
	BindStatusPending   BindStatus = 1
	BindStatusCompleted BindStatus = 2
	BindStatusExpired   BindStatus = 3
)

const (
	qrHostProduction = "q.qq.com"
	qrHostTest       = "test.q.qq.com"
	qrPollInterval   = 2 * time.Second
	qrHTTPTimeout    = 10 * time.Second
)

type QRCredentials struct {
	AppID      string
	AppSecret  string
	UserOpenID string
}

type BindTask struct {
	TaskID string
	Key    string
}

type BindPollResult struct {
	Status           BindStatus
	BotAppID         string
	BotEncryptSecret string
	UserOpenID       string
}

type QRConnectOptions struct {
	Source               string
	Env                  string // "production" (default) or "test"
	DisplayQRToConsole   bool
	HTTPClient           *http.Client
	OnQRDisplayed        func(connectURL string)
	OnQRExpired          func()
}

func getQRHost(env string) string {
	if env == "test" {
		return qrHostTest
	}
	return qrHostProduction
}

// GenerateBindKey generates a 32-byte random key encoded in standard base64,
// matching @tencent-connect/qqbot-connector's generateBindKey().
func GenerateBindKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}

// DecryptSecret decrypts bot_encrypt_secret using AES-256-GCM.
// Wire format: base64( 12-byte nonce || ciphertext || 16-byte auth tag ).
func DecryptSecret(encryptSecretB64, keyB64 string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(strings.TrimSpace(keyB64))
	if err != nil {
		return "", fmt.Errorf("decode bind key: %w", err)
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(encryptSecretB64))
	if err != nil {
		return "", fmt.Errorf("decode encrypted secret: %w", err)
	}
	if len(raw) < 12+16 {
		return "", errors.New("encrypted secret payload too short")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("init aes cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("init gcm: %w", err)
	}

	nonce := raw[:12]
	ciphertextAndTag := raw[12:]
	plain, err := gcm.Open(nil, nonce, ciphertextAndTag, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: %w", err)
	}
	return string(plain), nil
}

// BuildConnectURL builds the Mobile QQ scan URL matching @tencent-connect/qqbot-connector.
func BuildConnectURL(taskID, source string) string {
	return fmt.Sprintf(
		"https://%s/qqbot/openclaw/connect.html?task_id=%s&source=%s&_wv=2",
		qrHostProduction,
		url.QueryEscape(taskID),
		url.QueryEscape(source),
	)
}

func CreateBindTask(ctx context.Context, env string, hc *http.Client) (BindTask, error) {
	key, err := GenerateBindKey()
	if err != nil {
		return BindTask{}, err
	}

	endpoint := fmt.Sprintf("https://%s/lite/create_bind_task", getQRHost(env))
	var resp struct {
		RetCode int    `json:"retcode"`
		Msg     string `json:"msg"`
		Data    struct {
			TaskID string `json:"task_id"`
		} `json:"data"`
	}
	if err := postQRJSON(ctx, hc, endpoint, map[string]string{"key": key}, &resp); err != nil {
		return BindTask{}, err
	}
	if resp.RetCode != 0 {
		msg := resp.Msg
		if msg == "" {
			msg = "create_bind_task failed"
		}
		return BindTask{}, errors.New(msg)
	}
	if resp.Data.TaskID == "" {
		return BindTask{}, errors.New("create_bind_task: missing task_id")
	}
	return BindTask{TaskID: resp.Data.TaskID, Key: key}, nil
}

func PollBindResult(ctx context.Context, taskID, env string, hc *http.Client) (BindPollResult, error) {
	endpoint := fmt.Sprintf("https://%s/lite/poll_bind_result", getQRHost(env))
	var resp struct {
		RetCode int    `json:"retcode"`
		Msg     string `json:"msg"`
		Data    struct {
			Status           *int   `json:"status"`
			BotAppID         any    `json:"bot_appid"`
			BotEncryptSecret string `json:"bot_encrypt_secret"`
			UserOpenID       string `json:"user_openid"`
		} `json:"data"`
	}
	if err := postQRJSON(ctx, hc, endpoint, map[string]string{"task_id": taskID}, &resp); err != nil {
		return BindPollResult{}, err
	}
	if resp.RetCode != 0 {
		msg := resp.Msg
		if msg == "" {
			msg = "poll_bind_result failed"
		}
		return BindPollResult{}, errors.New(msg)
	}

	status := BindStatusNone
	if resp.Data.Status != nil {
		status = BindStatus(*resp.Data.Status)
	}
	return BindPollResult{
		Status:           status,
		BotAppID:         stringifyAppID(resp.Data.BotAppID),
		BotEncryptSecret: resp.Data.BotEncryptSecret,
		UserOpenID:       resp.Data.UserOpenID,
	}, nil
}

func stringifyAppID(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return fmt.Sprint(t)
	}
}

func postQRJSON(ctx context.Context, hc *http.Client, endpoint string, payload any, out any) error {
	if hc == nil {
		hc = &http.Client{Timeout: qrHTTPTimeout}
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reqCtx, cancel := context.WithTimeout(ctx, qrHTTPTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("HTTP %d from %s", resp.StatusCode, endpoint)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, out)
}

// RenderTerminalQR renders a compact terminal QR code using Unicode half-blocks
// (white background / black modules, matching qrcode-terminal { small: true }).
func RenderTerminalQR(content string) (string, error) {
	code, err := qr.Encode(content, qr.L)
	if err != nil {
		return "", err
	}

	size := code.Size
	const border = 1
	total := size + border*2

	isDark := func(x, y int) bool {
		if x < 0 || y < 0 || x >= size || y >= size {
			return false
		}
		return code.Black(x, y)
	}

	var sb strings.Builder
	for y := -border; y < size+border; y += 2 {
		for x := -border; x < size+border; x++ {
			top := isDark(x, y)
			bottom := false
			if y+1 < size+border {
				bottom = isDark(x, y+1)
			}
			// Inverted for dark terminal backgrounds: light modules use full/half blocks.
			switch {
			case !top && !bottom:
				sb.WriteRune('█')
			case !top && bottom:
				sb.WriteRune('▀')
			case top && !bottom:
				sb.WriteRune('▄')
			default:
				sb.WriteRune(' ')
			}
		}
		sb.WriteByte('\n')
	}
	_ = total
	return sb.String(), nil
}

// QRConnect runs the interactive Mobile QQ QR bind flow until the user confirms
// authorization or ctx is cancelled.
func QRConnect(ctx context.Context, opts QRConnectOptions) (QRCredentials, error) {
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: qrHTTPTimeout}
	}

	for {
		if ctx.Err() != nil {
			return QRCredentials{}, ctx.Err()
		}

		task, err := CreateBindTask(ctx, opts.Env, hc)
		if err != nil {
			return QRCredentials{}, fmt.Errorf("获取绑定任务失败: %w", err)
		}

		connectURL := BuildConnectURL(task.TaskID, opts.Source)
		if opts.DisplayQRToConsole {
			if qrArt, err := RenderTerminalQR(connectURL); err == nil {
				fmt.Println(qrArt)
			}
			fmt.Printf("请使用手机 QQ 扫描上方二维码，完成机器人绑定。\n链接: %s\n\n", connectURL)
		}
		if opts.OnQRDisplayed != nil {
			opts.OnQRDisplayed(connectURL)
		}

		creds, expired, err := pollUntilDone(ctx, task.TaskID, task.Key, opts.Env, hc)
		if err != nil {
			return QRCredentials{}, err
		}
		if !expired {
			return creds, nil
		}

		if opts.OnQRExpired != nil {
			opts.OnQRExpired()
		}
		if opts.DisplayQRToConsole {
			fmt.Println("二维码已过期，正在刷新…")
		}
	}
}

func pollUntilDone(ctx context.Context, taskID, key, env string, hc *http.Client) (QRCredentials, bool, error) {
	for {
		if ctx.Err() != nil {
			return QRCredentials{}, false, ctx.Err()
		}

		res, err := PollBindResult(ctx, taskID, env, hc)
		if err == nil {
			switch res.Status {
			case BindStatusCompleted:
				secret, decErr := DecryptSecret(res.BotEncryptSecret, key)
				if decErr != nil {
					return QRCredentials{}, false, decErr
				}
				return QRCredentials{
					AppID:      res.BotAppID,
					AppSecret:  secret,
					UserOpenID: res.UserOpenID,
				}, false, nil
			case BindStatusExpired:
				return QRCredentials{}, true, nil
			}
		}

		timer := time.NewTimer(qrPollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return QRCredentials{}, false, ctx.Err()
		case <-timer.C:
		}
	}
}
