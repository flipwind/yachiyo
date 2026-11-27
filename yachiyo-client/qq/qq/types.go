package qq

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Intents for QQ Bot Gateway v2.
const (
	IntentGroupAndC2CEvent = 1 << 25
)

// Gateway Opcodes.
const (
	OpDispatch        = 0
	OpHeartbeat       = 1
	OpIdentify        = 2
	OpResume          = 6
	OpReconnect       = 7
	OpInvalidSession  = 9
	OpHello           = 10
	OpHeartbeatACK    = 11
)

// Message types for POST /v2/groups/{group_openid}/messages.
const (
	MsgTypeText     = 0
	MsgTypeMarkdown = 2
	MsgTypeMedia    = 7
)

type GatewayPayload struct {
	Op int             `json:"op"`
	D  json.RawMessage `json:"d,omitempty"`
	S  *int64          `json:"s,omitempty"`
	T  string          `json:"t,omitempty"`
}

type HelloData struct {
	HeartbeatInterval int64 `json:"heartbeat_interval"`
}

type IdentifyData struct {
	Token      string `json:"token"`
	Intents    int    `json:"intents"`
	Shard      [2]int `json:"shard"`
	Properties any    `json:"properties,omitempty"`
}

type ResumeData struct {
	Token     string `json:"token"`
	SessionID string `json:"session_id"`
	Seq       int64  `json:"seq"`
}

type ReadyData struct {
	Version   int       `json:"version"`
	SessionID string    `json:"session_id"`
	User      ReadyUser `json:"user"`
}

type ReadyUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Bot      bool   `json:"bot"`
}

type GroupAuthor struct {
	ID           string `json:"id"`
	MemberOpenID string `json:"member_openid"`
	Username     string `json:"username"`
	MemberRole   string `json:"member_role"`
	Bot          bool   `json:"bot"`
}

type Attachment struct {
	ContentType string `json:"content_type"`
	Filename    string `json:"filename"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

type MentionUser struct {
	ID           string `json:"id"`
	MemberOpenID string `json:"member_openid"`
	Username     string `json:"username"`
	Bot          bool   `json:"bot"`
	IsYou        bool   `json:"is_you,omitempty"`
}

type MessageScene struct {
	Source string   `json:"source,omitempty"`
	Ext    []string `json:"ext,omitempty"`
}

// ExtractMsgIdx returns the msg_idx value from message_scene.ext (used as message_reference.message_id).
func ExtractMsgIdx(scene *MessageScene) string {
	if scene == nil {
		return ""
	}
	for _, entry := range scene.Ext {
		if strings.HasPrefix(entry, "msg_idx=") {
			if idx := strings.TrimSpace(strings.TrimPrefix(entry, "msg_idx=")); idx != "" {
				return idx
			}
		}
	}
	return ""
}

type GroupMessageEvent struct {
	ID           string        `json:"id"`
	GroupOpenID  string        `json:"group_openid"`
	GroupID      string        `json:"group_id"`
	Content      string        `json:"content"`
	Timestamp    string        `json:"timestamp"`
	Author       GroupAuthor   `json:"author"`
	Attachments  []Attachment  `json:"attachments"`
	Mentions     []MentionUser `json:"mentions"`
	MessageScene *MessageScene `json:"message_scene,omitempty"`
}

type GroupRobotEvent struct {
	Timestamp   int64  `json:"timestamp"`
	GroupOpenID string `json:"group_openid"`
	OpMember    string `json:"op_member_openid"`
}

type C2CAuthor struct {
	ID         string `json:"id"`
	UserOpenID string `json:"user_openid"`
	Username   string `json:"username"`
}

type C2CMessageEvent struct {
	ID           string        `json:"id"`
	Content      string        `json:"content"`
	Timestamp    string        `json:"timestamp"`
	Author       C2CAuthor     `json:"author"`
	Attachments  []Attachment  `json:"attachments"`
	MessageScene *MessageScene `json:"message_scene,omitempty"`
}

type FriendEvent struct {
	Timestamp int64  `json:"timestamp"`
	OpenID    string `json:"openid"`
}

type MarkdownPayload struct {
	Content string `json:"content"`
}

type MessageReference struct {
	MessageID string `json:"message_id"`
}

type SendGroupMessageRequest struct {
	MsgType          int               `json:"msg_type"`
	Content          string            `json:"content,omitempty"`
	Markdown         *MarkdownPayload  `json:"markdown,omitempty"`
	MsgID            string            `json:"msg_id,omitempty"`
	EventID          string            `json:"event_id,omitempty"`
	MsgSeq           int               `json:"msg_seq,omitempty"`
	MessageReference *MessageReference `json:"message_reference,omitempty"`
}

type SendMessageExtInfo struct {
	RefIdx string `json:"ref_idx,omitempty"`
}

type SendMessageResponse struct {
	ID            string              `json:"id"`
	Timestamp     any                 `json:"timestamp"`
	ExtInfo       *SendMessageExtInfo `json:"ext_info,omitempty"`
	AsyncAccepted bool                `json:"-"`
}

func (r SendMessageResponse) RefIdx() string {
	if r.ExtInfo == nil {
		return ""
	}
	return strings.TrimSpace(r.ExtInfo.RefIdx)
}

type GroupInfo struct {
	GroupOpenID      string          `json:"group_openid"`
	GroupName        string          `json:"group_name,omitempty"`
	GroupFingerMemo  string          `json:"group_finger_memo,omitempty"`
	GroupClassText   string          `json:"group_class_text,omitempty"`
	GroupTags        []string        `json:"group_tags,omitempty"`
	GroupMemberNum   int             `json:"group_member_num,omitempty"`
	UnionGroupOpenID string          `json:"union_group_openid,omitempty"`
	Raw              json.RawMessage `json:"raw,omitempty"`
}

type GroupMemberInfo struct {
	MemberOpenID string `json:"member_openid"`
	Username     string `json:"username"`
	MemberRole   string `json:"member_role"`
	Bot          bool   `json:"bot"`
	JoinedAt     string `json:"joined_at"`
}

type APIError struct {
	StatusCode int    `json:"-"`
	Code       int    `json:"code"`
	ErrCode    int    `json:"err_code"`
	Message    string `json:"message"`
	TraceID    string `json:"-"`
	RawBody    string `json:"-"`
}

func (e *APIError) Error() string {
	parts := make([]string, 0, 4)
	if e.StatusCode != 0 {
		parts = append(parts, fmt.Sprintf("http=%d", e.StatusCode))
	}
	if e.Code != 0 {
		parts = append(parts, fmt.Sprintf("code=%d", e.Code))
	}
	if e.ErrCode != 0 {
		parts = append(parts, fmt.Sprintf("err_code=%d", e.ErrCode))
	}
	if e.TraceID != "" {
		parts = append(parts, fmt.Sprintf("trace_id=%s", e.TraceID))
	}
	detail := strings.TrimSpace(e.Message)
	if detail == "" {
		detail = strings.TrimSpace(e.RawBody)
	}
	if len(parts) == 0 {
		return detail
	}
	if detail == "" {
		return fmt.Sprintf("qq api error (%s)", strings.Join(parts, ", "))
	}
	return fmt.Sprintf("qq api error (%s): %s", strings.Join(parts, ", "), detail)
}

func (e *APIError) IsAuthError() bool {
	return e.StatusCode == http.StatusUnauthorized || e.Code == 11243 || e.Code == 11244
}

func (e *APIError) IsRateLimit() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

func (e *APIError) IsPermissionDenied() bool {
	return e.Code == 11253 || e.StatusCode == http.StatusForbidden
}

// ShouldReleasePassiveOnError reports whether a failed send should release the reserved
// passive reply slot (true for known 4xx rejections; false for 5xx/network ambiguity).
func ShouldReleasePassiveOnError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500
}
