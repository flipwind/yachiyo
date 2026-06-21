package core

import (
	"fmt"
	"yachiyo/yachiyo-core/event"
	"yachiyo/yachiyo-core/memory"
	"yachiyo/yachiyo-core/memory/basic"
)

type Core struct{
	Memory memory.MemoryStorage
}

func New() *Core{
	return &Core{
		Memory: basic.New(),
	}
}

func (c *Core) Process(e *event.UserMessageEvent) string{
	c.Memory.Remember(memory.Memory{
		Content: e.Message.String(),
	})
	
	return fmt.Sprintf("%v", c.Memory.ListAll())
}