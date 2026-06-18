package plugin

import (
	"sync"
	"yachiyo/yachiyo-utils/logger"
)

type EventBus struct {
	mu sync.RWMutex
	EventMap map[string] []*PluginRuntime
}

type Event struct {
	Type string
}

func NewEventBus() *EventBus {
	return &EventBus{
		EventMap: make(map[string][]*PluginRuntime),
	}
}

func (b *EventBus) Register(d *PluginDriver){
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, plugin := range d.plugins{
		for _, event := range plugin.Config.SubscribedEvents {
			b.EventMap[event] = append(b.EventMap[event], plugin)
		}
	}
}

func (b *EventBus) Publish(e *Event){
	for _, plugin := range b.EventMap[e.Type] {
		logger.Debug(sourcename, "Distribute event {%v} to plugin {%v}", e.Type, plugin.Name)

		// TODO: gRPC Distribution
	}
}