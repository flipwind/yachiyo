package sqlite3

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/history"

	_ "github.com/mattn/go-sqlite3"
)

const history_schema_version = 1
const checkpoint_schema_version = 1
const compressnode_schema_version = 1

type Sqlite3Storage struct {
	Histories   []event.History
	db          *sql.DB
	nowSequence atomic.Int64 //exhaustruct:optional
}

func (b *Sqlite3Storage) getMaxSequence() int64 {
	var maxSequence int64
	for _, h := range b.Histories {
		maxSequence = max(maxSequence, h.GetSequence())
	}

	return maxSequence
}

func (b *Sqlite3Storage) saveMessage(m event.History) error {
	var messageType string
	switch m.(type) {
	case event.AssistantMessage:
		messageType = "assistant_message"
	case event.UserMessage:
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

func (b *Sqlite3Storage) restore() (history.RestoreResult, error) {
	var result history.RestoreResult
	// 1. histories restore
	rows, err := b.db.Query(`
		SELECT kind, schema_version, payload
		FROM histories
		ORDER BY id ASC
	`)

	if err != nil {
		return result, fmt.Errorf("Query histories error: %w", err)
	}

	defer rows.Close()

	histories := make([]event.History, 0)

	for rows.Next() {
		var kind string
		var version int
		var payload string

		if err := rows.Scan(&kind, &version, &payload); err != nil {
			return result, fmt.Errorf("Scaning histories error: %w", err)
		}

		if version != history_schema_version {
			return result, fmt.Errorf("Unsupported history version: %d", version)
		}

		switch kind {
		case "user_message":
			var message event.UserMessage
			if err := json.Unmarshal([]byte(payload), &message); err != nil {
				return result, fmt.Errorf("Decoding user message error: %w", err)
			}
			histories = append(histories, message)
		case "assistant_message":
			var message event.AssistantMessage
			if err := json.Unmarshal([]byte(payload), &message); err != nil {
				return result, fmt.Errorf("Decoding user message error: %w", err)
			}
			histories = append(histories, message)
		default:
			return result, fmt.Errorf("Unknown history kind: %v", kind)
		}
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	b.Histories = histories

	// 2. checkpoint restore
	rows, err = b.db.Query(`
		SELECT id, state_at, schema_version, payload
		FROM checkpoint
		ORDER BY id DESC
	`)

	if err != nil {
		return result, fmt.Errorf("Query checkpoint error: %w", err)
	}

	defer rows.Close()

	var checkpoint history.Checkpoint

	for rows.Next() {
		var id int
		var state_at int
		var version int
		var payload string
		if err := rows.Scan(&id, &state_at, &version, &payload); err != nil {
			return result, fmt.Errorf("Scanning checkpoint error: %w", err)
		}

		if version != checkpoint_schema_version {
			return result, fmt.Errorf("Unsupported checkpoint version: %d", version)
		}

		if err := json.Unmarshal([]byte(payload), &checkpoint); err != nil {
			return result, fmt.Errorf("Decoding checkpoint error: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return result, err
	}

	result.Checkpoint = checkpoint
	b.nowSequence.Store(max(checkpoint.NowSequence, b.getMaxSequence()))

	// 3. contextWindow restore
	rows, err = b.db.Query(`
		SELECT id, schema_version, node_id, payload
		FROM contextnode
		ORDER BY id DESC
		LIMIT 1;
	`)

	if err != nil {
		return result, fmt.Errorf("Query contextnode error: %w", err)
	}

	defer rows.Close()

	var compressnode history.CompressNode
	compressnode.Empty = true

	for rows.Next() {
		var id int
		var version int
		var node_id string
		var payload string
		if err := rows.Scan(&id, &version, &node_id, &payload); err != nil {
			return result, fmt.Errorf("Scanning contextnode error: %w", err)
		}

		if version != compressnode_schema_version {
			return result, fmt.Errorf("Unsupported contextnode version: %d", version)
		}

		if err := json.Unmarshal([]byte(payload), &compressnode); err != nil {
			return result, fmt.Errorf("Decoding contextnode error: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return result, err
	}

	result.LatestNode = compressnode
	return result, nil
}

func New() (*Sqlite3Storage, history.RestoreResult, error) {
	var restoreResult history.RestoreResult
	db, err := sql.Open("sqlite3", "yachiyo.db?_busy_timeout=5000")
	if err != nil {
		return nil, restoreResult, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, restoreResult, err
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
		return nil, restoreResult, fmt.Errorf("Processing storage of `histories` error: %w", err)
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
		return nil, restoreResult, fmt.Errorf("Processing storage of `checkpoint` error: %w", err)
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
		return nil, restoreResult, fmt.Errorf("Processing storage of `contextnode` error: %w", err)
	}

	storage := &Sqlite3Storage{
		Histories: make([]event.History, 0),
		db:        db,
	}

	if restoreResult, err = storage.restore(); err != nil {
		db.Close()
		return nil, restoreResult, fmt.Errorf("Restoring histories and state error: %w", err)
	}

	return storage, restoreResult, nil
}

func (b *Sqlite3Storage) Append(m event.History) error {
	if err := b.saveMessage(m); err != nil {
		return err
	}

	b.Histories = append(b.Histories, m)
	return nil
}

func (b *Sqlite3Storage) ListAll() []event.History {
	return b.Histories
}

func (b *Sqlite3Storage) ListAfter(uuid uuid.UUID) []event.History {
	var history_len = len(b.Histories)
	var index = history_len - 1
	var findflag = false
	for i := index; i >= 0; i-- {
		if b.Histories[i].GetID() == uuid {
			index = i
			findflag = true
			break
		}
	}

	if findflag == false {
		return b.Histories
	}

	return b.Histories[index+1:]
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

func (b *Sqlite3Storage) SaveCompressNode(node history.CompressNode) error {
	payload, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	_, err = b.db.Exec(`
    INSERT INTO contextnode
		(schema_version, node_id, payload)
		VALUES (?, ?, ?)
	`, compressnode_schema_version, node.ID, string(payload))

	if err != nil {
		return fmt.Errorf("store contextnode error: %w", err)
	}

	return nil
}

func (b *Sqlite3Storage) NextSequence() int64 {
	return b.nowSequence.Add(1)
}

func (b *Sqlite3Storage) NowSequence() int64 {
	return b.nowSequence.Load()
}
