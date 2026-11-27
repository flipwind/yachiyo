package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"yachiyo/yachiyo-util/logger"

	_ "github.com/mattn/go-sqlite3"
)

var storeLog = logger.New("Yachiyo.QQ.Store")

const (
	defaultMaxSizeMB     = 60
	pruneCheckInterval   = 64
	minPruneBatchSize    = 100
	pruneFractionPercent = 15
)

type Options struct {
	AppID              string // Bot AppID; each AppID gets its own isolated SQLite file when switching accounts.
	PrevAppID          string // Optional previous AppID when switching accounts (e.g. via -qr) to preserve legacy untagged DB ownership.
	DBPath             string // Empty or ":memory:" uses an in-memory SQLite DB.
	MaxSizeMB          int    // Maximum DB size in MB (default 60).
	MaxSizeBytes       int64  // Optional exact byte override (used in tests).
	RecentLimit        int    // Max messages returned by GetRecentMessages (default 100).
	DedupTTL           time.Duration
	PassiveMaxCount    int // Group passive reply limit per inbound message (default 5).
	C2CPassiveMaxCount int // C2C (private) passive reply limit per inbound message (default 4).
	PassiveTTL         time.Duration
}

type sessionMeta struct {
	Active          bool
	RejectProactive bool
	LastActiveAt    time.Time
	LastMsgID       string
	LastMsgExpireAt time.Time
	LastMsgSeq      int // Consumed passive reply quota count.
	LastMsgNextSeq  int // Monotonic next msg_seq cursor (persisted to prevent reuse across restarts).
	LastBotMsgID    string
}

func (m sessionMeta) CanPassiveReply(now time.Time, maxCount int) bool {
	return m.LastMsgID != "" && now.Before(m.LastMsgExpireAt) && m.LastMsgSeq < maxCount
}

type GroupState struct {
	GroupOpenID string
	Path        string
	Name        string
	sessionMeta
}

type PrivateState struct {
	UserOpenID string
	Path       string
	Username   string
	sessionMeta
}

type sessionRow struct {
	id   string
	name string
	sessionMeta
}

type scopeSpec struct {
	scopeType string
	table     string
	idCol     string
	nameCol   string
}

var (
	groupScope   = scopeSpec{"group", "groups", "group_openid", "name"}
	privateScope = scopeSpec{"private", "privates", "user_openid", "username"}
)

func specFor(scopeType string) scopeSpec {
	if scopeType == "private" {
		return privateScope
	}
	return groupScope
}

type CachedMessage struct {
	MessageID    string `json:"message_id"`
	RefIdx       string `json:"ref_idx,omitempty"`
	MemberOpenID string `json:"member_openid"`
	Author       string `json:"author"`
	Content      string `json:"content"`
	Timestamp    int64  `json:"timestamp"`
	IsSelf       bool   `json:"is_self"`
	Recalled     bool   `json:"recalled,omitempty"`
}

type CachedMember struct {
	MemberOpenID string    `json:"member_openid"`
	Username     string    `json:"username"`
	Role         string    `json:"member_role,omitempty"`
	JoinedAt     string    `json:"joined_at,omitempty"`
	UpdatedAt    time.Time `json:"-"`
}

type passiveAnchor struct {
	msgID        string
	expireAt     time.Time
	nextSeq      int
	used         int
	releasedSeqs []int
}

type Store struct {
	opts         Options
	dbPath       string
	maxSizeBytes int64
	db           *sql.DB

	mu            sync.RWMutex
	sessions      map[string]*sessionRow // key: "group/<id>" or "private/<id>"
	members       map[string]map[string]CachedMember
	anchors       map[string]*passiveAnchor
	cachedSeen    map[string]time.Time
	triggeredSeen map[string]time.Time
	writeCount    int
}

func New(opts Options) *Store {
	s, err := Open(opts)
	if err != nil {
		panic(fmt.Sprintf("open qq store: %v", err))
	}
	return s
}

func Open(opts Options) (*Store, error) {
	if opts.RecentLimit <= 0 {
		opts.RecentLimit = 100
	}
	if opts.DedupTTL <= 0 {
		opts.DedupTTL = 120 * time.Second
	}
	if opts.PassiveMaxCount <= 0 {
		opts.PassiveMaxCount = 5
	}
	if opts.C2CPassiveMaxCount <= 0 {
		opts.C2CPassiveMaxCount = 4
	}
	if opts.PassiveTTL <= 0 {
		opts.PassiveTTL = 3500 * time.Second
	}
	if opts.MaxSizeMB <= 0 {
		opts.MaxSizeMB = defaultMaxSizeMB
	}
	maxBytes := opts.MaxSizeBytes
	if maxBytes <= 0 {
		maxBytes = int64(opts.MaxSizeMB) * 1024 * 1024
	}

	dsn := strings.TrimSpace(opts.DBPath)
	appID, prevAppID := strings.TrimSpace(opts.AppID), strings.TrimSpace(opts.PrevAppID)
	if dsn == "" {
		dsn = ":memory:"
	} else if dsn != ":memory:" {
		var err error
		if dsn, err = resolveAppDBPath(dsn, appID, prevAppID); err != nil {
			return nil, err
		}
		if dir := filepath.Dir(dsn); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create db dir: %w", err)
			}
		}
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", dsn, err)
	}
	db.SetMaxOpenConns(1)

	if err := initSchema(db, dsn != ":memory:"); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := ensureAppIDMeta(db, appID); err != nil {
		_ = db.Close()
		return nil, err
	}

	s := &Store{
		opts:          opts,
		dbPath:        dsn,
		maxSizeBytes:  maxBytes,
		db:            db,
		sessions:      make(map[string]*sessionRow),
		members:       make(map[string]map[string]CachedMember),
		anchors:       make(map[string]*passiveAnchor),
		cachedSeen:    make(map[string]time.Time),
		triggeredSeen: make(map[string]time.Time),
	}
	if err := s.loadStateFromDB(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) DBPath() string         { return s.dbPath }
func (s *Store) PassiveMaxCount() int   { return s.opts.PassiveMaxCount }
func (s *Store) C2CPassiveMaxCount() int { return s.opts.C2CPassiveMaxCount }

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec(`PRAGMA wal_checkpoint(TRUNCATE);`)
	return s.db.Close()
}

func initSchema(db *sql.DB, isFile bool) error {
	pragmas := []string{
		`PRAGMA auto_vacuum = INCREMENTAL;`,
		`PRAGMA synchronous = NORMAL;`,
		`PRAGMA busy_timeout = 5000;`,
	}
	if isFile {
		pragmas = append(pragmas, `PRAGMA journal_mode = WAL;`)
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("exec %q: %w", p, err)
		}
	}

	const schema = `
	CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL DEFAULT '');
	CREATE TABLE IF NOT EXISTS groups (
		group_openid TEXT PRIMARY KEY, name TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1, reject_proactive INTEGER NOT NULL DEFAULT 0,
		last_active_at INTEGER NOT NULL DEFAULT 0, last_msg_id TEXT NOT NULL DEFAULT '',
		last_msg_expire_at INTEGER NOT NULL DEFAULT 0, last_msg_seq INTEGER NOT NULL DEFAULT 0,
		last_msg_next_seq INTEGER NOT NULL DEFAULT 0, last_bot_msg_id TEXT NOT NULL DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS privates (
		user_openid TEXT PRIMARY KEY, username TEXT NOT NULL DEFAULT '',
		active INTEGER NOT NULL DEFAULT 1, reject_proactive INTEGER NOT NULL DEFAULT 0,
		last_active_at INTEGER NOT NULL DEFAULT 0, last_msg_id TEXT NOT NULL DEFAULT '',
		last_msg_expire_at INTEGER NOT NULL DEFAULT 0, last_msg_seq INTEGER NOT NULL DEFAULT 0,
		last_msg_next_seq INTEGER NOT NULL DEFAULT 0, last_bot_msg_id TEXT NOT NULL DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS members (
		group_openid TEXT NOT NULL, member_openid TEXT NOT NULL,
		username TEXT NOT NULL DEFAULT '', role TEXT NOT NULL DEFAULT '',
		joined_at TEXT NOT NULL DEFAULT '', updated_at INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (group_openid, member_openid)
	);
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT, scope_type TEXT NOT NULL, scope_id TEXT NOT NULL,
		message_id TEXT NOT NULL DEFAULT '', ref_idx TEXT NOT NULL DEFAULT '',
		member_openid TEXT NOT NULL DEFAULT '', author TEXT NOT NULL DEFAULT '',
		content TEXT NOT NULL DEFAULT '', timestamp INTEGER NOT NULL DEFAULT 0,
		is_self INTEGER NOT NULL DEFAULT 0, recalled INTEGER NOT NULL DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_messages_scope_id ON messages(scope_type, scope_id, id DESC);
	CREATE INDEX IF NOT EXISTS idx_messages_scope_msgid ON messages(scope_type, scope_id, message_id);`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("init sqlite schema: %w", err)
	}
	for _, alter := range []string{
		`ALTER TABLE messages ADD COLUMN ref_idx TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE messages ADD COLUMN recalled INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE groups ADD COLUMN last_msg_next_seq INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE privates ADD COLUMN last_msg_next_seq INTEGER NOT NULL DEFAULT 0`,
	} {
		_, _ = db.Exec(alter)
	}
	return nil
}

// resolveAppDBPath keeps each AppID's SQLite cache isolated without wiping prior accounts.
func resolveAppDBPath(basePath, appID, prevAppID string) (string, error) {
	if appID == "" {
		return basePath, nil
	}
	if _, err := os.Stat(basePath); err != nil {
		if os.IsNotExist(err) {
			return basePath, nil
		}
		return "", fmt.Errorf("stat db %q: %w", basePath, err)
	}

	db, err := sql.Open("sqlite3", basePath)
	if err != nil {
		return "", fmt.Errorf("inspect sqlite %q: %w", basePath, err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	var storedAppID string
	_ = db.QueryRow(`SELECT value FROM meta WHERE key = 'app_id'`).Scan(&storedAppID)
	storedAppID = strings.TrimSpace(storedAppID)
	if storedAppID == appID {
		return basePath, nil
	}

	if storedAppID == "" {
		hasData := false
		for _, tbl := range []string{"groups", "privates", "messages"} {
			var cnt int64
			if db.QueryRow(`SELECT COUNT(*) FROM `+tbl).Scan(&cnt) == nil && cnt > 0 {
				hasData = true
				break
			}
		}
		if hasData && prevAppID != "" && prevAppID != appID {
			_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS meta (key TEXT PRIMARY KEY, value TEXT NOT NULL DEFAULT '')`)
			_, _ = db.Exec(`INSERT INTO meta (key, value) VALUES ('app_id', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, prevAppID)
			iso := scopedDBPath(basePath, appID)
			storeLog.Warn("Legacy cache DB %s attributed to previous AppID %s; using isolated DB %s for AppID %s", basePath, prevAppID, iso, appID)
			return iso, nil
		}
		return basePath, nil
	}

	iso := scopedDBPath(basePath, appID)
	storeLog.Warn("Cache DB %s belongs to AppID %s; preserving it and using isolated DB %s for AppID %s", basePath, storedAppID, iso, appID)
	return iso, nil
}

func scopedDBPath(basePath, appID string) string {
	ext := filepath.Ext(basePath)
	stem := strings.TrimSuffix(basePath, ext)
	if ext == "" {
		ext = ".db"
	}
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, appID)
	return fmt.Sprintf("%s.%s%s", stem, safe, ext)
}

func ensureAppIDMeta(db *sql.DB, appID string) error {
	if appID == "" {
		return nil
	}
	var stored string
	if err := db.QueryRow(`SELECT value FROM meta WHERE key = 'app_id'`).Scan(&stored); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("read stored app_id: %w", err)
	}
	if stored == "" {
		if _, err := db.Exec(`INSERT INTO meta (key, value) VALUES ('app_id', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, appID); err != nil {
			return fmt.Errorf("persist app_id: %w", err)
		}
	}
	return nil
}

func (s *Store) loadStateFromDB() error {
	for _, spec := range []scopeSpec{groupScope, privateScope} {
		q := fmt.Sprintf(`SELECT %s, %s, active, reject_proactive, last_active_at, last_msg_id, last_msg_expire_at, last_msg_seq, last_msg_next_seq, last_bot_msg_id FROM %s`,
			spec.idCol, spec.nameCol, spec.table)
		rows, err := s.db.Query(q)
		if err != nil {
			return fmt.Errorf("load %s: %w", spec.table, err)
		}
		for rows.Next() {
			var (
				r                      sessionRow
				active, reject         int
				lastActive, lastMsgExp int64
			)
			if err := rows.Scan(&r.id, &r.name, &active, &reject, &lastActive, &r.LastMsgID, &lastMsgExp, &r.LastMsgSeq, &r.LastMsgNextSeq, &r.LastBotMsgID); err != nil {
				_ = rows.Close()
				return err
			}
			if r.LastMsgNextSeq <= r.LastMsgSeq {
				r.LastMsgNextSeq = r.LastMsgSeq + 1
			}
			r.Active, r.RejectProactive = active != 0, reject != 0
			if lastActive > 0 {
				r.LastActiveAt = time.Unix(lastActive, 0)
			}
			if lastMsgExp > 0 {
				r.LastMsgExpireAt = time.Unix(lastMsgExp, 0)
			}
			rc := r
			s.sessions[spec.scopeType+"/"+r.id] = &rc
		}
		_ = rows.Close()
	}

	rows, err := s.db.Query(`SELECT group_openid, member_openid, username, role, joined_at, updated_at FROM members`)
	if err != nil {
		return fmt.Errorf("load members: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			groupID   string
			m         CachedMember
			updatedAt int64
		)
		if err := rows.Scan(&groupID, &m.MemberOpenID, &m.Username, &m.Role, &m.JoinedAt, &updatedAt); err != nil {
			return err
		}
		if updatedAt > 0 {
			m.UpdatedAt = time.Unix(updatedAt, 0)
		}
		if s.members[groupID] == nil {
			s.members[groupID] = make(map[string]CachedMember)
		}
		s.members[groupID][m.MemberOpenID] = m
	}
	return nil
}

// MarkSeen is an alias for MarkCached for backwards compatibility.
func (s *Store) MarkSeen(msgID string) bool      { return s.markDedup(msgID, true) }
func (s *Store) MarkCached(msgID string) bool    { return s.markDedup(msgID, true) }
func (s *Store) MarkTriggered(msgID string) bool { return s.markDedup(msgID, false) }

func (s *Store) markDedup(msgID string, cached bool) bool {
	if msgID == "" {
		return false
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()

	m := s.triggeredSeen
	if cached {
		m = s.cachedSeen
	}
	if exp, ok := m[msgID]; ok && now.Before(exp) {
		return true
	}
	m[msgID] = now.Add(s.opts.DedupTTL)
	if len(m) > 1024 {
		for k, exp := range m {
			if !now.Before(exp) {
				delete(m, k)
			}
		}
	}
	return false
}

func (s *Store) ensureSessionLocked(spec scopeSpec, id string) *sessionRow {
	key := spec.scopeType + "/" + id
	if r, ok := s.sessions[key]; ok {
		return r
	}
	r := &sessionRow{id: id, sessionMeta: sessionMeta{Active: true}}
	s.sessions[key] = r
	return r
}

func (s *Store) persistSessionLocked(spec scopeSpec, r *sessionRow) {
	q := fmt.Sprintf(`
		INSERT INTO %s (%s, %s, active, reject_proactive, last_active_at, last_msg_id, last_msg_expire_at, last_msg_seq, last_msg_next_seq, last_bot_msg_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(%s) DO UPDATE SET
			%s = excluded.%s, active = excluded.active, reject_proactive = excluded.reject_proactive,
			last_active_at = excluded.last_active_at, last_msg_id = excluded.last_msg_id,
			last_msg_expire_at = excluded.last_msg_expire_at, last_msg_seq = excluded.last_msg_seq,
			last_msg_next_seq = excluded.last_msg_next_seq, last_bot_msg_id = excluded.last_bot_msg_id`,
		spec.table, spec.idCol, spec.nameCol, spec.idCol, spec.nameCol, spec.nameCol)
	if _, err := s.db.Exec(q, r.id, r.name, boolToInt(r.Active), boolToInt(r.RejectProactive),
		unixOrZero(r.LastActiveAt), r.LastMsgID, unixOrZero(r.LastMsgExpireAt), r.LastMsgSeq, r.LastMsgNextSeq, r.LastBotMsgID); err != nil {
		storeLog.Warn("Persist %s %s failed: %v", spec.scopeType, r.id, err)
	}
}

func (s *Store) insertMessageLocked(scopeType, scopeID string, msg CachedMessage) {
	if _, err := s.db.Exec(`
		INSERT INTO messages (scope_type, scope_id, message_id, ref_idx, member_openid, author, content, timestamp, is_self, recalled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		scopeType, scopeID, msg.MessageID, msg.RefIdx, msg.MemberOpenID, msg.Author, msg.Content, msg.Timestamp, boolToInt(msg.IsSelf), boolToInt(msg.Recalled),
	); err != nil {
		storeLog.Warn("Insert message (%s/%s) failed: %v", scopeType, scopeID, err)
		return
	}
	s.writeCount++
	if s.writeCount%pruneCheckInterval == 0 {
		s.enforceSizeLimitLocked()
	}
}

func (s *Store) EnforceSizeLimit() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enforceSizeLimitLocked()
}

func (s *Store) DBSizeBytes() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dbSizeBytesLocked()
}

func (s *Store) dbSizeBytesLocked() int64 {
	var pageCount, pageSize int64
	_ = s.db.QueryRow(`PRAGMA page_count`).Scan(&pageCount)
	_ = s.db.QueryRow(`PRAGMA page_size`).Scan(&pageSize)
	return pageCount * pageSize
}

func (s *Store) enforceSizeLimitLocked() int64 {
	if s.maxSizeBytes <= 0 {
		return 0
	}
	curBytes := s.dbSizeBytesLocked()
	if curBytes <= s.maxSizeBytes {
		return 0
	}
	var totalRows int64
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&totalRows); err != nil || totalRows == 0 {
		return 0
	}
	toDelete := max(totalRows*pruneFractionPercent/100, 1)
	if toDelete < minPruneBatchSize && totalRows > minPruneBatchSize {
		toDelete = minPruneBatchSize
	}
	if toDelete >= totalRows && totalRows > 1 {
		toDelete = totalRows - 1
	}
	res, err := s.db.Exec(`DELETE FROM messages WHERE id IN (SELECT id FROM messages ORDER BY id ASC LIMIT ?)`, toDelete)
	if err != nil {
		storeLog.Warn("Prune old messages failed: %v", err)
		return 0
	}
	affected, _ := res.RowsAffected()
	_, _ = s.db.Exec(`PRAGMA incremental_vacuum`)
	storeLog.Info("Cache exceeded %.2fMB limit (was %.2fMB), pruned %d oldest messages (now %.2fMB)",
		float64(s.maxSizeBytes)/(1024*1024), float64(curBytes)/(1024*1024), affected, float64(s.dbSizeBytesLocked())/(1024*1024))
	return affected
}

func (s *Store) anchorKey(scopeType, scopeID, msgID string) string {
	return scopeType + ":" + scopeID + ":" + msgID
}

func (s *Store) trackAnchorLocked(scopeType, scopeID, msgID string, expireAt time.Time, used int) {
	if msgID == "" {
		return
	}
	now := time.Now()
	s.anchors[s.anchorKey(scopeType, scopeID, msgID)] = &passiveAnchor{
		msgID: msgID, expireAt: expireAt, nextSeq: used + 1, used: used,
	}
	if len(s.anchors) > 512 {
		for k, a := range s.anchors {
			if !now.Before(a.expireAt) {
				delete(s.anchors, k)
			}
		}
	}
}

func (s *Store) recordInbound(spec scopeSpec, scopeID, name string, msg CachedMessage, member *CachedMember) {
	if scopeID == "" {
		return
	}
	now := time.Now()
	eventTime := now
	if msg.Timestamp > 0 {
		eventTime = time.Unix(msg.Timestamp, 0)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	r := s.ensureSessionLocked(spec, scopeID)
	r.Active = true
	r.LastActiveAt = now
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		r.name = trimmed
	}
	if msg.MessageID != "" {
		expireAt := eventTime.Add(s.opts.PassiveTTL)
		s.trackAnchorLocked(spec.scopeType, scopeID, msg.MessageID, expireAt, 0)
		if r.LastMsgExpireAt.IsZero() || !expireAt.Before(r.LastMsgExpireAt) {
			r.LastMsgID, r.LastMsgExpireAt, r.LastMsgSeq, r.LastMsgNextSeq = msg.MessageID, expireAt, 0, 1
		}
	}
	s.persistSessionLocked(spec, r)
	s.insertMessageLocked(spec.scopeType, scopeID, msg)
	if member != nil && member.MemberOpenID != "" {
		s.upsertMemberLocked(scopeID, *member, now)
	}
}

func (s *Store) recordBot(spec scopeSpec, scopeID, msgID, refIdx, content string, ts int64) {
	if scopeID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	r := s.ensureSessionLocked(spec, scopeID)
	r.LastActiveAt = time.Now()
	if msgID != "" {
		r.LastBotMsgID = msgID
	}
	s.persistSessionLocked(spec, r)
	s.insertMessageLocked(spec.scopeType, scopeID, CachedMessage{
		MessageID: msgID, RefIdx: refIdx, MemberOpenID: "bot", Author: "bot",
		Content: content, Timestamp: ts, IsSelf: true,
	})
}

func (s *Store) RecordInboundMessage(groupOpenID string, msg CachedMessage, member *CachedMember) {
	s.recordInbound(groupScope, groupOpenID, "", msg, member)
}

func (s *Store) RecordBotMessage(groupOpenID, msgID, content string, ts int64) {
	s.recordBot(groupScope, groupOpenID, msgID, "", content, ts)
}

func (s *Store) RecordBotMessageWithRef(groupOpenID, msgID, refIdx, content string, ts int64) {
	s.recordBot(groupScope, groupOpenID, msgID, refIdx, content, ts)
}

func (s *Store) RecordInboundC2C(userOpenID, msgID, username, author, content string, ts int64) {
	s.RecordInboundC2CWithRef(userOpenID, msgID, "", username, author, content, ts)
}

func (s *Store) RecordInboundC2CWithRef(userOpenID, msgID, refIdx, username, author, content string, ts int64) {
	s.recordInbound(privateScope, userOpenID, username, CachedMessage{
		MessageID: msgID, RefIdx: refIdx, MemberOpenID: userOpenID,
		Author: author, Content: content, Timestamp: ts, IsSelf: false,
	}, nil)
}

func (s *Store) RecordBotC2C(userOpenID, msgID, content string, ts int64) {
	s.recordBot(privateScope, userOpenID, msgID, "", content, ts)
}

func (s *Store) RecordBotC2CWithRef(userOpenID, msgID, refIdx, content string, ts int64) {
	s.recordBot(privateScope, userOpenID, msgID, refIdx, content, ts)
}

func (s *Store) maxPassiveFor(scopeType string) int {
	if scopeType == "private" {
		return s.opts.C2CPassiveMaxCount
	}
	return s.opts.PassiveMaxCount
}

func (s *Store) LatestPassiveAnchorID(scopeType, scopeID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.sessions[specFor(scopeType).scopeType+"/"+scopeID]
	if !ok || !r.CanPassiveReply(time.Now(), s.maxPassiveFor(scopeType)) {
		return ""
	}
	return r.LastMsgID
}

// ReservePassiveAnchor atomically allocates the next passive (msgID, seq) slot under lock.
func (s *Store) ReservePassiveAnchor(scopeType, scopeID, preferredMsgID string) (string, int, bool) {
	spec := specFor(scopeType)
	now, maxCount := time.Now(), s.maxPassiveFor(scopeType)

	s.mu.Lock()
	defer s.mu.Unlock()

	if preferredMsgID != "" {
		if a, ok := s.anchors[s.anchorKey(spec.scopeType, scopeID, preferredMsgID)]; ok {
			return s.allocateFromAnchorLocked(spec, scopeID, a, now, maxCount)
		}
	}
	r, ok := s.sessions[spec.scopeType+"/"+scopeID]
	if !ok || r.LastMsgID == "" || (preferredMsgID != "" && r.LastMsgID != preferredMsgID) {
		return "", 0, false
	}
	nextSeq := max(r.LastMsgNextSeq, r.LastMsgSeq+1)
	key := s.anchorKey(spec.scopeType, scopeID, r.LastMsgID)
	a, exists := s.anchors[key]
	if !exists {
		a = &passiveAnchor{msgID: r.LastMsgID, expireAt: r.LastMsgExpireAt, nextSeq: nextSeq, used: r.LastMsgSeq}
		s.anchors[key] = a
	}
	return s.allocateFromAnchorLocked(spec, scopeID, a, now, maxCount)
}

func (s *Store) allocateFromAnchorLocked(spec scopeSpec, scopeID string, a *passiveAnchor, now time.Time, maxCount int) (string, int, bool) {
	if !now.Before(a.expireAt) || a.used >= maxCount {
		return "", 0, false
	}
	var seq int
	if len(a.releasedSeqs) > 0 {
		seq, a.releasedSeqs = a.releasedSeqs[0], a.releasedSeqs[1:]
	} else {
		seq = a.nextSeq
		a.nextSeq++
	}
	a.used++
	s.syncPersistedSeqLocked(spec, scopeID, a.msgID, a.used, a.nextSeq)
	return a.msgID, seq, true
}

// ReleasePassiveAnchor returns a reserved slot when a known 4xx rejection occurs.
func (s *Store) ReleasePassiveAnchor(scopeType, scopeID, msgID string, seq int) {
	if scopeID == "" || msgID == "" || seq <= 0 {
		return
	}
	spec := specFor(scopeType)
	s.mu.Lock()
	defer s.mu.Unlock()

	a, ok := s.anchors[s.anchorKey(spec.scopeType, scopeID, msgID)]
	if !ok || a.used <= 0 {
		return
	}
	a.used--
	if seq == a.nextSeq-1 {
		a.nextSeq--
	} else {
		a.releasedSeqs = append(a.releasedSeqs, seq)
	}
	s.syncPersistedSeqLocked(spec, scopeID, msgID, a.used, a.nextSeq)
}

func (s *Store) syncPersistedSeqLocked(spec scopeSpec, scopeID, msgID string, used, nextSeq int) {
	if r, ok := s.sessions[spec.scopeType+"/"+scopeID]; ok && r.LastMsgID == msgID {
		r.LastMsgSeq, r.LastMsgNextSeq = used, nextSeq
		s.persistSessionLocked(spec, r)
	}
}

func (s *Store) PeekPassiveAnchor(scopeType, scopeID, preferredMsgID string) (string, int, bool) {
	spec := specFor(scopeType)
	now, maxCount := time.Now(), s.maxPassiveFor(scopeType)

	s.mu.RLock()
	defer s.mu.RUnlock()

	if preferredMsgID != "" {
		if a, ok := s.anchors[s.anchorKey(spec.scopeType, scopeID, preferredMsgID)]; ok {
			if !now.Before(a.expireAt) || a.used >= maxCount {
				return "", 0, false
			}
			if len(a.releasedSeqs) > 0 {
				return a.msgID, a.releasedSeqs[0], true
			}
			return a.msgID, a.nextSeq, true
		}
	}
	r, ok := s.sessions[spec.scopeType+"/"+scopeID]
	if !ok || !r.CanPassiveReply(now, maxCount) || (preferredMsgID != "" && r.LastMsgID != preferredMsgID) {
		return "", 0, false
	}
	return r.LastMsgID, max(r.LastMsgNextSeq, r.LastMsgSeq+1), true
}

func (s *Store) NextPassiveAnchor(groupOpenID string) (string, int, bool) {
	return s.ReservePassiveAnchor("group", groupOpenID, "")
}

func (s *Store) NextC2CPassiveAnchor(userOpenID string) (string, int, bool) {
	return s.ReservePassiveAnchor("private", userOpenID, "")
}

func (s *Store) mutateSession(spec scopeSpec, id string, fn func(*sessionRow)) {
	if id == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.ensureSessionLocked(spec, id)
	fn(r)
	s.persistSessionLocked(spec, r)
}

func (s *Store) SetGroupActive(id string, active bool) {
	s.mutateSession(groupScope, id, func(r *sessionRow) { r.Active, r.LastActiveAt = active, time.Now() })
}
func (s *Store) SetGroupRejectProactive(id string, reject bool) {
	s.mutateSession(groupScope, id, func(r *sessionRow) { r.RejectProactive = reject })
}
func (s *Store) UpdateGroupName(id, name string) {
	if trimmed := strings.TrimSpace(name); trimmed != "" {
		s.mutateSession(groupScope, id, func(r *sessionRow) { r.name = trimmed })
	}
}
func (s *Store) SetFriendActive(id string, active bool) {
	s.mutateSession(privateScope, id, func(r *sessionRow) { r.Active, r.LastActiveAt = active, time.Now() })
}
func (s *Store) SetC2CRejectProactive(id string, reject bool) {
	s.mutateSession(privateScope, id, func(r *sessionRow) { r.RejectProactive = reject })
}

func (s *Store) isRejectProactive(spec scopeSpec, id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.sessions[spec.scopeType+"/"+id]
	return ok && r.RejectProactive
}

func (s *Store) IsGroupRejectProactive(id string) bool { return s.isRejectProactive(groupScope, id) }
func (s *Store) IsC2CRejectProactive(id string) bool   { return s.isRejectProactive(privateScope, id) }

func (s *Store) GetGroup(id string) (GroupState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.sessions["group/"+id]
	if !ok {
		return GroupState{}, false
	}
	return GroupState{GroupOpenID: r.id, Path: "group/" + r.id, Name: r.name, sessionMeta: r.sessionMeta}, true
}

func (s *Store) GetPrivate(id string) (PrivateState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.sessions["private/"+id]
	if !ok {
		return PrivateState{}, false
	}
	return PrivateState{UserOpenID: r.id, Path: "private/" + r.id, Username: r.name, sessionMeta: r.sessionMeta}, true
}

func (s *Store) ListGroups(onlyActive bool) []GroupState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []GroupState
	for k, r := range s.sessions {
		if !strings.HasPrefix(k, "group/") || (onlyActive && !r.Active) {
			continue
		}
		res = append(res, GroupState{GroupOpenID: r.id, Path: k, Name: r.name, sessionMeta: r.sessionMeta})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].LastActiveAt.After(res[j].LastActiveAt) })
	return res
}

func (s *Store) ListPrivates(onlyActive bool) []PrivateState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var res []PrivateState
	for k, r := range s.sessions {
		if !strings.HasPrefix(k, "private/") || (onlyActive && !r.Active) {
			continue
		}
		res = append(res, PrivateState{UserOpenID: r.id, Path: k, Username: r.name, sessionMeta: r.sessionMeta})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].LastActiveAt.After(res[j].LastActiveAt) })
	return res
}

func (s *Store) UpdateMember(groupOpenID string, member CachedMember) {
	if groupOpenID == "" || member.MemberOpenID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.upsertMemberLocked(groupOpenID, member, time.Now())
}

func (s *Store) upsertMemberLocked(groupOpenID string, member CachedMember, now time.Time) {
	if s.members[groupOpenID] == nil {
		s.members[groupOpenID] = make(map[string]CachedMember)
	}
	ex := s.members[groupOpenID][member.MemberOpenID]
	if u := strings.TrimSpace(member.Username); u != "" {
		ex.Username = u
	}
	if member.Role != "" {
		ex.Role = member.Role
	}
	if member.JoinedAt != "" {
		ex.JoinedAt = member.JoinedAt
	}
	ex.MemberOpenID, ex.UpdatedAt = member.MemberOpenID, now
	s.members[groupOpenID][member.MemberOpenID] = ex
	if _, err := s.db.Exec(`
		INSERT INTO members (group_openid, member_openid, username, role, joined_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(group_openid, member_openid) DO UPDATE SET
			username = excluded.username, role = excluded.role, joined_at = excluded.joined_at, updated_at = excluded.updated_at`,
		groupOpenID, ex.MemberOpenID, ex.Username, ex.Role, ex.JoinedAt, unixOrZero(ex.UpdatedAt),
	); err != nil {
		storeLog.Warn("Persist member %s failed: %v", ex.MemberOpenID, err)
	}
}

func (s *Store) LookupMemberName(groupOpenID, memberOpenID string) (string, bool) {
	m, ok := s.GetMember(groupOpenID, memberOpenID)
	return m.Username, ok && m.Username != ""
}

func (s *Store) GetMember(groupOpenID, memberOpenID string) (CachedMember, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.members[groupOpenID][memberOpenID]
	return m, ok
}

func (s *Store) GetRecentMessages(groupOpenID string, limit int) []CachedMessage {
	return s.getRecentByScope("group", groupOpenID, limit)
}

func (s *Store) GetRecentPrivateMessages(userOpenID string, limit int) []CachedMessage {
	return s.getRecentByScope("private", userOpenID, limit)
}

func (s *Store) getRecentByScope(scopeType, scopeID string, limit int) []CachedMessage {
	if limit <= 0 {
		limit = 20
	}
	limit = min(limit, s.opts.RecentLimit)

	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`
		SELECT message_id, ref_idx, member_openid, author, content, timestamp, is_self, recalled
		FROM messages WHERE scope_type = ? AND scope_id = ? ORDER BY id DESC LIMIT ?`, scopeType, scopeID, limit)
	if err != nil {
		storeLog.Warn("Query recent messages (%s=%s) failed: %v", scopeType, scopeID, err)
		return []CachedMessage{}
	}
	defer rows.Close()

	desc := make([]CachedMessage, 0, limit)
	for rows.Next() {
		if m, err := scanMessage(rows); err == nil {
			desc = append(desc, m)
		}
	}
	for i, j := 0, len(desc)-1; i < j; i, j = i+1, j-1 {
		desc[i], desc[j] = desc[j], desc[i]
	}
	return desc
}

func (s *Store) FindMessage(groupOpenID, messageID string) (CachedMessage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, err := scanMessage(s.db.QueryRow(`
		SELECT message_id, ref_idx, member_openid, author, content, timestamp, is_self, recalled
		FROM messages WHERE scope_type = 'group' AND scope_id = ? AND message_id = ? ORDER BY id DESC LIMIT 1`, groupOpenID, messageID))
	return m, err == nil
}

func (s *Store) LatestRecallableBotMessage(groupOpenID string) (CachedMessage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, err := scanMessage(s.db.QueryRow(`
		SELECT message_id, ref_idx, member_openid, author, content, timestamp, is_self, recalled
		FROM messages WHERE scope_type = 'group' AND scope_id = ? AND is_self = 1 AND recalled = 0 AND message_id != ''
		ORDER BY id DESC LIMIT 1`, groupOpenID))
	return m, err == nil
}

func (s *Store) MarkMessageRecalled(groupOpenID, messageID string) {
	if groupOpenID == "" || messageID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.db.Exec(`UPDATE messages SET recalled = 1 WHERE scope_type = 'group' AND scope_id = ? AND message_id = ?`, groupOpenID, messageID); err != nil {
		storeLog.Warn("Mark message recalled (%s/%s) failed: %v", groupOpenID, messageID, err)
	}
	if r, ok := s.sessions["group/"+groupOpenID]; ok && r.LastBotMsgID == messageID {
		var prevID string
		_ = s.db.QueryRow(`
			SELECT message_id FROM messages
			WHERE scope_type = 'group' AND scope_id = ? AND is_self = 1 AND recalled = 0 AND message_id != ''
			ORDER BY id DESC LIMIT 1`, groupOpenID).Scan(&prevID)
		r.LastBotMsgID = prevID
		s.persistSessionLocked(groupScope, r)
	}
}

func (s *Store) CountMessages() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var n int64
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM messages`).Scan(&n)
	return n
}

func scanMessage(scanner interface{ Scan(...any) error }) (CachedMessage, error) {
	var (
		m                CachedMessage
		isSelf, recalled int
	)
	if err := scanner.Scan(&m.MessageID, &m.RefIdx, &m.MemberOpenID, &m.Author, &m.Content, &m.Timestamp, &isSelf, &recalled); err != nil {
		return CachedMessage{}, err
	}
	m.IsSelf, m.Recalled = isSelf != 0, recalled != 0
	return m, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func unixOrZero(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.Unix()
}
