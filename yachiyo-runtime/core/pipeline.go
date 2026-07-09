package core

import (
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/trigger"
)

type Pipeline struct {
	Raw          chan trigger.Trigger
	Distribution chan action.Action
	Gateways     map[string]chan action.Action // This should be GatewayChannel.ToClient for distribute

	core *Core
}

func (c *Core) NewPipeline() *Pipeline {
	return &Pipeline{
		Raw:          make(chan trigger.Trigger),
		Distribution: make(chan action.Action),
		Gateways:     make(map[string]chan action.Action),
		core:         c,
	}
}

func (p *Pipeline) Register(scheme string, outputChan chan action.Action) {
	p.Gateways[scheme] = outputChan
}

// Listen and process.
func (p *Pipeline) Listen() {
	for trig := range p.Raw {
		switch t := trig.(type) {
		case *trigger.Message:
			dispatch := p.core.Process(t)
			p.Distribution <- dispatch
		}
	}
}

func (p *Pipeline) DistributionListen() {
	for trig := range p.Distribution {
		switch t := trig.(type) {
		case *action.Message:
			scheme := t.Address.Scheme()
			outputChan := p.Gateways[scheme]
			outputChan <- t
		}
	}
}
