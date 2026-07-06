package core

import (
	"yachiyo/yachiyo-runtime/trigger"
)

type Pipeline struct {
	Raw  chan *trigger.Message
	Distribution chan *trigger.Message
	Gateways map[string]chan trigger.Message	// This should be GatewayChannel.ToClient for distribute

	core *Core
}

func (c *Core) NewPipeline() *Pipeline {
	return &Pipeline{
		Raw:  make(chan *trigger.Message),
		Distribution: make(chan *trigger.Message),
		Gateways: make(map[string]chan trigger.Message),
		core: c,
	}
}

func (p *Pipeline) Register(scheme string, outputChan *chan trigger.Message) {
	p.Gateways[scheme] = *outputChan
}

// Listen and process.
func (p *Pipeline) Listen() {
	for e := range p.Raw {
		dispatch := p.core.Process(e)
		p.Distribution <- dispatch
	}
}

func (p *Pipeline) DistributionListen() {
	for e := range p.Distribution {
		scheme := e.Address.Scheme()
		outputChan := p.Gateways[scheme]
		outputChan <- *e
	}
}
