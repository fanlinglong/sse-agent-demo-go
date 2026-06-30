package agent

import (
	"fmt"

	"github.com/example/sse-agent-demo-go/internal/rag"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type RagAgent struct {
	client *openai.Client
	model  string
	store  *rag.VectorStoreService
}

func NewRagAgent(client *openai.Client, model string, store *rag.VectorStoreService) *RagAgent {
	return &RagAgent{client: client, model: model, store: store}
}

func (r *RagAgent) Handle(ctx *gin.Context, userInput, history string) (string, error) {
	knowledge, err := r.store.Search(ctx, userInput)
	if err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(`你是知识问答Agent。基于以下知识回答问题。如果知识不足以回答，请如实告知。

相关知识：
%s`, knowledge)
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: prompt},
		{Role: openai.ChatMessageRoleUser, Content: fmt.Sprintf("历史：%s\n问题：%s", history, userInput)},
	}

	resp, err := r.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    r.model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
