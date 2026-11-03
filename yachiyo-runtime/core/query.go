package core

import (
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/event"
	"yachiyo/yachiyo-runtime/trigger"
)

func (c *Core) processRuntimeStateRequest(t *trigger.RuntimeStateRequest) action.Action {
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
		Energy:   c.Energy.Value,
		Affect: struct {
			Octants   string
			Pleasure  float64
			Arousal   float64
			Dominance float64
			Actives   []action.OCCEmotion
		}{
			Octants:   c.Affect.Current.Octants(),
			Pleasure:  c.Affect.Current.Pleasure(),
			Arousal:   c.Affect.Current.Arousal(),
			Dominance: c.Affect.Current.Dominance(),
			Actives:   actives,
		},
		Address: t.Address,
	}
}

func (c *Core) processMessageHistoryRequest(t *trigger.MessageHistoryRequest) action.Action {
	hist := c.getHistory()
	msgs := make([]event.Message, 0)

	for _, h := range hist {
		switch m := h.(type) {
		case event.AssistantMessage:
			if m.Address != t.Address {
				break
			}

			if m.Initiative == true && m.Reply == false {
				break
			}

			msgs = append(msgs, m)
		case event.UserMessage:
			if m.Address != t.Address {
				break
			}

			msgs = append(msgs, m)
		}
	}

	return &action.MessageHistory{
		Messages: msgs,
		Address:  t.Address,
	}
}
