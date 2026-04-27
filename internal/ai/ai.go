package ai

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type AIService struct {
	llmClient *openai.Client
	model     string
}

func NewAIService(provider, baseURL, apiKey, model string) *AIService {
	if provider == "ollama" {
		if !strings.HasSuffix(baseURL, "/v1") {
			baseURL = strings.TrimSuffix(baseURL, "/") + "/v1"
		}
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &AIService{
		llmClient: openai.NewClientWithConfig(config),
		model:     model,
	}
}

type ChatMessage struct {
	Role    string
	Content string
}

type StreamEvent struct {
	Type             string
	Tool             string
	Content          string
	PromptTokens     int
	CompletionTokens int
}

type StreamCallback func(StreamEvent)

type AskResponse struct {
	Answer           string
	Model            string
	PromptTokens     int
	CompletionTokens int
}

func (s *AIService) AskStream(ctx context.Context, question string, history []ChatMessage, cb StreamCallback) error {
	agent := NewAgent(s.llmClient, s.model)
	return agent.AskStream(ctx, question, history, cb)
}

func (s *AIService) Ask(ctx context.Context, question string, history []ChatMessage) (*AskResponse, error) {
	agent := NewAgent(s.llmClient, s.model)
	return agent.Ask(ctx, question, history)
}
