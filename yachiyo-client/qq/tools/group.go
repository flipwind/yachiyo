package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/yws"
	"yachiyo/yachiyo-util/logger"
)

var toolLog = logger.New("Yachiyo.QQ.Tools")

func (r *Registry) registerGroupTools() {
	r.register(yws.ToolItem{
		Name:        "qq_list_groups",
		Description: "List known QQ groups this bot has joined or seen, including group_openid, sub-path ('group/<group_openid>'), and full address ('wsjson://<client_id>/group/<group_openid>') for commit_message.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"only_active":{"type":"boolean","description":"If true (default), exclude groups the bot has been removed from."}}}`),
	}, r.handleListGroups)

	r.register(yws.ToolItem{
		Name:        "qq_list_privates",
		Description: "List known QQ private (C2C) chats this bot has interacted with, including user_openid, sub-path ('private/<user_openid>'), and full address ('wsjson://<client_id>/private/<user_openid>') for commit_message.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"only_active":{"type":"boolean","description":"If true (default), exclude users who removed the bot as friend."}}}`),
	}, r.handleListPrivates)

	r.register(yws.ToolItem{
		Name:        "qq_get_group_info",
		Description: "Get basic information about a QQ group and the bot's status in it. Falls back to locally cached metadata if the platform API restricts access.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"group_openid":{"type":"string","description":"Target QQ group_openid (the part after 'group/' in path)."}},"required":["group_openid"]}`),
	}, r.handleGetGroupInfo)

	r.register(yws.ToolItem{
		Name:        "qq_get_member_info",
		Description: "Get information (username, member_role, joined_at) for a specific member in a QQ group by member_openid. Falls back to locally cached author info if the platform API restricts access.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"group_openid":{"type":"string","description":"Target QQ group_openid."},"member_openid":{"type":"string","description":"Target member's member_openid."}},"required":["group_openid","member_openid"]}`),
	}, r.handleGetMemberInfo)
}

func parseOnlyActive(payload json.RawMessage) (bool, error) {
	if len(payload) == 0 {
		return true, nil
	}
	var args struct {
		OnlyActive *bool `json:"only_active"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return false, fmt.Errorf("invalid arguments: %w", err)
	}
	if args.OnlyActive != nil {
		return *args.OnlyActive, nil
	}
	return true, nil
}

func (r *Registry) handleListGroups(_ context.Context, payload json.RawMessage) (string, error) {
	onlyActive, err := parseOnlyActive(payload)
	if err != nil {
		return "", err
	}
	now, maxPassive := time.Now(), r.store.PassiveMaxCount()
	groups := r.store.ListGroups(onlyActive)
	items := make([]map[string]any, 0, len(groups))
	for _, g := range groups {
		items = append(items, map[string]any{
			"group_openid":      g.GroupOpenID,
			"path":              g.Path,
			"address":           fmt.Sprintf("wsjson://%s/%s", r.clientID, g.Path),
			"name":              g.Name,
			"active":            g.Active,
			"last_active_at":    g.LastActiveAt.Unix(),
			"can_passive_reply": g.CanPassiveReply(now, maxPassive),
			"reject_proactive":  g.RejectProactive,
		})
	}
	return marshalJSON(map[string]any{"groups": items})
}

func (r *Registry) handleListPrivates(_ context.Context, payload json.RawMessage) (string, error) {
	onlyActive, err := parseOnlyActive(payload)
	if err != nil {
		return "", err
	}
	now, maxPassive := time.Now(), r.store.C2CPassiveMaxCount()
	privates := r.store.ListPrivates(onlyActive)
	items := make([]map[string]any, 0, len(privates))
	for _, p := range privates {
		items = append(items, map[string]any{
			"user_openid":       p.UserOpenID,
			"path":              p.Path,
			"address":           fmt.Sprintf("wsjson://%s/%s", r.clientID, p.Path),
			"username":          p.Username,
			"active":            p.Active,
			"last_active_at":    p.LastActiveAt.Unix(),
			"can_passive_reply": p.CanPassiveReply(now, maxPassive),
			"reject_proactive":  p.RejectProactive,
		})
	}
	return marshalJSON(map[string]any{"privates": items})
}

func (r *Registry) handleGetGroupInfo(ctx context.Context, payload json.RawMessage) (string, error) {
	var args struct {
		GroupOpenID string `json:"group_openid"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	gid := strings.TrimSpace(args.GroupOpenID)
	if gid == "" {
		return "", errors.New("group_openid is required")
	}

	var apiErr error
	var apiErrMsg, apiErrType string
	if r.api != nil {
		if info, err := r.api.GetGroupInfo(ctx, gid); err == nil {
			if info.GroupName != "" {
				r.store.UpdateGroupName(gid, info.GroupName)
			}
			botState, botStateErr := r.api.GetGroupBotState(ctx, gid)
			res := map[string]any{
				"source":              "api",
				"group_info_source":   "api",
				"group_openid":        gid,
				"path":                "group/" + gid,
				"address":             fmt.Sprintf("wsjson://%s/group/%s", r.clientID, gid),
				"name":                info.GroupName,
				"info":                info,
				"bot_state_available": botStateErr == nil,
			}
			if botStateErr == nil {
				res["bot_state_source"], res["bot_state"] = "api", botState
			} else {
				res["source"], res["bot_state_source"] = "api_partial", "unavailable"
				res["bot_state_error_type"], res["bot_state_error"] = classifyAPIError(botStateErr), botStateErr.Error()
				toolLog.Warn("qq_get_group_info bot_state failed for %s: %v", gid, botStateErr)
			}
			return marshalJSON(res)
		} else {
			apiErr, apiErrMsg, apiErrType = err, err.Error(), classifyAPIError(err)
			toolLog.Warn("qq_get_group_info API failed for %s (%s: %v), falling back to local cache", gid, apiErrType, err)
		}
	}

	if cached, ok := r.store.GetGroup(gid); ok {
		return marshalJSON(map[string]any{
			"source": "cache_fallback", "group_info_source": "cache",
			"bot_state_available": false, "bot_state_source": "unavailable",
			"api_error_type": apiErrType, "api_error": apiErrMsg,
			"group_openid": cached.GroupOpenID, "path": cached.Path,
			"address": fmt.Sprintf("wsjson://%s/%s", r.clientID, cached.Path),
			"name": cached.Name, "active": cached.Active, "last_active_at": cached.LastActiveAt.Unix(),
			"can_passive_reply": cached.CanPassiveReply(time.Now(), r.store.PassiveMaxCount()),
			"reject_proactive":  cached.RejectProactive,
		})
	}
	if apiErr != nil {
		return "", fmt.Errorf("group %s lookup failed (%s, cache_miss): %w", gid, apiErrType, apiErr)
	}
	return "", fmt.Errorf("group %s not found via API or local cache (cache_miss)", gid)
}

func (r *Registry) handleGetMemberInfo(ctx context.Context, payload json.RawMessage) (string, error) {
	var args struct {
		GroupOpenID  string `json:"group_openid"`
		MemberOpenID string `json:"member_openid"`
	}
	if err := json.Unmarshal(payload, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	gid, mid := strings.TrimSpace(args.GroupOpenID), strings.TrimSpace(args.MemberOpenID)
	if gid == "" || mid == "" {
		return "", errors.New("group_openid and member_openid are required")
	}

	var apiErr error
	var apiErrMsg, apiErrType string
	if r.api != nil {
		if info, err := r.api.GetGroupMember(ctx, gid, mid); err == nil {
			r.store.UpdateMember(gid, store.CachedMember{
				MemberOpenID: info.MemberOpenID, Username: info.Username, Role: info.MemberRole, JoinedAt: info.JoinedAt,
			})
			return marshalJSON(map[string]any{
				"source": "api", "available": true, "group_openid": gid, "member_openid": info.MemberOpenID,
				"username": info.Username, "member_role": info.MemberRole, "bot": info.Bot, "joined_at": info.JoinedAt,
			})
		} else {
			apiErr, apiErrMsg, apiErrType = err, err.Error(), classifyAPIError(err)
			toolLog.Warn("qq_get_member_info API failed for %s/%s (%s: %v), falling back to local cache", gid, mid, apiErrType, err)
		}
	}

	if cached, ok := r.store.GetMember(gid, mid); ok {
		return marshalJSON(map[string]any{
			"source": "cache_fallback", "available": true, "api_error_type": apiErrType, "api_error": apiErrMsg,
			"group_openid": gid, "member_openid": cached.MemberOpenID,
			"username": cached.Username, "member_role": cached.Role, "joined_at": cached.JoinedAt,
		})
	}
	if apiErr != nil {
		return "", fmt.Errorf("member %s in group %s lookup failed (%s, cache_miss): %w", mid, gid, apiErrType, apiErr)
	}
	return "", fmt.Errorf("member %s in group %s not found via API or local cache (cache_miss)", mid, gid)
}

func classifyAPIError(err error) string {
	var apiErr *qq.APIError
	if !errors.As(err, &apiErr) {
		return "network_error"
	}
	switch {
	case apiErr.IsPermissionDenied():
		return "permission_denied"
	case apiErr.IsAuthError():
		return "auth_error"
	case apiErr.IsRateLimit():
		return "rate_limited"
	default:
		return "api_error"
	}
}
