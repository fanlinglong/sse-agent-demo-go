package controller

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/sse-agent-demo-go/internal/agent"
	"github.com/example/sse-agent-demo-go/internal/memory"
)

type MultiAgentController struct {
	routerAgent *agent.RouterAgent
	orderAgent  *agent.OrderAgent
	ragAgent    *agent.RagAgent
	memory      *memory.ConversationMemory
}

func NewMultiAgentController(router *agent.RouterAgent, order *agent.OrderAgent, rag *agent.RagAgent, memoryStore *memory.ConversationMemory) *MultiAgentController {
	return &MultiAgentController{
		routerAgent: router,
		orderAgent:  order,
		ragAgent:    rag,
		memory:      memoryStore,
	}
}

func (c *MultiAgentController) RegisterRoutes(r *gin.Engine) {
	r.POST("/multi-agent/chat", c.Chat)
}

func (c *MultiAgentController) Chat(ctx *gin.Context) {
	sessionID := ctx.DefaultQuery("sessionId", "default")
	payload, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.String(http.StatusBadRequest, "读取请求失败: %v", err)
		return
	}
	userInput := strings.TrimSpace(string(payload))
	if userInput == "" {
		ctx.String(http.StatusBadRequest, "用户输入不能为空")
		return
	}

	history := c.memory.GetSummary(sessionID)
	routing, err := c.routerAgent.Route(ctx, userInput, history)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "路由失败: %v", err)
		return
	}
	agentType := c.extractAgentType(routing)
	c.memory.AddMessage(sessionID, memory.Message{Role: "user", Content: userInput})

	response, err := c.handleWithRetry(ctx, agentType, userInput, history)
	if err != nil {
		ctx.String(http.StatusInternalServerError, "处理失败: %v", err)
		return
	}

	c.memory.AddMessage(sessionID, memory.Message{Role: "assistant", Content: response})

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", sseEncode(response))
	ctx.Writer.Flush()
}

func (c *MultiAgentController) handleWithRetry(ctx *gin.Context, agentType, userInput, history string) (string, error) {
	response, err := c.handleAgent(ctx, agentType, userInput, history)
	if err == nil && strings.TrimSpace(response) != "" {
		return response, nil
	}

	fallbackResponse, retryErr := c.handleAgent(ctx, agentType, userInput, history)
	if retryErr == nil && strings.TrimSpace(fallbackResponse) != "" {
		return fallbackResponse, nil
	}
	if err != nil {
		return "", err
	}
	return "抱歉，无法处理该问题", nil
}

func (c *MultiAgentController) handleAgent(ctx *gin.Context, agentType, userInput, history string) (string, error) {
	switch agentType {
	case "order":
		return c.orderAgent.Handle(ctx, userInput, history)
	case "rag":
		return c.ragAgent.Handle(ctx, userInput, history)
	default:
		return "抱歉，无法处理该问题", nil
	}
}

func (c *MultiAgentController) extractAgentType(routing string) string {
	routing = strings.TrimSpace(routing)
	if strings.Contains(routing, "\"agent\":\"order\"") || strings.Contains(strings.ToLower(routing), "order") {
		return "order"
	}
	if strings.Contains(routing, "\"agent\":\"rag\"") || strings.Contains(strings.ToLower(routing), "rag") {
		return "rag"
	}
	return "order"
}

func sseEncode(value string) string {
	escaped := strings.ReplaceAll(value, "\n", "\ndata: ")
	if strings.HasSuffix(escaped, "\n") {
		escaped = strings.TrimSuffix(escaped, "\n")
	}
	return strings.TrimRight(escaped, "\n")
}
