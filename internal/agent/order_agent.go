package agent

import (
	"encoding/json"
	"fmt"

	"github.com/example/sse-agent-demo-go/internal/tools"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type OrderAgent struct {
	client          *openai.Client
	model           string
	toolDefinitions []openai.Tool
	toolset         *tools.Toolset
}

type orderFunctionRequest struct {
	OrderID      string `json:"orderId"`
	OrderIDSnake string `json:"order_id"`
}

type logisticsFunctionRequest struct {
	OrderID      string `json:"orderId"`
	OrderIDSnake string `json:"order_id"`
}

type notificationFunctionRequest struct {
	PhoneNumber      string `json:"phoneNumber"`
	PhoneNumberSnake string `json:"phone_number"`
	Message          string `json:"message"`
}

type returnApplyFunctionRequest struct {
	OrderID      string `json:"orderId"`
	OrderIDSnake string `json:"order_id"`
	Reason       string `json:"reason"`
}

func NewOrderAgent(client *openai.Client, model string, orderTool *tools.OrderQueryTool, logisticsTool *tools.LogisticsQueryTool, notificationTool *tools.NotificationTool, returnApplyTool *tools.ReturnApplyTool) *OrderAgent {
	functionDefinitions := []openai.FunctionDefinition{
		{
			Name:        "query_order",
			Description: "根据订单号查询订单状态",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{"orderId": map[string]any{"type": "string"}},
				"required":   []string{"orderId"},
			},
		},
		{
			Name:        "query_logistics",
			Description: "查询订单的物流轨迹详情",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{"orderId": map[string]any{"type": "string"}},
				"required":   []string{"orderId"},
			},
		},
		{
			Name:        "send_notification",
			Description: "向用户发送短信通知",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"phoneNumber": map[string]any{"type": "string"},
					"message":     map[string]any{"type": "string"},
				},
				"required": []string{"phoneNumber", "message"},
			},
		},
		{
			Name:        "apply_return",
			Description: "为用户创建退货申请，需要订单号和退货原因",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"orderId": map[string]any{"type": "string"},
					"reason":  map[string]any{"type": "string"},
				},
				"required": []string{"orderId", "reason"},
			},
		},
	}

	toolDefinitions := make([]openai.Tool, 0, len(functionDefinitions))
	for i := range functionDefinitions {
		toolDefinitions = append(toolDefinitions, openai.Tool{Type: openai.ToolTypeFunction, Function: &functionDefinitions[i]})
	}

	return &OrderAgent{
		client:          client,
		model:           model,
		toolDefinitions: toolDefinitions,
		toolset:         tools.NewToolset(orderTool, logisticsTool, notificationTool, returnApplyTool),
	}
}

func (o *OrderAgent) Handle(ctx *gin.Context, userInput, history string) (string, error) {
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: `你是订单处理Agent。你有以下工具：
- query_order: 查询订单状态
- query_logistics: 查询物流信息
- send_notification: 发送短信通知
- apply_return: 创建退货申请

请首先判断是否需要调用工具，如果需要则生成结构化函数调用，否则直接给出详细回答。`},
		{Role: openai.ChatMessageRoleUser, Content: fmt.Sprintf("历史：%s\n问题：%s", history, userInput)},
	}

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:      o.model,
		Messages:   messages,
		Tools:      o.toolDefinitions,
		ToolChoice: "auto",
	})
	if err != nil {
		return "", err
	}

	choice := resp.Choices[0]
	functionCall, toolCallID := o.extractFunctionCall(choice)
	if functionCall != nil {
		for attempt := 0; attempt < 3; attempt++ {
			toolResult, err := o.invokeFunction(functionCall)
			if err != nil {
				return "", err
			}

			messages = append(messages, openai.ChatCompletionMessage{
				Role:      openai.ChatMessageRoleAssistant,
				Content:   choice.Message.Content,
				ToolCalls: choice.Message.ToolCalls,
			})
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: toolCallID,
				Content:    toolResult,
			})

			nextResp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:    o.model,
				Messages: messages,
			})
			if err != nil {
				return "", err
			}

			choice = nextResp.Choices[0]
			functionCall, toolCallID = o.extractFunctionCall(choice)
			if functionCall == nil {
				return choice.Message.Content, nil
			}
		}
		return "", fmt.Errorf("函数调用未返回文本响应")
	}

	return choice.Message.Content, nil
}

func (o *OrderAgent) extractFunctionCall(choice openai.ChatCompletionChoice) (*openai.FunctionCall, string) {
	if len(choice.Message.ToolCalls) > 0 {
		toolCall := choice.Message.ToolCalls[0]
		return &toolCall.Function, toolCall.ID
	}
	if choice.Message.FunctionCall != nil {
		return choice.Message.FunctionCall, ""
	}
	return nil, ""
}

func (o *OrderAgent) invokeFunction(call *openai.FunctionCall) (string, error) {
	switch call.Name {
	case "query_order":
		var request orderFunctionRequest
		if err := json.Unmarshal([]byte(call.Arguments), &request); err != nil {
			return "", err
		}
		orderID := request.OrderID
		if orderID == "" {
			orderID = request.OrderIDSnake
		}
		return o.toolset.OrderQueryTool.QueryOrder(orderID), nil
	case "query_logistics":
		var request logisticsFunctionRequest
		if err := json.Unmarshal([]byte(call.Arguments), &request); err != nil {
			return "", err
		}
		orderID := request.OrderID
		if orderID == "" {
			orderID = request.OrderIDSnake
		}
		return o.toolset.LogisticsQueryTool.QueryLogistics(orderID), nil
	case "send_notification":
		var request notificationFunctionRequest
		if err := json.Unmarshal([]byte(call.Arguments), &request); err != nil {
			return "", err
		}
		phoneNumber := request.PhoneNumber
		if phoneNumber == "" {
			phoneNumber = request.PhoneNumberSnake
		}
		return o.toolset.NotificationTool.SendNotification(phoneNumber, request.Message), nil
	case "apply_return":
		var request returnApplyFunctionRequest
		if err := json.Unmarshal([]byte(call.Arguments), &request); err != nil {
			return "", err
		}
		orderID := request.OrderID
		if orderID == "" {
			orderID = request.OrderIDSnake
		}
		return o.toolset.ReturnApplyTool.ApplyReturn(orderID, request.Reason), nil
	default:
		return "", fmt.Errorf("未知函数: %s", call.Name)
	}
}
