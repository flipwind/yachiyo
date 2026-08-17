package core

import (
	"time"
	"yachiyo/yachiyo-runtime/action"
	"yachiyo/yachiyo-runtime/trigger"
)

type Pipeline struct {
	Raw          chan trigger.Trigger
	Distribution chan action.Action
	Gateways     map[string]chan action.Action // This should be GatewayChannel.ToClient for distribute

	handler func(trigger.Trigger) action.Action
}

func NewPipeline(handler func(trigger.Trigger) action.Action) *Pipeline {
	return &Pipeline{
		// channel length change to 65535
		// This is a test on purpose, a better solution is on the way.
		// TODO: a better process solution
		Raw:          make(chan trigger.Trigger, 65535),
		Distribution: make(chan action.Action, 65535),
		Gateways:     make(map[string]chan action.Action),
		handler:      handler,
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
			dispatch := p.handler(t)
			p.Distribution <- dispatch
		case *trigger.InitiativeMessage:
			dispatch := p.handler(t)
			p.Distribution <- dispatch
		}
	}
}

func (p *Pipeline) DistributionListen() {
	for trig := range p.Distribution {
		ylog.Debug("Receiving trigger %#v", trig)
		switch t := trig.(type) {
		case *action.Message:
			scheme := t.Address.Scheme()
			outputChan := p.Gateways[scheme]
			if outputChan == nil {
				ylog.Error("scheme <%v> is not registered", scheme)
			}
			select {
			case outputChan <- t:
			case <-time.After(time.Second):
				ylog.Error("gateway %s timeout", scheme)
			}
		case *action.RuntimeState:
			scheme := t.Address.Scheme()
			outputChan := p.Gateways[scheme]
			if outputChan == nil {
				ylog.Error("scheme <%v> is not registered", scheme)
			}
			select {
			case outputChan <- t:
			case <-time.After(time.Second):
				ylog.Error("gateway %s timeout", scheme)
			}
		}
	}
}
