package core

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"uuid"
	"yachiyo/yachiyo-runtime/history"
	"yachiyo/yachiyo-runtime/llm"
	"yachiyo/yachiyo-runtime/prompt"
	"yachiyo/yachiyo-util/yerror"
)

func (c *Core) checkTokenThreshold(percentage float64, deltaLength int) (estimate bool) {
	const MAXTOKEN = 500 * 1000
	var contextWindow history.ContextWindow
	contextWindow = c.getContextWindow()

	charnum := 0 + deltaLength
	charnum += len(c.getSystemPrompt())
	charnum += len(contextWindow.CurrentNode.Rend())
	for _, h := range contextWindow.CurrentHistories {
		charnum += len(prompt.Event(h))
	}

	token := charnum / 3

	return token > int(math.Round(MAXTOKEN*percentage/100))
}

func (c *Core) processCompress() error {
	clearbusy := func() {
		c.mu.Lock()
		c.CompressBusy = false
		c.mu.Unlock()
	}

	c.mu.Lock()
	if c.CompressBusy {
		c.mu.Unlock()
		return nil
	}
	c.CompressBusy = true
	c.mu.Unlock()

	defer clearbusy()

	window := c.getContextWindow()

	if len(window.CurrentHistories) == 0 {
		return fmt.Errorf("Current history list is empty. Skip compress.")
	}

	node := history.CompressNode{
		Empty:               false,
		ID:                  uuid.NewV7(),
		CompressTimestamp:   time.Now(),
		SubCompressNodeID:   window.CurrentNode.ID,
		CompressEventIDFrom: window.CurrentHistories[0].GetMeta().ID,
		CompressEventIDTo:   window.CurrentHistories[len(window.CurrentHistories)-1].GetMeta().ID,
		Content:             "NO COMPRESS RESULT",
	}

	var messages = prompt.HistoryPrompt(window)
	messages = append(messages, llm.Message{
		Role:    llm.RoleSystem,
		Content: c.getCompressPrompt(),
	})
	var lastErr error
	var successflag bool = false

	for range 3 {
		result, err := c.LLM.Generate(context.Background(), llm.Request{
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
			ylog.Error("LLM compressing error, retrying: %v", err)
			continue
		}

		if strings.TrimSpace(result.Content) == "" && !strings.ContainsAny(result.Content, "#*-~") {
			lastErr = fmt.Errorf("LLM returns a empty message.")
			ylog.Error("LLM compressing error, retrying: %v", err)
			continue
		}

		// If successfully generated, break.
		node.Content = result.Content
		successflag = true
		break
	}

	if successflag {
		if err := c.commitCompressNode(node); err != nil {
			saveErr := fmt.Errorf("Compressed but failed to save in storage: %w", err)
			ylog.Error("Compressed failed: %v", saveErr)
			return saveErr
		}
	}

	if successflag {
		ylog.Success("Compressed.")
		return nil
	}

	err := fmt.Errorf("LLM failed to compress after 3 attempts: %w", errors.Join(yerror.ErrLLMGenerate, lastErr))
	ylog.Error("Compress failed: %s", err)

	return err
}
