package core

import (
	"fmt"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/prompt"
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
		// TODO: Let core process it.
		histories := prompt.PromptBuilder(&prompt.Context{
			History: p.core.History,
			Emotion: p.core.Emotion,
			State:   p.core.State,
			Note:    p.core.Note,
		}, e)

		var result, answer string
		var err error

		// <debug>
		var core_copy Core
		core_copy = *p.core
		// </debug>

		for i := range 3 {
			if i == 1 {
				p.core.JSONConstraint = true
			}

			if p.core.JSONConstraint {
				histories = append(histories, history.History{
					Role:    "user",
					Content: "<PROCESS HINT> YOUR LAST REPLY IS NOT A VALID JSON, REGENERATE IT. YOU MUST FOLLOW THE OUTPUT ROLE.",
				})
			}

			result, err = p.core.LLM.Gen(histories)

			if err != nil {
				ylog.Error("LLM Generating error: %v", err)
				continue
			}

			answer, err = p.core.OutputProcess(result)

			if err != nil {
				ylog.Error("%d request failed. Retrying...", i+1)
				continue
			}

			p.core.History.Remember(history.History{
				Role:    "assistant",
				Content: answer,
			})

			break
		}

		// <debug>
		fmt.Printf("Yachiyo > %s\n", answer)

		fmt.Printf("\nDEBUG\nFORMAL:")
		fmt.Println(core_copy.Emotion.String())
		fmt.Println(core_copy.State.Prompt())
		fmt.Printf("\nLATER:\n")
		fmt.Println(p.core.Emotion.String())
		fmt.Println(p.core.State.Prompt())
		fmt.Println(p.core.Determination)

		fmt.Println("Session Context: " + p.core.Note)
		// </debug>

		// TODO: e => PipelineResult
		e.Content = answer
		p.Distribution <- e
	}
}

func (p *Pipeline) DistributionListen() {
	for e := range p.Distribution {
		scheme := e.Address.Scheme()
		outputChan := p.Gateways[scheme]
		outputChan <- *e
	}
}
