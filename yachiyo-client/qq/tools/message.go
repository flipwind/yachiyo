package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/yws"
)

const (
	maxCachedContentRunes = 500
	recallWindow          = 2 * time.Minute
)

func (r *Registry) registerMessageTools() {
	r.register(yws.ToolItem{
		Name:        "qq_get_recent_messages",
		Description: "Get up to 100 recent locally cached messages (both user messages and bot replies) in a QQ group or private chat, including message_id and sender's member_openid/user_openid. Pass either path ('group/<group_openid>' or 'private/<user_openid>') or group_openid.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"Target sub-path: 'group/<group_openid>' or 'private/<user_openid>'."},"group_openid":{"type":"string","description":"Shortcut for group chats (equivalent to path='group/<group_openid>')."},"limit":{"type":"integer","minimum":1,"maximum":100,"description":"Maximum number of recent messages to return (default 20, max 100)."}}}`),
	}, r.handleGetRecentMessages)

	r.register(yws.ToolItem{
		Name:        "qq_reply_group_message",
		Description: "Send a Markdown message in a QQ group that explicitly quotes a specific message_id. AVOID using this tool for normal conversation — always prefer commit_message (which preserves conversation history and supports mentioning users via <qqbot-at-user id=\"<member_openid>\" />). Only use when quoting an earlier message is strictly necessary.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"group_openid":{"type":"string","description":"Target QQ group_openid."},"message_id":{"type":"string","description":"The message_id to quote/reference, obtained from qq_get_recent_messages."},"content":{"type":"string","description":"Markdown message content to send. You may include <qqbot-at-user id=\"<member_openid>\" /> to mention a user."}},"required":["group_openid","message_id","content"]}`),
	}, r.handleReplyGroupMessage)

	r.register(yws.ToolItem{
		Name:        "qq_recall_group_message",
		Description: "Recall a message previously sent by THIS bot in a QQ group. Because the bot is not a group admin, it can only recall its own messages. If message_id is omitted, recalls the bot's most recent message in that group.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"group_openid":{"type":"string","description":"Target QQ group_openid."},"message_id":{"type":"string","description":"Optional message_id of the bot's own message to recall. Omit to recall the bot's last sent message in this group."}},"required":["group_openid"]}`),
	}, r.handleRecallGroupMessage)
}

func (r *Registry) handleGetRecentMessages(_ context.Context, payload json.RawMessage) (string, error) {
	var args struct {
		Path        string `json:"path"`
		GroupOpenID string `json:"group_openid"`
		Limit       int    `json:"limit"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	args.Path, args.GroupOpenID = strings.TrimSpace(args.Path), strings.TrimSpace(args.GroupOpenID)
	if args.Limit <= 0 {
		args.Limit = 20
	}
	args.Limit = min(args.Limit, 100)

	var resolvedPath string
	var msgs []store.CachedMessage
	switch {
	case strings.HasPrefix(args.Path, "private/"):
		uid := strings.TrimPrefix(args.Path, "private/")
		if uid == "" {
			return "", errors.New("invalid private path")
		}
		resolvedPath, msgs = args.Path, r.store.GetRecentPrivateMessages(uid, args.Limit)
	case strings.HasPrefix(args.Path, "group/"):
		gid := strings.TrimPrefix(args.Path, "group/")
		if gid == "" {
			return "", errors.New("invalid group path")
		}
		resolvedPath, msgs = args.Path, r.store.GetRecentMessages(gid, args.Limit)
	case args.GroupOpenID != "":
		resolvedPath, msgs = "group/"+args.GroupOpenID, r.store.GetRecentMessages(args.GroupOpenID, args.Limit)
	default:
		return "", errors.New("either path ('group/...' or 'private/...') or group_openid is required")
	}

	for i := range msgs {
		msgs[i].Content = truncateRunes(msgs[i].Content, maxCachedContentRunes)
	}
	return marshalJSON(map[string]any{"path": resolvedPath, "count": len(msgs), "messages": msgs})
}

func (r *Registry) handleReplyGroupMessage(ctx context.Context, payload json.RawMessage) (string, error) {
	var args struct {
		GroupOpenID string `json:"group_openid"`
		MessageID   string `json:"message_id"`
		Content     string `json:"content"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	gid, mid, content := strings.TrimSpace(args.GroupOpenID), strings.TrimSpace(args.MessageID), strings.TrimSpace(args.Content)
	if gid == "" || mid == "" || content == "" {
		return "", errors.New("group_openid, message_id, and content are required")
	}
	if r.api == nil {
		return "", errors.New("qq api client not initialized")
	}

	refIdx := mid
	if !strings.HasPrefix(mid, "REFIDX_") {
		cached, found := r.store.FindMessage(gid, mid)
		if !found {
			return "", fmt.Errorf("cannot quote message %s: not found in local cache so ref_idx cannot be resolved", mid)
		}
		if strings.TrimSpace(cached.RefIdx) == "" {
			return "", fmt.Errorf("cannot quote message %s: cached message is missing ref_idx", mid)
		}
		refIdx = cached.RefIdx
	}

	sendOpts := qq.SendOptions{RefMessageID: refIdx}
	if msgID, seq, ok := r.store.ReservePassiveAnchor("group", gid, ""); ok {
		sendOpts.MsgID, sendOpts.MsgSeq = msgID, seq
	} else if r.store.IsGroupRejectProactive(gid) {
		return "", fmt.Errorf("cannot send quoted group message to %s: passive reply window unavailable and group has rejected proactive messages", gid)
	}

	resp, fallback, err := r.api.SendGroupMessage(ctx, gid, content, sendOpts)
	if err != nil {
		if sendOpts.MsgID != "" && qq.ShouldReleasePassiveOnError(err) {
			r.store.ReleasePassiveAnchor("group", gid, sendOpts.MsgID, sendOpts.MsgSeq)
		}
		return "", fmt.Errorf("send quoted group message failed: %w", err)
	}
	if resp.ID != "" {
		r.store.RecordBotMessageWithRef(gid, resp.ID, resp.RefIdx(), content, time.Now().Unix())
	}
	return marshalJSON(map[string]any{
		"sent": true, "async_accepted": resp.AsyncAccepted,
		"message_id": resp.ID, "ref_idx": resp.RefIdx(), "fallback_plain_text": fallback,
	})
}

func (r *Registry) handleRecallGroupMessage(ctx context.Context, payload json.RawMessage) (string, error) {
	var args struct {
		GroupOpenID string `json:"group_openid"`
		MessageID   string `json:"message_id"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	gid, mid := strings.TrimSpace(args.GroupOpenID), strings.TrimSpace(args.MessageID)
	if gid == "" {
		return "", errors.New("group_openid is required")
	}
	if r.api == nil {
		return "", errors.New("qq api client not initialized")
	}

	var targetTimestamp int64
	if mid == "" {
		latest, ok := r.store.LatestRecallableBotMessage(gid)
		if !ok || latest.MessageID == "" {
			return "", errors.New("no recent bot message found to recall in this group")
		}
		mid, targetTimestamp = latest.MessageID, latest.Timestamp
	} else {
		cached, found := r.store.FindMessage(gid, mid)
		if !found {
			return "", fmt.Errorf("cannot recall message %s: not found in local cache so bot ownership cannot be verified", mid)
		}
		if !cached.IsSelf {
			return "", errors.New("cannot recall another user's message: bot is not a group admin")
		}
		if cached.Recalled {
			return "", fmt.Errorf("message %s has already been recalled", mid)
		}
		targetTimestamp = cached.Timestamp
	}

	if targetTimestamp > 0 && time.Since(time.Unix(targetTimestamp, 0)) > recallWindow {
		return "", fmt.Errorf("cannot recall message %s: exceeded 2-minute recall window", mid)
	}
	if err := r.api.RecallGroupMessage(ctx, gid, mid); err != nil {
		return "", fmt.Errorf("recall group message failed: %w", err)
	}
	r.store.MarkMessageRecalled(gid, mid)
	return marshalJSON(map[string]any{"recalled": true, "group_openid": gid, "message_id": mid})
}

func truncateRunes(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes]) + "..."
}
