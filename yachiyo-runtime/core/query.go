package core

import (
	"time"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-shared/action"
	"yachiyo/yachiyo-shared/trigger"
)

func (c *Core) processRuntimeStateRequest(t trigger.RuntimeStateRequest) action.Action {
	actives := make([]action.OCCEmotion, 0)
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, emo := range c.Affect.Emotions {
		actives = append(actives, action.OCCEmotion{
			Type:      emo.Type.Name,
			Intensity: emo.Intensity,
		})
	}

	return &action.RuntimeState{
		Timetick: time.Now().Unix(),
		Energy:   c.Energy.Value(),
		Affect: action.Affect{
			Octants:   c.Affect.Current.Octants(),
			Pleasure:  c.Affect.Current.Pleasure(),
			Arousal:   c.Affect.Current.Arousal(),
			Dominance: c.Affect.Current.Dominance(),
			Actives:   actives,
		},
		Address: t.Address,
	}
}

func (c *Core) processMessageHistoryRequest(t trigger.MessageHistoryRequest) action.Action {
	hist := c.getHistory()
	msgs := make([]action.HistoryMessage, 0)

	for _, h := range hist {
		switch m := h.(type) {
		case event.AssistantMessage:
			if m.Address != t.Address {
				break
			}

			if m.Initiative == true && m.Reply == false {
				break
			}

			msgs = append(msgs, &action.AssistantHistoryMessage{
				Content:    m.Content,
				Timestamp:  m.Timestamp,
				Initiative: m.Initiative,
				Reply:      m.Reply,
			})
		case event.UserMessage:
			if m.Address != t.Address {
				break
			}

			msgs = append(msgs, &action.UserHistoryMessage{
				Content:   m.Content,
				Timestamp: m.Timestamp,
			})
		}
	}

	return &action.MessageHistory{
		Messages: msgs,
		Address:  t.Address,
	}
}
