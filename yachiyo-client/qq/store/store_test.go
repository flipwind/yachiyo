package store

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSQLitePersistenceAcrossReopen(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "qq_cache.db")

	s1, err := Open(Options{
		DBPath:      dbPath,
		MaxSizeMB:   60,
		RecentLimit: 100,
	})
	if err != nil {
		t.Fatalf("Open s1 failed: %v", err)
	}

	s1.RecordInboundMessage("GRP_PERSIST", CachedMessage{
		MessageID:    "MSG_1",
		MemberOpenID: "U1",
		Author:       "小红 [openid:U1]",
		Content:      "持久化测试",
		Timestamp:    1700000001,
	}, &CachedMember{
		MemberOpenID: "U1",
		Username:     "小红",
		Role:         "member",
	})
	s1.RecordBotMessage("GRP_PERSIST", "BOT_MSG_1", "收到啦", 1700000002)
	s1.RecordInboundC2C("USER_1", "C2C_1", "小蓝", "小蓝 [openid:USER_1]", "私聊测试", 1700000003)

	if err := s1.Close(); err != nil {
		t.Fatalf("Close s1 failed: %v", err)
	}

	// Reopen from disk and verify everything survived.
	s2, err := Open(Options{
		DBPath:      dbPath,
		MaxSizeMB:   60,
		RecentLimit: 100,
	})
	if err != nil {
		t.Fatalf("Open s2 failed: %v", err)
	}
	defer s2.Close()

	g, ok := s2.GetGroup("GRP_PERSIST")
	if !ok || g.LastBotMsgID != "BOT_MSG_1" || g.LastMsgID != "MSG_1" {
		t.Fatalf("group state not persisted properly: ok=%v g=%+v", ok, g)
	}
	name, ok := s2.LookupMemberName("GRP_PERSIST", "U1")
	if !ok || name != "小红" {
		t.Fatalf("member cache not persisted: ok=%v name=%q", ok, name)
	}
	msgs := s2.GetRecentMessages("GRP_PERSIST", 10)
	if len(msgs) != 2 || msgs[0].MessageID != "MSG_1" || msgs[1].MessageID != "BOT_MSG_1" || !msgs[1].IsSelf {
		t.Fatalf("unexpected persisted messages: %+v", msgs)
	}
	p, ok := s2.GetPrivate("USER_1")
	if !ok || p.LastMsgID != "C2C_1" || p.Username != "小蓝" {
		t.Fatalf("private state not persisted: ok=%v p=%+v", ok, p)
	}
}

func TestSQLiteSizeLimitPrunesOldestMessages(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "prune.db")
	st, err := Open(Options{
		DBPath:       dbPath,
		MaxSizeBytes: 64 * 1024, // 64KB cap for fast deterministic test
		RecentLimit:  100,
	})
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer st.Close()

	payload := strings.Repeat("A", 1024)
	for i := 0; i < 150; i++ {
		st.RecordInboundMessage("GRP_BIG", CachedMessage{
			MessageID:    fmt.Sprintf("M_%03d", i),
			MemberOpenID: "U1",
			Author:       "u [openid:U1]",
			Content:      payload,
			Timestamp:    int64(1700000000 + i),
		}, nil)
	}

	if st.DBSizeBytes() > 64*1024 {
		pruned := st.EnforceSizeLimit()
		if pruned == 0 {
			t.Fatalf("expected EnforceSizeLimit to prune oldest messages when DBSizeBytes=%d > 64KB", st.DBSizeBytes())
		}
	}

	remaining := st.CountMessages()
	if remaining >= 150 {
		t.Fatalf("expected total messages < 150 after pruning, got %d", remaining)
	}

	// Newest message must still be present.
	recent := st.GetRecentMessages("GRP_BIG", 1)
	if len(recent) != 1 || recent[0].MessageID != "M_149" {
		t.Fatalf("expected newest message M_149 to survive pruning, got %+v", recent)
	}
}


func TestAppIDSwitchIsolatesCache(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/qq_cache.db"

	st1, err := Open(Options{AppID: "APP_1", DBPath: dbPath})
	if err != nil {
		t.Fatalf("Open APP_1: %v", err)
	}
	st1.RecordInboundMessage("G1", CachedMessage{MessageID: "M1", Content: "hello", Timestamp: 1700000000}, &CachedMember{MemberOpenID: "U1", Username: "Alice"})
	st1.RecordInboundC2C("U1", "C1", "Alice", "Alice [openid:U1]", "dm", 1700000001)
	_ = st1.Close()

	// Reopening with the same AppID preserves all data.
	stSame, err := Open(Options{AppID: "APP_1", DBPath: dbPath})
	if err != nil {
		t.Fatalf("Open same APP_1: %v", err)
	}
	if len(stSame.ListGroups(false)) != 1 || len(stSame.ListPrivates(false)) != 1 || stSame.CountMessages() != 2 {
		t.Fatalf("same AppID should preserve data: groups=%d privates=%d msgs=%d",
			len(stSame.ListGroups(false)), len(stSame.ListPrivates(false)), stSame.CountMessages())
	}
	_ = stSame.Close()

	// Switching to a different AppID opens an isolated DB file and preserves APP_1's DB intact.
	st2, err := Open(Options{AppID: "APP_2", DBPath: dbPath})
	if err != nil {
		t.Fatalf("Open APP_2: %v", err)
	}
	if st2.DBPath() == dbPath {
		t.Fatalf("expected APP_2 to use an isolated DB path, got base path %q", st2.DBPath())
	}
	if len(st2.ListGroups(false)) != 0 || len(st2.ListPrivates(false)) != 0 || st2.CountMessages() != 0 {
		t.Fatalf("expected isolated empty cache for APP_2: groups=%d privates=%d msgs=%d",
			len(st2.ListGroups(false)), len(st2.ListPrivates(false)), st2.CountMessages())
	}
	st2.RecordInboundMessage("G2", CachedMessage{MessageID: "M2", Content: "bot2", Timestamp: 1700000010}, nil)
	_ = st2.Close()

	// Switching back to APP_1 must still have all of APP_1's original history preserved!
	st1Again, err := Open(Options{AppID: "APP_1", DBPath: dbPath})
	if err != nil {
		t.Fatalf("reopen APP_1: %v", err)
	}
	defer st1Again.Close()
	if len(st1Again.ListGroups(false)) != 1 || len(st1Again.ListPrivates(false)) != 1 || st1Again.CountMessages() != 2 {
		t.Fatalf("APP_1 data must be preserved after APP_2 switch: groups=%d privates=%d msgs=%d",
			len(st1Again.ListGroups(false)), len(st1Again.ListPrivates(false)), st1Again.CountMessages())
	}

	// Legacy untagged DB with existing data + PrevAppID switch preserves legacy DB for PrevAppID.
	legacyPath := dir + "/legacy.db"
	stLeg, err := Open(Options{DBPath: legacyPath})
	if err != nil {
		t.Fatalf("Open legacy: %v", err)
	}
	stLeg.RecordInboundMessage("G_LEG", CachedMessage{MessageID: "ML1", Content: "legacy", Timestamp: 1700000000}, nil)
	_ = stLeg.Close()

	stNew, err := Open(Options{AppID: "NEW_APP", PrevAppID: "OLD_APP", DBPath: legacyPath})
	if err != nil {
		t.Fatalf("Open NEW_APP with PrevAppID: %v", err)
	}
	defer stNew.Close()
	if stNew.DBPath() == legacyPath || stNew.CountMessages() != 0 {
		t.Fatalf("expected NEW_APP isolated from legacy DB, got path=%q msgs=%d", stNew.DBPath(), stNew.CountMessages())
	}
	stOld, err := Open(Options{AppID: "OLD_APP", DBPath: legacyPath})
	if err != nil {
		t.Fatalf("Open OLD_APP on legacy DB: %v", err)
	}
	defer stOld.Close()
	if stOld.DBPath() != legacyPath || stOld.CountMessages() != 1 {
		t.Fatalf("expected OLD_APP to inherit legacy DB with 1 msg, got path=%q msgs=%d", stOld.DBPath(), stOld.CountMessages())
	}
}

func TestConcurrentReservePassiveAnchorUniqueSeq(t *testing.T) {
	st := New(Options{RecentLimit: 100, PassiveMaxCount: 5})
	st.RecordInboundMessage("G_CONC", CachedMessage{
		MessageID: "IN_CONC",
		Content:   "hi",
		Timestamp: time.Now().Unix(),
	}, nil)

	const workers = 6
	results := make(chan int, workers)
	start := make(chan struct{})
	for i := 0; i < workers; i++ {
		go func() {
			<-start
			_, seq, ok := st.ReservePassiveAnchor("group", "G_CONC", "")
			if ok {
				results <- seq
			} else {
				results <- 0
			}
		}()
	}
	close(start)

	seen := make(map[int]bool)
	rejected := 0
	for i := 0; i < workers; i++ {
		seq := <-results
		if seq == 0 {
			rejected++
			continue
		}
		if seen[seq] {
			t.Fatalf("duplicate msg_seq %d allocated concurrently", seq)
		}
		seen[seq] = true
	}
	if len(seen) != 5 || rejected != 1 {
		t.Fatalf("expected 5 unique seqs and 1 rejected, got seen=%v rejected=%d", seen, rejected)
	}
}

func TestPassiveSeqNotReusedAfterReleaseAndRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "seq_restart.db")
	now := time.Now().Unix()

	st1, err := Open(Options{DBPath: dbPath, PassiveMaxCount: 5, C2CPassiveMaxCount: 4})
	if err != nil {
		t.Fatalf("Open st1: %v", err)
	}
	st1.RecordInboundMessage("G1", CachedMessage{MessageID: "GM1", Content: "hi", Timestamp: now}, nil)
	st1.RecordInboundC2C("U1", "CM1", "Alice", "Alice [openid:U1]", "hi", now)

	_, gSeq1, ok := st1.ReservePassiveAnchor("group", "G1", "")
	if !ok || gSeq1 != 1 {
		t.Fatalf("group seq1 = %d, ok=%t", gSeq1, ok)
	}
	_, gSeq2, ok := st1.ReservePassiveAnchor("group", "G1", "")
	if !ok || gSeq2 != 2 {
		t.Fatalf("group seq2 = %d, ok=%t", gSeq2, ok)
	}
	st1.ReleasePassiveAnchor("group", "G1", "GM1", 1)

	_, cSeq1, ok := st1.ReservePassiveAnchor("private", "U1", "")
	if !ok || cSeq1 != 1 {
		t.Fatalf("c2c seq1 = %d, ok=%t", cSeq1, ok)
	}
	_, cSeq2, ok := st1.ReservePassiveAnchor("private", "U1", "")
	if !ok || cSeq2 != 2 {
		t.Fatalf("c2c seq2 = %d, ok=%t", cSeq2, ok)
	}
	st1.ReleasePassiveAnchor("private", "U1", "CM1", 1)
	_ = st1.Close()

	// Reopen DB: next allocated seq MUST be 3 (never reallocate 2 which was already used).
	st2, err := Open(Options{DBPath: dbPath, PassiveMaxCount: 5, C2CPassiveMaxCount: 4})
	if err != nil {
		t.Fatalf("Open st2: %v", err)
	}
	defer st2.Close()

	_, gNext, ok := st2.ReservePassiveAnchor("group", "G1", "")
	if !ok || gNext != 3 {
		t.Fatalf("after restart group next seq = %d (ok=%t), want 3", gNext, ok)
	}
	_, cNext, ok := st2.ReservePassiveAnchor("private", "U1", "")
	if !ok || cNext != 3 {
		t.Fatalf("after restart c2c next seq = %d (ok=%t), want 3", cNext, ok)
	}
}
