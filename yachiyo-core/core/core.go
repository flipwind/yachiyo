package core

import (
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