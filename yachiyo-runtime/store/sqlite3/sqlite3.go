package sqlite3

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/store"

	_ "github.com/mattn/go-sqlite3"
)

const historySchemaVersion = 1
const checkpointSchemaVersion = 1
const compressnodeSchemaVersion = 1

const (
	kindUserMessage      = "user_message"
	kindAssistantMessage = "assistant_message"
)

// role as a fuse
var _ store.Store = (*Storage)(nil)

type Storage struct {
	db *sql.DB
}

func Open(path string) (*Storage, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_busy_timeout=5000", path))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	// Create or check tables
	if _, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS histories(
		id INTEGER PRIMARY KEY,
		kind TEXT NOT NULL,
		schema_version INTEGER NOT NULL,
		create_at INTEGER NOT NULL,
		payload TEXT NOT NULL
	);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("Processing storage of `histories` error: %w", err)
	}

	if _, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS checkpoint(
		id INTEGER PRIMARY KEY,
		state_at INTEGER NOT NULL,
		schema_version INTEGER NOT NULL,
		payload TEXT NOT NULL
	);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("Processing storage of `checkpoint` error: %w", err)
	}

	if _, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS contextnode(
		id INTEGER PRIMARY KEY,
		schema_version INTEGER NOT NULL,
		node_id TEXT NOT NULL,
		payload TEXT NOT NULL
	);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("Processing storage of `contextnode` error: %w", err)
	}

	return &Storage{
		db: db,
	}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) Load() (store.Restored, error) {
	events, err := s.loadEvents()
	if err != nil {
		return store.Restored{}, fmt.Errorf("Error when loading events: %w", err)
	}

	checkpoint, err := s.loadCheckpoint()
	if err != nil {
		return store.Restored{}, fmt.Errorf("Error when loading checkpoint: %w", err)
	}

	summary, err := s.loadSummary()
	if err != nil {
		return store.Restored{}, fmt.Errorf("Error when loading summary: %w", err)
	}

	return store.Restored{
		Events:     events,
		Checkpoint: checkpoint,
		Summary:    summary,
	}, nil
}

func (s *Storage) loadEvents() ([]event.Event, error) {
	rows, err := s.db.Query(`
		SELECT kind, schema_version, payload
		FROM histories
		ORDER BY id ASC
	`)

	if err != nil {
		return []event.Event{}, fmt.Errorf("Query histories error: %w", err)
	}

	defer rows.Close()

	histories := make([]event.Event, 0)

	for rows.Next() {
		var kind string
		var version int
		var payload string

		if err := rows.Scan(&kind, &version, &payload); err != nil {
			return []event.Event{}, fmt.Errorf("Scaning histories error: %w", err)
		}

		if version != historySchemaVersion {
			return []event.Event{}, fmt.Errorf("Unsupported history version: %d", version)
		}

		switch kind {
		case kindUserMessage:
			var message userMessageRecord
			if err := json.Unmarshal([]byte(payload), &message); err != nil {
				return []event.Event{}, fmt.Errorf("Decoding user message error: %w", err)
			}
			histories = append(histories, message.event())
		case kindAssistantMessage:
			var message assistantMessageRecord
			if err := json.Unmarshal([]byte(payload), &message); err != nil {
				return []event.Event{}, fmt.Errorf("Decoding assistant message error: %w", err)
			}
			histories = append(histories, message.event())
		default:
			return []event.Event{}, fmt.Errorf("Unknown history kind: %v", kind)
		}
	}

	if err := rows.Err(); err != nil {
		return []event.Event{}, err
	}
	return histories, nil
}

func (s *Storage) loadCheckpoint() (*store.Checkpoint, error) {
	row := s.db.QueryRow(`
		SELECT id, state_at, schema_version, payload
		FROM checkpoint
		ORDER BY id DESC
	`)

	var checkpoint checkpointRecord

	var id int
	var state_at int
	var version int
	var payload string
	if err := row.Scan(&id, &state_at, &version, &payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("Scanning checkpoint error: %w", err)
	}

	if version != checkpointSchemaVersion {
		return nil, fmt.Errorf("Unsupported checkpoint version: %d", version)
	}

	if err := json.Unmarshal([]byte(payload), &checkpoint); err != nil {
		return nil, fmt.Errorf("Decoding checkpoint error: %w", err)
	}

	cp := checkpoint.checkpoint()

	return &cp, nil
}

func (s *Storage) loadSummary() (*conversation.Summary, error) {
	row := s.db.QueryRow(`
		SELECT schema_version, payload
		FROM contextnode
		ORDER BY id DESC
		LIMIT 1;
	`)

	var summary summaryRecord

	var version int
	var payload string
	if err := row.Scan(&version, &payload); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("Scanning contextnode error: %w", err)
	}

	if version != compressnodeSchemaVersion {
		return nil, fmt.Errorf("Unsupported contextnode version: %d", version)
	}

	if err := json.Unmarshal([]byte(payload), &summary); err != nil {
		return nil, fmt.Errorf("Decoding contextnode error: %w", err)
	}

	sum := summary.summary()

	return &sum, nil
}

func (s *Storage) Append(e event.Event) error {
	var messageType string
	var record any
	switch m := e.(type) {
	case event.AssistantMessage:
		messageType, record = kindAssistantMessage, toAssistantMessageRecord(m)
	case event.UserMessage:
		messageType, record = kindUserMessage, toUserMessageRecord(m)
	default:
		return fmt.Errorf("Unknown event type: %T", e)
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
	INSERT INTO histories
		(kind, schema_version, create_at, payload)
		VALUES (?, ?, ?, ?)
	`, messageType, historySchemaVersion, time.Now().Unix(), string(payload))

	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) SaveCheckpoint(p store.Checkpoint) error {
	point := toCheckpointRecord(p)
	payload, err := json.Marshal(point)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = s.db.Exec(`
    INSERT INTO checkpoint (id, schema_version, state_at, payload)
    VALUES (1, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
        schema_version = excluded.schema_version,
        state_at = excluded.state_at,
        payload = excluded.payload
	`, checkpointSchemaVersion, point.SavedAt.UnixMilli(), string(payload))

	if err != nil {
		return fmt.Errorf("store checkpoint error: %w", err)
	}

	return nil
}

func (s *Storage) SaveSummary(sum conversation.Summary) error {
	summary := toSummaryRecord(sum)
	payload, err := json.Marshal(summary)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = s.db.Exec(`
    INSERT INTO contextnode
		(schema_version, node_id, payload)
		VALUES (?, ?, ?)
	`, compressnodeSchemaVersion, summary.ID, string(payload))

	if err != nil {
		return fmt.Errorf("store contextnode error: %w", err)
	}

	return nil
}
