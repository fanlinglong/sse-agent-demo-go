package memory

import (
	"fmt"
	"strings"
	"sync"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ConversationMemory struct {
	mu          sync.RWMutex
	sessions    map[string][]Message
	historySize int
}

func NewConversationMemory(historySize int) *ConversationMemory {
	return &ConversationMemory{
		sessions:    make(map[string][]Message),
		historySize: historySize,
	}
}

func (c *ConversationMemory) AddMessage(sessionID string, message Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessions[sessionID] = append(c.sessions[sessionID], message)
}

func (c *ConversationMemory) GetHistory(sessionID string) []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]Message(nil), c.sessions[sessionID]...)
}

func (c *ConversationMemory) GetSummary(sessionID string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	messages := c.sessions[sessionID]
	if len(messages) == 0 {
		return ""
	}

	start := len(messages) - c.historySize*2
	if start < 0 {
		start = 0
	}
	summaryMessages := messages[start:]

	lines := make([]string, 0, len(summaryMessages))
	for _, msg := range summaryMessages {
		lines = append(lines, fmt.Sprintf("%s: %s", msg.Role, msg.Content))
	}
	return fmt.Sprint(strings.Join(lines, "\n"))
}
