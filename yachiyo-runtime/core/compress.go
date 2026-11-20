package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/conversation"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-util/yerror"
)

// It uses MAXTOKEN, which is the context budget,
// to decide when to compress.
// Currently the MAXTOKEN is fixed at 500K.
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

type compressJob struct{}

func (compressJob) isJob() {}

type compressed struct {
	summary conversation.Summary
	err     error
}

func (compressed) isResult() {}

func (c *Core) startCompress() {
	window := c.conv.Window()

	if len(window.Events) == 0 {
		ylog.Warn("Current history list is empty. Skip compress.")
		return
	}

	var parent uuid.UUID
	if window.Summary == nil {
		parent = uuid.Nil()
	} else {
		parent = window.Summary.ID
	}

	summary := conversation.Summary{
		ID:          uuid.NewV7(),
		CreateAt:    time.Now(),
		Parent:      parent,
		FromEventID: window.Events[0].GetMeta().ID,
		ToEventID:   window.Events[len(window.Events)-1].GetMeta().ID,
		Content:     "NO COMPRESS RESULT",
	}

	messages := prompt.BuildCompression(window, c.opts.CompressPrompt)

	c.busy = true
	go func() {
		sum, err := summarize(c.ctx, c.llm, messages)
		summary.Content = sum
		c.send(compressed{summary: summary, err: err})
	}()
}

func (c *Core) finishCompress(r compressed) {
	if r.err != nil {
		ylog.Error("Process compression failed: %s", r.err)
		return
	}

	if err := c.commitSummary(r.summary); err != nil {
		saveErr := fmt.Errorf("Compressed but failed to save in storage: %w", err)
		ylog.Error("Compressed failed: %v", saveErr)
		return
	}

	ylog.Success("Compressed.")
}
