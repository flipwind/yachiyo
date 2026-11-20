package core

import (
	"time"
	"yachiyo/yachiyo-runtime/state"
)

type InitiativePolicy interface {
	ShouldInitiative(snap state.Snapshot, now time.Time) bool
}

type NeverInitiative struct{}

func (NeverInitiative) ShouldInitiative(snap state.Snapshot, now time.Time) bool { return false }
