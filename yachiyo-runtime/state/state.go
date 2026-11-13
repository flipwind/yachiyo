package state

import (
	"fmt"
	"slices"
	"time"
)

type State struct {
	drives        Drives
	determination Determination
	energy        Energy
	affect        Engine
	lastActive    time.Time
}

type Snapshot struct {
	Drives        Drives
	Determination Determination
	Energy        float64
	LastActive    time.Time

	DefaultMood PAD
	Mood        PAD
	Emotions    []OCCEmotion
}

// Change represents a change applied to State.
type Change struct {
	// Play as some delta increasement.
	SocialDesire int
	Interest     int
	Emotions     []OCCEmotion

	// Designed to cover State.
	Determination Determination
}

func Initial(personality BigFive, now time.Time) Snapshot {
	affect := NewEngine(personality, nil, nil)
	return Snapshot{
		Drives:        NewDrives(),
		Determination: NewDetermination(),
		Energy:        0,
		LastActive:    now,
		DefaultMood:   affect.Default,
		Mood:          affect.Current.Position,
		Emotions:      affect.Emotions,
	}
}

func Restore(personality BigFive, energy Energy, snap Snapshot) *State {
	emotion := slices.Clone(snap.Emotions)
	affect := NewEngine(personality, &snap.Mood, &emotion)

	return &State{
		drives:        snap.Drives,
		determination: snap.Determination,
		energy:        energy,
		affect:        affect,
		lastActive:    snap.LastActive,
	}
}

func (s *State) Snapshot() Snapshot {
	return Snapshot{
		Drives:        s.drives,
		Determination: s.determination,
		Energy:        s.energy.Value(),
		LastActive:    s.lastActive,
		DefaultMood:   s.affect.Default,
		Mood:          s.affect.Current.Position,
		Emotions:      slices.Clone(s.affect.Emotions),
	}
}

func (s *State) Advance(d time.Duration, now time.Time) error {
	// State press
	for _, named := range s.drives.All() {
		named.Drive.Press((1.4 * 5) / (3600 * 1) * d.Seconds())
	}

	// Affect Engine
	s.affect.Simulate(d.Seconds())

	// Initiative
	if err := s.energy.Update(now); err != nil {
		return fmt.Errorf("Energy updated error: %w", err)
	}

	return nil
}

func (s *State) Relieve(d time.Duration) {
	// Relieve
	for _, named := range s.drives.All() {
		named.Drive.Relieve(0.5 * d.Seconds())
	}
}

func (s *State) Touch(now time.Time) {
	s.lastActive = now
}

func (s *State) Apply(c Change) {
	switch c.SocialDesire {
	case 1:
		s.drives.SocialDesire.Increase()
	case -1:
		s.drives.SocialDesire.Decrease()
	}

	switch c.Interest {
	case 1:
		s.drives.Interest.Increase()
	case -1:
		s.drives.Interest.Decrease()
	}

	for _, affect := range c.Emotions {
		s.affect.Press(affect)
	}

	s.determination = c.Determination
}
