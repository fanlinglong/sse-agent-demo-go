package agent

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type RouterAgent struct {
	client *openai.Client
	model  string
}

func NewRouterAgent(client *openai.Client, model string) *RouterAgent {
	return &RouterAgent{client: client, model: model}
}

func (r *RouterAgent) Route(ctx *gin.Context, userInput, history string) (string, error) {
	prompt := fmt.Sprintf(`你是路由Agent。根据用户问题和历史，判断应该交给哪个子Agent处理。你必须只输出一个 JSON 对象，不能输出任何多余文字或解释。格式：
{"agent":"order","reason":"原因"} 或 {"agent":"rag","reason":"原因"}

- order: 涉及订单查询、物流、退货等具体订单操作
- rag: 咨询政策、规则、常见问题等知识性问题

历史对话：%s
当前问题：%s`, history, userInput)

	resp, err := r.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: r.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "{\"agent\":\"order\",\"reason\":\"默认路由\"}", nil
	}
	return resp.Choices[0].Message.Content, nil
}
