package web

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rodrigonpaiva/fclx/chatservice/internal/usecase/chatcompletionstream"
)

type ChatGPTHandler struct {
	CompletionUseCase chatcompletionstream.ChatCompletionUseCase
	Config           chatcompletionstream.ChatCompletionConfigInputDTO
	AuthToken        string
}

func NewChatGPTHandler(
	usecase chatcompletionstream.ChatCompletionUseCase,
	config chatcompletionstream.ChatCompletionConfigInputDTO,
	authToken string,
) *ChatGPTHandler {
	return &ChatGPTHandler{
		CompletionUseCase: usecase,
		Config:           config,
		AuthToken:        authToken,
	}
}

func (h *ChatGPTHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Authorization") != h.AuthToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !json.Valid(body) {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var dot chatcompletionstream.ChatCompletionInputDTO
	err = json.Unmarshal(body, &dot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dot.Config = h.Config

	result, err := h.CompletionUseCase.Execute(r.Context(), dot)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(result)
}
