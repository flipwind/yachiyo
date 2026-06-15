package storage

import (
	"context"
	"errors"
	"yachiyo/yachiyo-core/chat"
)

type MemoryStorage interface {
	SaveMessage(ctx context.Context, msg *chat.Message) error
	GetMessage(ctx context.Context, uuid string) (*chat.Message, error)
	GetHistory(ctx context.Context, session_id string, length int64) ([]*chat.Message, error)
	Close()
}

func NewMemoryStorage(driverType string, dbPath string) (MemoryStorage, error){
	switch driverType {
	case "sqlite3":
		return NewSqliteMemoryStorage(dbPath)
	default:
		return nil, errors.New("Unsupported driver: " + driverType)
	}
}