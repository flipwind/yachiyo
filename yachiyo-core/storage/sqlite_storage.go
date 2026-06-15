package storage

import (
	"context"
	"slices"
	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-utils/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var sourcename = "Yachiyo.Storage/sqlite3"

type SqliteMemoryStorage struct {
	db *sqlx.DB
}

type MessageRow struct {
	ID int64 `db:"id"`
	UUID string `db:"uuid"`
	SessionId string `db:"session_id"`
	Role string `db:"role"`
	Content string `db:"content"`
	Timestamp int64 `db:"timestamp"`
}

func MessageToRow(m *chat.Message) *MessageRow{
	return &MessageRow{
		UUID: m.Uuid,
		SessionId: m.SessionId,
		Role: m.Role,
		Content: m.Content,
		Timestamp: m.Timestamp,
	}
}

func RowToMessage(m *MessageRow) *chat.Message{
	return &chat.Message{
		Id: m.ID,
		Uuid: m.UUID,
		SessionId: m.SessionId,
		Role: m.Role,
		Content: m.Content,
		Timestamp: m.Timestamp,
	}
}

func NewSqliteMemoryStorage(dbPath string) (*SqliteMemoryStorage, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		logger.Error(sourcename, "Memory DB creation failed: %v", err)
		return nil, err
	}

	schema := `
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT NOT NULL,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp INTEGER NOT NULL
		);`

	if _, err := db.Exec(schema); err != nil {
		logger.Error(sourcename, "Initialize memory DB failed: %v", err)
		db.Close()
		return nil, err
	}

	return &SqliteMemoryStorage{db: db}, nil
}

func (s *SqliteMemoryStorage) Close() {
	s.db.Close()
}

func (s *SqliteMemoryStorage) SaveMessage(ctx context.Context, msg *chat.Message) error {
	schema := `
		INSERT INTO messages (uuid, session_id, role, content, timestamp)
	    VALUES (:uuid, :session_id, :role, :content, :timestamp)`
	_, err := s.db.NamedExecContext(ctx, schema, MessageToRow(msg))
	if err != nil {
		logger.Error(sourcename, "Saving message {%#v} failed: %v", msg, err)
		return err
	}
	return nil
}

func (s *SqliteMemoryStorage) GetMessage(ctx context.Context, uuid string) (*chat.Message, error) {
	query := `
		SELECT id, uuid, session_id, role, content, timestamp
		FROM messages WHERE uuid = ?`
	var row MessageRow
	err := s.db.GetContext(ctx, &row, query, uuid)
	if err != nil {
		logger.Error(sourcename, "Getting message {%v} failed: %v", uuid, err)
		return nil, err
	}

	return RowToMessage(&row), nil
}

func (s *SqliteMemoryStorage) GetHistory(ctx context.Context, session_id string, length int64) ([]*chat.Message, error) {
	var rows []MessageRow
	var query string
	var err error

	if length != -1 {
		query = `
			SELECT id, uuid, session_id, role, content, timestamp
			FROM messages WHERE session_id = ?
			ORDER BY id DESC
			LIMIT ?`

		err = s.db.SelectContext(ctx, &rows, query, session_id, length)
		if err != nil {
			logger.Error(sourcename, "Getting history of {%v} failed: %v", session_id, err)
			return nil, err
		} 

		slices.Reverse(rows)
	} else {
		query = `
			SELECT id, uuid, session_id, role, content, timestamp
			FROM messages WHERE session_id = ?
			ORDER BY id ASC`

		err = s.db.SelectContext(ctx, &rows, query, session_id)
		if err != nil {
			logger.Error(sourcename, "Getting history of {%v} failed: %v", session_id, err)
			return nil, err
		} 
	}

	var messages []*chat.Message
	for _, row := range rows {
		messages = append(messages, RowToMessage(&row))
	}

	return messages, nil
}