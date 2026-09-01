package ycontext

import "sync"

const version = "0.1.0"

type MutableContext struct {
	context Context

	mu sync.RWMutex
	update func(*Context)
}

type Context struct {
	Name    string
	Version string
}

func Generate(update func(*Context)) *MutableContext {
	mc := MutableContext{
		update: update,

		// Const values
		context: Context{
			Version: version,
		},
	}

	mc.Update()
	return &mc
}

func (mc *MutableContext) Update() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.update(&mc.context)
}

// Get functions

func (mc *MutableContext) GetContext() Context {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.context
}
