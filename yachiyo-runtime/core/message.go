package core

import (
	"yachiyo/yachiyo-runtime/prompt"
)

const keepNote = "-1"

func debugOutput(answer []string, c_formal Snapshot, c_later Snapshot) {
	ylog.Debug(`== DEBUG MESSAGE ==
Yachiyo > %v
* Formal:
%s
* Later:
%s
%+v
* Session Context: %s`,
		answer, prompt.Drives(c_formal.State.Drives),
		prompt.Drives(c_later.State.Drives), c_later.State.Determination,
		c_later.Note)
}

func (c *Core) applyOutput(out output) {
	c.state.Apply(out.Change)

	if out.Note != keepNote {
		c.conv.SetNote(out.Note)
	}
}
