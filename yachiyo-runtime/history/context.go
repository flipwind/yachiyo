package history

import (
	"fmt"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/event"
)

type CompressNode struct {
	Empty             bool
	ID                uuid.UUID
	CompressTimestamp time.Time
	SubCompressNodeID uuid.UUID

	CompressEventIDFrom uuid.UUID
	CompressEventIDTo   uuid.UUID

	Content string
}

// ContextWindow manages a context window which LLM would use.
//
// In a ContextWindow, it may include a list [CompressNode<3>, Event<168>, Event<169>],
// and CompressNode<3> may be [CompressNode<2>, Event<23>, ..., Event<167>].
//
// It's clear that this is a recursive struct.
type ContextWindow struct {
	CurrentNode      CompressNode
	CurrentHistories []event.Event
}

func (c CompressNode) Rend() string {
	return fmt.Sprintf(`<CompressNode>[Compressed at %s]
%s`, c.CompressTimestamp.Format("2006.01.02 15:04:05"), c.Content)
}

func (c ContextWindow) IsEmpty() bool {
	return len(c.CurrentHistories) == 0 && c.CurrentNode.Empty == true
}
