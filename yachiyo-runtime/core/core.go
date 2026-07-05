package core

import (
	"yachiyo/yachiyo-runtime/config"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/history/basic"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/llm/provider"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-runtime/state"
	"yachiyo/yachiyo-runtime/trigger"
	"yachiyo/yachiyo-util/logger"
)

var ylog = logger.New("Yachiyo.Core")

type Core struct {
	History       history.HistoryStorage
	State         state.State
	Emotion       state.Emotion
	LLM           llm.LLM
	Config        config.ConfigManager
	Determination state.Determination

	JSONConstraint bool
	Note           string
}

func New() *Core {
	config, err := config.LoadConfig("../yachiyo-runtime/assets/config.yaml")
	if err != nil {
	}

	// TODO: LLM List
	llmConfig := config.CurrentConfig.LLM.Key[0]
	llm := provider.NewOpenAIProvider(llmConfig.BaseUrl, llmConfig.Secret, llmConfig.ModelName)

	return &Core{
		History:        basic.New(),
		State:          state.NewState(),
		Emotion:        state.NewEmotion(),
		LLM:            llm,
		Config:         config,
		Determination:  state.NewDetermination(),
		JSONConstraint: false,
		Note:           "",
	}
}

func (c *Core) Process(e *trigger.Message) string {
	histories := prompt.PromptBuilder(&prompt.Context{
		History: c.History,
		Emotion: c.Emotion,
		State:   c.State,
		Note:    c.Note,
	}, e)

	var result, answer string
	var err error

	for i := range 3 {
		if i == 1 {
			c.JSONConstraint = true
		}

		if c.JSONConstraint {
			histories = append(histories, history.History{
				Role:    "user",
				Content: "<PROCESS HINT> YOUR LAST REPLY IS NOT A VALID JSON, REGENERATE IT. YOU MUST FOLLOW THE OUTPUT ROLE.",
			})
		}

		result, err = c.LLM.Gen(histories)

		if err != nil {
			ylog.Error("LLM Generating error: %v", err)
			continue
		}

		answer, err = c.OutputProcess(result)

		if err != nil {
			ylog.Error("%d request failed. Retrying...", i+1)
			continue
		}

		c.History.Remember(history.History{
			Role:    "assistant",
			Content: answer,
		})

		break
	}

	return answer
}
