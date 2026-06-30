package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sashabaranov/go-openai"

	"github.com/example/sse-agent-demo-go/internal/agent"
	"github.com/example/sse-agent-demo-go/internal/config"
	"github.com/example/sse-agent-demo-go/internal/controller"
	"github.com/example/sse-agent-demo-go/internal/memory"
	"github.com/example/sse-agent-demo-go/internal/rag"
	"github.com/example/sse-agent-demo-go/internal/tools"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	clientConfig := openai.DefaultConfig(cfg.OpenAIKey)
	clientConfig.BaseURL = cfg.OpenAIBaseURL
	openaiClient := openai.NewClientWithConfig(clientConfig)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer pool.Close()

	store, err := rag.NewVectorStoreService(context.Background(), pool, openaiClient, cfg.OpenAIEmbeddingModel)
	if err != nil {
		log.Fatalf("初始化向量存储失败: %v", err)
	}

	memoryStore := memory.NewConversationMemory(cfg.SessionHistorySize)
	orderTool := tools.NewOrderQueryTool()
	logisticsTool := tools.NewLogisticsQueryTool()
	notificationTool := tools.NewNotificationTool()
	returnApplyTool := tools.NewReturnApplyTool()

	routerAgent := agent.NewRouterAgent(openaiClient, cfg.OpenAIModel)
	orderAgent := agent.NewOrderAgent(openaiClient, cfg.OpenAIModel, orderTool, logisticsTool, notificationTool, returnApplyTool)
	ragAgent := agent.NewRagAgent(openaiClient, cfg.OpenAIModel, store)

	agentController := controller.NewMultiAgentController(routerAgent, orderAgent, ragAgent, memoryStore)

	//gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	agentController.RegisterRoutes(engine)

	log.Printf("服务启动端口: %s", cfg.Port)
	if err := engine.Run(fmt.Sprintf(":%s", cfg.Port)); err != nil {
		log.Fatalf("Gin 启动失败: %v", err)
	}
}
