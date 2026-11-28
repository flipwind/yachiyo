package sqlite3

import (
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/store"
	"yachiyo/yachiyo-shared/address"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Store.Sqlite3")

type padRecord struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
	Z float64 `json:"Z"`
}

func toPADRecord(p state.PAD) padRecord {
	return padRecord{
		X: p.X,
		Y: p.Y,
		Z: p.Z,
	}
}

func (r padRecord) pad() state.PAD {
	return state.PAD{
		X: r.X,
		Y: r.Y,
		Z: r.Z,
	}
}

type driveRecord struct {
	Value float64 `json:"value"`
}

func toDriveRecord(d state.Drive) driveRecord {
	return driveRecord{
		Value: d.Value,
	}
}

func (r driveRecord) drive() state.Drive {
	return state.Drive{
		Value: r.Value,
	}
}

type drivesRecord struct {
	SocialDesire driveRecord `json:"social_desire"`
	Interest     driveRecord `json:"interest"`
}

func toDrivesRecord(d state.Drives) drivesRecord {
	return drivesRecord{
		SocialDesire: toDriveRecord(d.SocialDesire),
		Interest:     toDriveRecord(d.Interest),
	}
}

func (r drivesRecord) drives() state.Drives {
	return state.Drives{
		SocialDesire: r.SocialDesire.drive(),
		Interest:     r.Interest.drive(),
	}
}

type determinationRecord struct {
	NewSession   bool `json:"new_session"`
	WaitForReply bool `json:"wait_for_reply"`
}

func toDeterminationRecord(d state.Determination) determinationRecord {
	return determinationRecord{
		NewSession:   d.NewSession,
		WaitForReply: d.WaitForReply,
	}
}

func (r determinationRecord) determination() state.Determination {
	return state.Determination{
		NewSession:   r.NewSession,
		WaitForReply: r.WaitForReply,
	}
}

type addressRecord struct {
	Content string `json:"content"`
}

func toAddressRecord(a address.Address) addressRecord {
	return addressRecord{
		Content: a.Content,
	}
}

func (r addressRecord) address() address.Address {
	return address.Address{
		Content: r.Content,
	}
}

type occemotionRecord struct {
	Type      string  `json:"type"`
	Intensity float64 `json:"intensity"`
}

func toOCCEmotionRecord(e state.OCCEmotion) occemotionRecord {
	return occemotionRecord{
		Type:      e.Type.Name,
		Intensity: e.Intensity,
	}
}

func (r occemotionRecord) emotion() (state.OCCEmotion, bool) {
	typ, ok := state.OCCEmotionMap[r.Type]
	return state.OCCEmotion{
		Type:      typ,
		Intensity: r.Intensity,
	}, ok
}

type snapshotRecord struct {
	Drives        drivesRecord        `json:"drives"`
	Determination determinationRecord `json:"determination"`
	Energy        float64             `json:"energy"`
	LastActive    time.Time           `json:"last_active"`

	DefaultMood padRecord          `json:"default_mood"`
	Mood        padRecord          `json:"mood"`
	Emotions    []occemotionRecord `json:"emotions"`
}

func toSnapshotRecord(snap state.Snapshot) snapshotRecord {
	emotionRecord := make([]occemotionRecord, 0, len(snap.Emotions))
	for _, emo := range snap.Emotions {
		emotionRecord = append(emotionRecord, toOCCEmotionRecord(emo))
	}

	return snapshotRecord{
		Drives:        toDrivesRecord(snap.Drives),
		Determination: toDeterminationRecord(snap.Determination),
		Energy:        snap.Energy,
		LastActive:    snap.LastActive,
		DefaultMood:   toPADRecord(snap.DefaultMood),
		Mood:          toPADRecord(snap.Mood),
		Emotions:      emotionRecord,
	}
}

func (r snapshotRecord) snapshot() state.Snapshot {
	emotions := make([]state.OCCEmotion, 0, len(r.Emotions))
	for _, emotion := range r.Emotions {
		e, ok := emotion.emotion()
		if !ok {
			ylog.Warn("Dropping unknown emotion type <%s>", emotion.Type)
			continue
		}
		emotions = append(emotions, e)
	}

	return state.Snapshot{
		Drives:        r.Drives.drives(),
		Determination: r.Determination.determination(),
		Energy:        r.Energy,
		LastActive:    r.LastActive,
		DefaultMood:   r.DefaultMood.pad(),
		Mood:          r.Mood.pad(),
		Emotions:      emotions,
	}

}

type metaRecord struct {
	ID        uuid.UUID      `json:"id"`
	Sequence  int64          `json:"sequence"`
	Timestamp time.Time      `json:"timestamp"`
	Snapshot  snapshotRecord `json:"snapshot"`
}

func toMetaRecord(m event.Meta) metaRecord {
	return metaRecord{
		ID:        m.ID,
		Sequence:  m.Sequence,
		Timestamp: m.Timestamp,
		Snapshot:  toSnapshotRecord(m.Snapshot),
	}
}

func (r metaRecord) meta() event.Meta {
	return event.Meta{
		ID:        r.ID,
		Sequence:  r.Sequence,
		Timestamp: r.Timestamp,
		Snapshot:  r.Snapshot.snapshot(),
	}
}

type userMessageRecord struct {
	metaRecord
	Content  string        `json:"content"`
	Author   string        `json:"author"`
	Platform string        `json:"platform"`
	Address  addressRecord `json:"address"`
}

func toUserMessageRecord(m event.UserMessage) userMessageRecord {
	return userMessageRecord{
		metaRecord: toMetaRecord(m.Meta),
		Content:    m.Content,
		Author:     m.Author,
		Platform:   m.Platform,
		Address:    toAddressRecord(m.Address),
	}
}

func (r userMessageRecord) event() event.UserMessage {
	return event.UserMessage{
		Meta:     r.meta(),
		Content:  r.Content,
		Author:   r.Author,
		Platform: r.Platform,
		Address:  r.Address.address(),
	}
}

type assistantMessageRecord struct {
	metaRecord
	Content    []string      `json:"content"`
	Initiative bool          `json:"initiative"`
	Address    addressRecord `json:"address"`
}

func toAssistantMessageRecord(m event.AssistantMessage) assistantMessageRecord {
	return assistantMessageRecord{
		metaRecord: toMetaRecord(m.Meta),
		Content:    m.Content,
		Initiative: m.Initiative,
		Address:    toAddressRecord(m.Address),
	}
}

func (r assistantMessageRecord) event() event.AssistantMessage {
	return event.AssistantMessage{
		Meta:       r.meta(),
		Content:    r.Content,
		Initiative: r.Initiative,
		Address:    r.Address.address(),
	}
}

type checkpointRecord struct {
	SavedAt  time.Time      `json:"saved_at"`
	State    snapshotRecord `json:"state"`
	Note     string         `json:"note"`
	Sequence int64          `json:"sequence"`
}

func toCheckpointRecord(c store.Checkpoint) checkpointRecord {
	return checkpointRecord{
		SavedAt:  c.SavedAt,
		State:    toSnapshotRecord(c.State),
		Note:     c.Note,
		Sequence: c.Sequence,
	}
}

func (r checkpointRecord) checkpoint() store.Checkpoint {
	return store.Checkpoint{
		SavedAt:  r.SavedAt,
		State:    r.State.snapshot(),
		Note:     r.Note,
		Sequence: r.Sequence,
	}
}

type summaryRecord struct {
	ID          uuid.UUID `json:"id"`
	CreateAt    time.Time `json:"create_at"`
	Parent      uuid.UUID `json:"parent"`
	FromEventID uuid.UUID `json:"from_event_id"`
	ToEventID   uuid.UUID `json:"to_event_id"`
	Content     string    `json:"content"`
}

func toSummaryRecord(s conversation.Summary) summaryRecord {
	return summaryRecord{
		ID:          s.ID,
		CreateAt:    s.CreateAt,
		Parent:      s.Parent,
		FromEventID: s.FromEventID,
		ToEventID:   s.ToEventID,
		Content:     s.Content,
	}
}

func (r summaryRecord) summary() conversation.Summary {
	return conversation.Summary{
		ID:          r.ID,
		CreateAt:    r.CreateAt,
		Parent:      r.Parent,
		FromEventID: r.FromEventID,
		ToEventID:   r.ToEventID,
		Content:     r.Content,
	}
}
