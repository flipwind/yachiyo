package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"yachiyo/yachiyo-runtime/history"

	_ "github.com/mattn/go-sqlite3"
)

const history_schema_version = 1
const checkpoint_schema_version = 1

type Sqlite3Storage struct {
	Histories  []history.History
	Checkpoint history.Checkpoint
	db         *sql.DB
}

func (b *Sqlite3Storage) saveMessage(m history.History) error {
	var messageType string
	switch m.(type) {
	case history.AssistantMessage:
		messageType = "assistant_message"
	case history.UserMessage:
		messageType = "user_message"
	default:
		return fmt.Errorf("Unknown message type: %T", m)
	}

	payload, err := json.Marshal(m)
	if err != nil {
		return err
	}

	_, err = b.db.Exec(`
	INSERT INTO histories
		(kind, schema_version, create_at, payload)
		VALUES (?, ?, ?, ?)
	`, messageType, history_schema_version, time.Now().Unix(), string(payload))

	if err != nil {
		return err
	}

	return nil
}

func (b *Sqlite3Storage) restore() error {
	// 1. histories restore
	rows, err := b.db.Query(`
		SELECT kind, schema_version, payload
		FROM histories
		ORDER BY id ASC
	`)

	if err != nil {
		return fmt.Errorf("Query histories error: %w", err)
	}

	defer rows.Close()

	histories := make([]history.History, 0)

	for rows.Next() {
		var kind string
		var version int
		var payload string

		if err := rows.Scan(&kind, &version, &payload); err != nil {
			return fmt.Errorf("Scaning histories error: %w", err)
		}

		if version != history_schema_version {
			return fmt.Errorf("Unsupported history version: %d", version)
		}

		switch kind {
		case "user_message":
			var message history.UserMessage
			if err := json.Unmarshal([]byte(payload), &message); err != nil {
				return fmt.Errorf("Decoding user message error: %w", err)
			}
			histories = append(histories, message)
		case "assistant_message":
			var message history.AssistantMessage
			if err := json.Unmarshal([]byte(payload), &message); err != nil {
				return fmt.Errorf("Decoding user message error: %w", err)
			}
			histories = append(histories, message)
		default:
			return fmt.Errorf("Unknown history kind: %v", kind)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	b.Histories = histories

	// 2. checkpoint restore
	rows, err = b.db.Query(`
		SELECT id, state_at, schema_version, payload
		FROM checkpoint
		ORDER BY id DESC
	`)

	if err != nil {
		return fmt.Errorf("Query checkpoint error: %w", err)
	}

	defer rows.Close()

	var checkpoint history.Checkpoint

	for rows.Next() {
		var id int
		var state_at int
		var version int
		var payload string
		if err := rows.Scan(&id, &state_at, &version, &payload); err != nil {
			return fmt.Errorf("Scanning state error: %w", err)
		}
		
		if version != checkpoint_schema_version {
			return fmt.Errorf("Unsupported history version: %d", version)
		}

		if err := json.Unmarshal([]byte(payload), &checkpoint); err != nil {
			return fmt.Errorf("Decoding checkpoint error: %w", err)
		}

		if id == 1 {
			checkpoint.Exist = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	b.Checkpoint = checkpoint

	return nil
}

func New() (*Sqlite3Storage, error) {
	db, err := sql.Open("sqlite3", "yachiyo.db?_busy_timeout=5000")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

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

	storage := &Sqlite3Storage{
		Histories: make([]history.History, 0),
		db:        db,
	}

	if err := storage.restore(); err != nil {
		db.Close()
		return nil, fmt.Errorf("Restoring histories and state error: %w", err)
	}

	return storage, nil
}

func (b *Sqlite3Storage) Append(m history.History) error {
	if err := b.saveMessage(m); err != nil {
		return err
	}

	b.Histories = append(b.Histories, m)
	return nil
}

func (b *Sqlite3Storage) ListAll() []history.History {
	return b.Histories
}

func (b *Sqlite3Storage) Close() error {
	return b.db.Close()
}

func (b *Sqlite3Storage) SaveCheckpoint(point history.Checkpoint) error {
	payload, err := json.Marshal(point)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = b.db.Exec(`
    INSERT INTO checkpoint (id, schema_version, state_at, payload)
    VALUES (1, ?, ?, ?)
    ON CONFLICT(id) DO UPDATE SET
        schema_version = excluded.schema_version,
        state_at = excluded.state_at,
        payload = excluded.payload
	`, checkpoint_schema_version, point.StateAt.UnixMilli(), string(payload))

	if err != nil {
		return fmt.Errorf("store checkpoint error: %w", err)
	}

	return nil
}
