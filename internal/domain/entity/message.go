package entity

import (
	"time"

	"errors"

	"github.com/google/uuid"

	tiktoken "github.com/j178/tiktoken-go"
)

type Message struct {
	ID        string
	Role      string
	Content   string
	Tokens    int
	Model     *Model
	CreatedAt time.Time
}

func NewMessage(role, content string, model *Model) (*Message, error) {
	codec, err := tiktoken.ForModel(model.GetModelName())
	if err != nil {
		return nil, err
	}
	totalTokens, err := codec.Count(content)
	if err != nil {
		return nil, err
	}
	msg := &Message{
		ID:        uuid.New().String(),
		Role:      role,
		Content:   content,
		Tokens:    totalTokens,
		Model:     model,
		CreatedAt: time.Now(),
	}
	if err := msg.Validate(); err != nil {
		return nil, err
	}
	return msg, nil
}

func (m *Message) Validate() error {
	if m.Role != "user" && m.Role != "system" && m.Role != "assistant" {
		return errors.New("invalid role")
	}
	if m.Content == "" {
		return errors.New("content is empty")
	}
	if m.CreatedAt.IsZero() {
		return errors.New("invalid creation at")
	}
	return nil
}

func (m *Message) GetQtdTokens() int {
	return m.Tokens
}
