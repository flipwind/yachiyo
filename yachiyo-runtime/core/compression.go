package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-util/yerror"
)

func (c *Core) checkTokenThreshold(percentage float64, deltaLength int) (estimate bool) {
	const MAXTOKEN = 500 * 1000
	token := prompt.EstimateTokens(c.opts.SystemPrompt, c.conv.Window(), deltaLength)

	return token > int(math.Round(MAXTOKEN*percentage/100))
}

func summarize(ctx context.Context, client llm.Client, messages []llm.Message) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	var lastErr error
	for range 3 {
		result, err := client.Generate(ctx, llm.Request{
			Messages: messages,
			JSON:     false,
		})

		if err != nil {
			lastErr = err
			ylog.Error("LLM compressing error, retrying: %v", err)
			continue
		}

		if result.FinishReason != llm.FinishStop {
			lastErr = fmt.Errorf("finish_reason of LLM generation is not `stop`.")
			ylog.Error("LLM compressing error, retrying: %v", lastErr)
			continue
		}

		if strings.TrimSpace(result.Content) == "" || !strings.ContainsAny(result.Content, "#*-~") {
			lastErr = fmt.Errorf("LLM returns an empty or non-markdown message.")
			ylog.Error("LLM compressing error, retrying: %v", lastErr)
			continue
		}

		// If successfully generated, break.
		return result.Content, nil
	}

	err := fmt.Errorf("LLM failed to compress after 3 attempts: %w", errors.Join(yerror.ErrLLMGenerate, lastErr))

	return "", err
}
