package bridge

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"yachiyo/yachiyo-client/qq/qq"
	"yachiyo/yachiyo-client/qq/store"
	"yachiyo/yachiyo-client/qq/yws"
	"yachiyo/yachiyo-util/logger"
)

const (
	maxPendingMessages  = 100
	pendingMessageTTL   = 10 * time.Minute
	maxSessionQueueSize = 1024
	logPreviewMaxRunes  = 80
	forwardWriteTimeout = 5 * time.Second
)

var (
	bridgeLog   = logger.New("Yachiyo.QQ.Bridge")
	reAtAll     = regexp.MustCompile(`<@all>`)
	reAtUserOld = regexp.MustCompile(`<@!?([a-zA-Z0-9_-]+)>`)
	reAtUserNew = regexp.MustCompile(`<qqbot-at-user\s+id="([^"]+)"\s*/?>`)
)

type Sender interface {
	SendClientMessage(ctx context.Context, msg yws.ClientMessage) error
}

type connectionChecker interface {
	IsConnected() bool
}

type Options struct {
	GroupReplyMode string // "at_only" (default) or "all"
	LogContent     bool   // If true, include sanitized/truncated content previews in INFO logs.
}

type pendingMessage struct {
	msg        yws.ClientMessage
	enqueuedAt time.Time
}

type sessionWorker struct {
	key   string
	mu    sync.Mutex
	cond  *sync.Cond
	queue []func()
}

func newSessionWorker(key string) *sessionWorker {
	w := &sessionWorker{key: key}
	w.cond = sync.NewCond(&w.mu)
	go func() {
		for {
			w.mu.Lock()
			for len(w.queue) == 0 {
				w.cond.Wait()
			}
			fn := w.queue[0]
			w.queue[0] = nil
			w.queue = w.queue[1:]
			w.mu.Unlock()
			fn()
		}
	}()
	return w
}

func (w *sessionWorker) enqueue(fn func()) {
	w.mu.Lock()
	if len(w.queue) >= maxSessionQueueSize {
		bridgeLog.Warn("Session queue %s exceeded %d items, dropping oldest queued task", w.key, maxSessionQueueSize)
		w.queue[0] = nil
		w.queue = w.queue[1:]
	}
	w.queue = append(w.queue, fn)
	w.cond.Signal()
	w.mu.Unlock()
}

type Bridge struct {
	api            *qq.APIClient
	store          *store.Store
	sender         Sender
	sleepFn        func(ctx context.Context, d time.Duration) bool
	groupReplyMode string
	logContent     bool

	mu          sync.RWMutex
	botID       string
	botUsername string

	forwardMu sync.Mutex
	pending   []pendingMessage

	workersMu sync.Mutex
	workers   map[string]*sessionWorker
}

func New(api *qq.APIClient, st *store.Store, sender Sender) *Bridge {
	return NewWithOptions(api, st, sender, Options{})
}

func NewWithOptions(api *qq.APIClient, st *store.Store, sender Sender, opts Options) *Bridge {
	mode := strings.ToLower(strings.TrimSpace(opts.GroupReplyMode))
	if mode != "all" {
		mode = "at_only"
	}
	return &Bridge{
		api:            api,
		store:          st,
		sender:         sender,
		sleepFn:        defaultSleep,
		groupReplyMode: mode,
		logContent:     opts.LogContent,
		workers:        make(map[string]*sessionWorker),
	}
}

func defaultSleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// BubbleDelay matches Pancake's inter-bubble pacing: clamp(300ms, 5000ms, 200 + 100 * runes ms).
func BubbleDelay(msg string) time.Duration {
	ms := min(max(200+100*utf8.RuneCountInString(msg), 300), 5000)
	return time.Duration(ms) * time.Millisecond
}

func (b *Bridge) SetBotIdentity(id, username string) {
	b.mu.Lock()
	b.botID, b.botUsername = strings.TrimSpace(id), strings.TrimSpace(username)
	b.mu.Unlock()
}

func (b *Bridge) enqueueSession(key string, fn func()) {
	b.workersMu.Lock()
	w, ok := b.workers[key]
	if !ok {
		w = newSessionWorker(key)
		b.workers[key] = w
	}
	b.workersMu.Unlock()
	w.enqueue(fn)
}

func (b *Bridge) runInSession(ctx context.Context, key string, fn func()) {
	done := make(chan struct{})
	b.enqueueSession(key, func() {
		defer close(done)
		fn()
	})
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (b *Bridge) appendPendingLocked(msg yws.ClientMessage) {
	if len(b.pending) >= maxPendingMessages {
		b.pending = b.pending[1:]
	}
	b.pending = append(b.pending, pendingMessage{msg: msg, enqueuedAt: time.Now()})
}

// forwardClientMessage keeps pending and new messages strictly ordered, uses an independent
// write timeout so a cancelled QQ session ctx never closes wsjson, and retries immediately
// if a late write error came from an already-replaced stale connection.
func (b *Bridge) forwardClientMessage(msg yws.ClientMessage) (bool, error) {
	b.forwardMu.Lock()
	defer b.forwardMu.Unlock()

	if len(b.pending) > 0 {
		if _, err := b.flushPendingLocked(); err != nil || len(b.pending) > 0 {
			b.appendPendingLocked(msg)
			return false, err
		}
	}

	writeCtx, cancel := context.WithTimeout(context.Background(), forwardWriteTimeout)
	err := b.sender.SendClientMessage(writeCtx, msg)
	cancel()
	if err != nil {
		b.appendPendingLocked(msg)
		if checker, ok := b.sender.(connectionChecker); ok && checker.IsConnected() {
			if n, flushErr := b.flushPendingLocked(); flushErr == nil && len(b.pending) == 0 && n > 0 {
				return true, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (b *Bridge) FlushPending(_ context.Context) int {
	b.forwardMu.Lock()
	defer b.forwardMu.Unlock()
	delivered, _ := b.flushPendingLocked()
	return delivered
}

func (b *Bridge) flushPendingLocked() (int, error) {
	if len(b.pending) == 0 {
		return 0, nil
	}
	batch := b.pending
	b.pending = nil

	now := time.Now()
	delivered := 0
	var lastErr error
	for i, item := range batch {
		if now.Sub(item.enqueuedAt) > pendingMessageTTL {
			continue
		}
		writeCtx, cancel := context.WithTimeout(context.Background(), forwardWriteTimeout)
		err := b.sender.SendClientMessage(writeCtx, item.msg)
		cancel()
		if err != nil {
			b.pending = append([]pendingMessage(nil), batch[i:]...)
			lastErr = err
			bridgeLog.Warn("Flush pending inbound messages paused (%d delivered, %d remaining): %v", delivered, len(b.pending), err)
			break
		}
		delivered++
	}
	if delivered > 0 {
		bridgeLog.Info("Flushed %d buffered QQ inbound message(s) after gateway reconnect", delivered)
	}
	return delivered, lastErr
}

func firstNonEmpty(a, b string) string {
	if s := strings.TrimSpace(a); s != "" {
		return s
	}
	return strings.TrimSpace(b)
}

func (b *Bridge) EnqueueGroupMessage(_ context.Context, ev qq.GroupMessageEvent, isAtEvent bool) {
	b.enqueueSession("in:group/"+firstNonEmpty(ev.GroupOpenID, ev.GroupID), func() {
		b.handleGroupMessageLocked(ev, isAtEvent)
	})
}

func (b *Bridge) HandleGroupMessage(ctx context.Context, ev qq.GroupMessageEvent, isAtEvent bool) {
	b.runInSession(ctx, "in:group/"+firstNonEmpty(ev.GroupOpenID, ev.GroupID), func() {
		b.handleGroupMessageLocked(ev, isAtEvent)
	})
}

func (b *Bridge) EnqueueC2CMessage(_ context.Context, ev qq.C2CMessageEvent) {
	b.enqueueSession("in:private/"+firstNonEmpty(ev.Author.UserOpenID, ev.Author.ID), func() {
		b.handleC2CMessageLocked(ev)
	})
}

func (b *Bridge) HandleC2CMessage(ctx context.Context, ev qq.C2CMessageEvent) {
	b.runInSession(ctx, "in:private/"+firstNonEmpty(ev.Author.UserOpenID, ev.Author.ID), func() {
		b.handleC2CMessageLocked(ev)
	})
}

func (b *Bridge) EnqueueRuntimeMessage(ctx context.Context, msg yws.RuntimeMessage) {
	b.enqueueSession("out:"+msg.Path, func() {
		b.handleRuntimeMessageLocked(ctx, msg)
	})
}

func (b *Bridge) HandleRuntimeMessage(ctx context.Context, msg yws.RuntimeMessage) {
	b.runInSession(ctx, "out:"+msg.Path, func() {
		b.handleRuntimeMessageLocked(ctx, msg)
	})
}

func (b *Bridge) handleGroupMessageLocked(ev qq.GroupMessageEvent, isAtEvent bool) {
	groupOpenID := firstNonEmpty(ev.GroupOpenID, ev.GroupID)
	if groupOpenID == "" {
		bridgeLog.Warn("Ignored group message %s without group_openid", ev.ID)
		return
	}
	memberOpenID := firstNonEmpty(ev.Author.MemberOpenID, ev.Author.ID)
	username := strings.TrimSpace(ev.Author.Username)

	b.mu.RLock()
	botID := b.botID
	b.mu.RUnlock()
	isSelf := botID != "" && memberOpenID == botID

	alreadyCached := b.store.MarkCached(ev.ID)
	mentionNames := make(map[string]string, len(ev.Mentions))
	for _, m := range ev.Mentions {
		id := firstNonEmpty(m.MemberOpenID, m.ID)
		if name := strings.TrimSpace(m.Username); id != "" && name != "" {
			mentionNames[id] = name
			if !alreadyCached {
				b.store.UpdateMember(groupOpenID, store.CachedMember{MemberOpenID: id, Username: name})
			}
		}
	}

	author := FormatAuthor(username, memberOpenID)
	content := b.formatInboundContent(groupOpenID, ev.Content, ev.Attachments, mentionNames, isAtEvent)
	ts, refIdx := parseTimestamp(ev.Timestamp), qq.ExtractMsgIdx(ev.MessageScene)

	if !alreadyCached {
		var member *store.CachedMember
		if memberOpenID != "" {
			member = &store.CachedMember{MemberOpenID: memberOpenID, Username: username, Role: ev.Author.MemberRole}
		}
		b.store.RecordInboundMessage(groupOpenID, store.CachedMessage{
			MessageID: ev.ID, RefIdx: refIdx, MemberOpenID: memberOpenID,
			Author: author, Content: content, Timestamp: ts, IsSelf: isSelf,
		}, member)
	}

	if strings.TrimSpace(content) == "" || ev.Author.Bot || isSelf {
		return
	}
	if b.groupReplyMode == "at_only" && !b.isMentioningBot(ev, isAtEvent) {
		bridgeLog.Debug("Cached non-@ group message [%s] msg_id=%s (at_only)", shortPath("group/"+groupOpenID), ev.ID)
		return
	}
	if b.store.MarkTriggered(ev.ID) {
		bridgeLog.Debug("Deduplicated group trigger %s", ev.ID)
		return
	}
	b.deliverInbound("group/"+groupOpenID, username, ev.ID, author, content, refIdx, ts)
}

func (b *Bridge) handleC2CMessageLocked(ev qq.C2CMessageEvent) {
	userOpenID := firstNonEmpty(ev.Author.UserOpenID, ev.Author.ID)
	if userOpenID == "" {
		bridgeLog.Warn("Ignored C2C message %s without user_openid", ev.ID)
		return
	}
	username := strings.TrimSpace(ev.Author.Username)
	author := FormatAuthor(username, userOpenID)
	content := appendAttachments(strings.TrimSpace(ev.Content), ev.Attachments)
	ts, refIdx := parseTimestamp(ev.Timestamp), qq.ExtractMsgIdx(ev.MessageScene)

	if !b.store.MarkCached(ev.ID) {
		b.store.RecordInboundC2CWithRef(userOpenID, ev.ID, refIdx, username, author, content, ts)
	}
	if content == "" || b.store.MarkTriggered(ev.ID) {
		return
	}
	b.deliverInbound("private/"+userOpenID, username, ev.ID, author, content, refIdx, ts)
}

func (b *Bridge) deliverInbound(path, username, msgID, author, content, refIdx string, ts int64) {
	if delivered, err := b.forwardClientMessage(yws.ClientMessage{
		Message: content, Time: ts, Author: author, Path: path,
	}); !delivered {
		bridgeLog.Warn("Forward QQ message failed (path=%s, msg_id=%s), buffered: %v", shortPath(path), msgID, err)
		return
	}
	name := firstNonEmpty(username, "user")
	if b.logContent {
		bridgeLog.Info("QQ -> Yachiyo [%s] author=%s msg_id=%s len=%d preview=%q",
			shortPath(path), name, shortID(msgID), utf8.RuneCountInString(content), sanitizePreview(content, logPreviewMaxRunes))
	} else {
		bridgeLog.Info("QQ -> Yachiyo [%s] author=%s msg_id=%s len=%d",
			shortPath(path), name, shortID(msgID), utf8.RuneCountInString(content))
	}
	bridgeLog.Debug("QQ -> Yachiyo full path=%s author=%q msg_id=%s ref_idx=%s", path, author, msgID, refIdx)
}

func (b *Bridge) isMentioningBot(ev qq.GroupMessageEvent, isAtEvent bool) bool {
	if isAtEvent {
		return true
	}
	b.mu.RLock()
	botID := b.botID
	b.mu.RUnlock()
	for _, m := range ev.Mentions {
		if m.IsYou || (botID != "" && (strings.TrimSpace(m.MemberOpenID) == botID || strings.TrimSpace(m.ID) == botID)) {
			return true
		}
	}
	return botID != "" && (strings.Contains(ev.Content, "<@"+botID+">") ||
		strings.Contains(ev.Content, "<@!"+botID+">") ||
		strings.Contains(ev.Content, `id="`+botID+`"`))
}

func (b *Bridge) handleRuntimeMessageLocked(ctx context.Context, msg yws.RuntimeMessage) {
	if len(msg.Message) == 0 {
		return
	}
	var isGroup bool
	var scopeType, targetID string
	switch {
	case strings.HasPrefix(msg.Path, "group/"):
		isGroup, scopeType, targetID = true, "group", strings.TrimPrefix(msg.Path, "group/")
	case strings.HasPrefix(msg.Path, "private/"):
		isGroup, scopeType, targetID = false, "private", strings.TrimPrefix(msg.Path, "private/")
	default:
		bridgeLog.Warn("Ignored runtime_message with unsupported path %q", msg.Path)
		return
	}
	if targetID == "" {
		return
	}

	bubbles := make([]string, 0, len(msg.Message))
	for _, raw := range msg.Message {
		if text := strings.TrimSpace(raw); text != "" {
			bubbles = append(bubbles, text)
		}
	}
	if len(bubbles) == 0 {
		return
	}

	sleep := b.sleepFn
	if sleep == nil {
		sleep = defaultSleep
	}
	turnAnchorID := ""
	if !msg.IsInitiative {
		turnAnchorID = b.store.LatestPassiveAnchorID(scopeType, targetID)
	}

	for idx, text := range bubbles {
		if idx != 0 && !sleep(ctx, BubbleDelay(text)) {
			return
		}
		var sendOpts qq.SendOptions
		if !msg.IsInitiative && turnAnchorID != "" {
			if msgID, seq, ok := b.store.ReservePassiveAnchor(scopeType, targetID, turnAnchorID); ok {
				sendOpts.MsgID, sendOpts.MsgSeq = msgID, seq
			}
		}
		if sendOpts.MsgID == "" && ((isGroup && b.store.IsGroupRejectProactive(targetID)) || (!isGroup && b.store.IsC2CRejectProactive(targetID))) {
			bridgeLog.Warn("Skipped proactive QQ send (path=%s, bubble=%d/%d): target rejected proactive messages",
				shortPath(msg.Path), idx+1, len(bubbles))
			return
		}

		start := time.Now()
		sendCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
		var resp qq.SendMessageResponse
		var fallback bool
		var err error
		if isGroup {
			resp, fallback, err = b.api.SendGroupMessage(sendCtx, targetID, text, sendOpts)
		} else {
			resp, fallback, err = b.api.SendC2CMessage(sendCtx, targetID, text, sendOpts)
		}
		cancel()
		latency := time.Since(start)

		if err != nil {
			if sendOpts.MsgID != "" && qq.ShouldReleasePassiveOnError(err) {
				b.store.ReleasePassiveAnchor(scopeType, targetID, sendOpts.MsgID, sendOpts.MsgSeq)
			}
			bridgeLog.Error("Send QQ message failed (path=%s, bubble=%d/%d, latency=%s): %v",
				shortPath(msg.Path), idx+1, len(bubbles), latency.Round(time.Millisecond), err)
			continue
		}

		b.logOutbound(msg.Path, idx+1, len(bubbles), resp.ID, fallback, latency, text)
		bridgeLog.Debug("Yachiyo -> QQ full path=%s id=%s ref_idx=%s passive_msg_id=%s seq=%d fallback=%t",
			msg.Path, resp.ID, resp.RefIdx(), sendOpts.MsgID, sendOpts.MsgSeq, fallback)

		if resp.ID != "" {
			if isGroup {
				b.store.RecordBotMessageWithRef(targetID, resp.ID, resp.RefIdx(), text, time.Now().Unix())
			} else {
				b.store.RecordBotC2CWithRef(targetID, resp.ID, resp.RefIdx(), text, time.Now().Unix())
			}
		}
	}
}

func (b *Bridge) logOutbound(path string, idx, total int, msgID string, fallback bool, latency time.Duration, content string) {
	mode := "md"
	if fallback {
		mode = "plain"
	}
	progress := ""
	if total > 1 {
		progress = fmt.Sprintf("(%d/%d) ", idx, total)
	}
	if b.logContent {
		bridgeLog.Info("Yachiyo -> QQ [%s] %smsg_id=%s mode=%s len=%d latency=%s preview=%q",
			shortPath(path), progress, shortID(msgID), mode, utf8.RuneCountInString(content), latency.Round(time.Millisecond), sanitizePreview(content, logPreviewMaxRunes))
		return
	}
	bridgeLog.Info("Yachiyo -> QQ [%s] %smsg_id=%s mode=%s len=%d latency=%s",
		shortPath(path), progress, shortID(msgID), mode, utf8.RuneCountInString(content), latency.Round(time.Millisecond))
}

func sanitizePreview(s string, maxRunes int) string {
	var sb strings.Builder
	sb.Grow(len(s))
	for count, r := 0, rune(0); len(s) > 0; {
		var sz int
		r, sz = utf8.DecodeRuneInString(s)
		s = s[sz:]
		if r == '\r' || r == '\n' || r == '\t' {
			r = ' '
		} else if r < 0x20 || r == 0x7f {
			continue
		}
		if count >= maxRunes {
			sb.WriteString("...")
			break
		}
		sb.WriteRune(r)
		count++
	}
	return sb.String()
}

func shortPath(path string) string {
	if idx := strings.IndexByte(path, '/'); idx >= 0 && idx < len(path)-1 {
		return path[:idx+1] + shortID(path[idx+1:])
	}
	return shortID(path)
}

func shortID(id string) string {
	if len(id) <= 10 {
		return id
	}
	return id[:4] + ".." + id[len(id)-4:]
}

func FormatAuthor(username, memberOpenID string) string {
	name := firstNonEmpty(username, "user")
	if id := strings.TrimSpace(memberOpenID); id != "" {
		return fmt.Sprintf("%s [openid:%s]", name, id)
	}
	return name
}

func (b *Bridge) formatInboundContent(groupOpenID, raw string, attachments []qq.Attachment, mentionNames map[string]string, isAtEvent bool) string {
	b.mu.RLock()
	botID, botName := b.botID, b.botUsername
	b.mu.RUnlock()

	resolve := func(id string) string {
		if id == "" {
			return "@user"
		}
		if botID != "" && id == botID {
			return "@" + firstNonEmpty(botName, "bot")
		}
		if name := mentionNames[id]; name != "" {
			return "@" + name
		}
		if name, ok := b.store.LookupMemberName(groupOpenID, id); ok {
			return "@" + name
		}
		return fmt.Sprintf("@user [openid:%s]", id)
	}

	text := reAtAll.ReplaceAllString(raw, "@全体成员")
	for _, re := range []*regexp.Regexp{reAtUserNew, reAtUserOld} {
		text = re.ReplaceAllStringFunc(text, func(match string) string {
			if sub := re.FindStringSubmatch(match); len(sub) >= 2 {
				return resolve(strings.TrimSpace(sub[1]))
			}
			return match
		})
	}
	text = strings.TrimSpace(text)
	if isAtEvent && botName != "" && !strings.Contains(text, "@"+botName) {
		text = "@" + botName + " " + text
	}
	return appendAttachments(text, attachments)
}

func appendAttachments(text string, attachments []qq.Attachment) string {
	if len(attachments) == 0 {
		return strings.TrimSpace(text)
	}
	parts := make([]string, 0, len(attachments))
	for _, att := range attachments {
		ct := strings.ToLower(att.ContentType)
		switch {
		case strings.HasPrefix(ct, "image/"):
			parts = append(parts, "[图片]")
		case strings.HasPrefix(ct, "audio/"):
			parts = append(parts, "[语音]")
		case strings.HasPrefix(ct, "video/"):
			parts = append(parts, "[视频]")
		case att.Filename != "":
			parts = append(parts, fmt.Sprintf("[文件: %s]", att.Filename))
		default:
			parts = append(parts, "[附件]")
		}
	}
	attText := strings.Join(parts, " ")
	if text == "" {
		return strings.TrimSpace(attText)
	}
	return strings.TrimSpace(text + "\n" + attText)
}

func parseTimestamp(raw string) int64 {
	if t, err := time.Parse(time.RFC3339, strings.TrimSpace(raw)); err == nil {
		return t.Unix()
	}
	return time.Now().Unix()
}
