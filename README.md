# SSE Agent Demo Go

基于 Spring Boot 示例的 Go 实现，使用 Gin + OpenAI + PGVector + Viper。项目目标是实现一个多 Agent 协同系统，包括路由 Agent、订单 Agent、RAG Agent、会话记忆和函数调用工具。

## 运行方式

1. 设置环境变量，参考 `.env.example`
2. 启动 PostgreSQL，并确保 `vector` 扩展已可用
3. 执行 `go run ./cmd/server`

## API

`POST /multi-agent/chat?sessionId=default`

请求体是纯文本用户输入，响应使用 `text/event-stream` SSE 流返回。
