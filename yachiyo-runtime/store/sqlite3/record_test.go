package sqlite3

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
	"yachiyo/yachiyo-shared/address"
)

var embedTime = time.Date(2026, 1, 6, 8, 1, 6, 8, time.UTC)
var embedSnapshot = state.Snapshot{
	Drives:        state.NewDrives(),
	Determination: state.NewDetermination(),
	Energy:        1,
	LastActive:    embedTime,
	DefaultMood:   state.PAD{X: 0, Y: 1, Z: 1},
	Mood:          state.PAD{X: 0, Y: 1, Z: 1},
	Emotions:      []state.OCCEmotion{state.OCCEmotion{Type: state.Joy, Intensity: 1}},
}

func TestSummaryRound(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	want := conversation.Summary{
		ID:          uuid.NewV7(),
		CreateAt:    embedTime,
		Parent:      uuid.Nil(),
		FromEventID: uuid.NewV7(),
		ToEventID:   uuid.NewV7(),
		Content:     "test_summary",
	}
	if err := s.SaveSummary(want); err != nil {
		t.Fatal(err)
	}

	restored, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if restored.Summary == nil {
		t.Fatal("Summary not restored.")
	}
	if *restored.Summary != want {
		t.Errorf("\nGot > %+v, \nWant> %+v", *restored.Summary, want)
	}
}

func TestEventRound(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	want := event.UserMessage{
		ID:        uuid.NewV7(),
		Sequence:  168,
		Timestamp: embedTime,
		Snapshot:  embedSnapshot,
		Content:   "Hello from Yachiyo!",
		Author:    "kaguya",
		Platform:  "tsukuyomi",
		Address:   address.New("tsukuyomi"),
	}
	if err := s.Append(want); err != nil {
		t.Fatal(err)
	}

	restored, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored.Events[0], want) {
		t.Errorf("\nGot > %+v, \nWant> %+v", restored.Events[0], want)
	}
}

func TestCheckpointRound(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	want := store.Checkpoint{
		SavedAt:  embedTime,
		State:    embedSnapshot,
		Note:     "Iroha suki!",
		Sequence: 168,
	}

	if err := s.SaveCheckpoint(want); err != nil {
		t.Fatal(err)
	}

	restored, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if restored.Checkpoint == nil {
		t.Fatal("Checkpoint not restored.")
	}
	if !reflect.DeepEqual(*restored.Checkpoint, want) {
		t.Errorf("\nGot > %+v, \nWant> %+v", *restored.Checkpoint, want)
	}
}
