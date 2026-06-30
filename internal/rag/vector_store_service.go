package rag

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/sashabaranov/go-openai"
)

type VectorStoreService struct {
	pool           *pgxpool.Pool
	client         *openai.Client
	embeddingModel string
}

func NewVectorStoreService(ctx context.Context, pool *pgxpool.Pool, client *openai.Client, embeddingModel string) (*VectorStoreService, error) {
	s := &VectorStoreService{pool: pool, client: client, embeddingModel: embeddingModel}
	if err := s.ensureSchema(ctx); err != nil {
		return nil, err
	}
	if err := s.seedDocuments(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *VectorStoreService) ensureSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS vector; CREATE TABLE IF NOT EXISTS documents (id TEXT PRIMARY KEY, content TEXT NOT NULL, embedding VECTOR(2048) NOT NULL);`)
	return err
}

func (s *VectorStoreService) seedDocuments(ctx context.Context) error {
	docs := []string{
		"退货政策：签收后7天内可申请无理由退货",
		"退款时效：退货审核通过后3-5个工作日到账",
		"运费说明：质量问题卖家承担运费，非质量问题买家承担",
		"物流更新：订单发货后可通过物流单号查询最新配送状态",
		"客服信息：如需帮助，可联系客服 400-100-2000",
	}

	for _, content := range docs {
		embedding, err := s.embedText(ctx, content)
		if err != nil {
			return err
		}
		_, err = s.pool.Exec(ctx, `INSERT INTO documents (id, content, embedding) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, embedding = EXCLUDED.embedding`, generateID(content), content, pgvector.NewVector(embedding))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *VectorStoreService) Search(ctx context.Context, query string) (string, error) {
	embedding, err := s.embedText(ctx, query)
	if err != nil {
		return "", err
	}

	rows, err := s.pool.Query(ctx, `SELECT content FROM documents ORDER BY embedding <-> $1 LIMIT 5`, pgvector.NewVector(embedding))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	parts := make([]string, 0)
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			return "", err
		}
		parts = append(parts, content)
	}
	return strings.Join(parts, "\n---\n"), nil
}

func (s *VectorStoreService) embedText(ctx context.Context, text string) ([]float32, error) {
	req := openai.EmbeddingRequestStrings{
		Model: openai.EmbeddingModel(s.embeddingModel),
		Input: []string{text},
	}
	resp, err := s.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("empty embedding result")
	}
	return resp.Data[0].Embedding, nil
}

func generateID(text string) string {
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("doc-%x", sum)
}
