package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	configs "github.com/rodrigonpaiva/fclx/chatservice/config"
	"github.com/rodrigonpaiva/fclx/chatservice/internal/infra/repository"
	"github.com/rodrigonpaiva/fclx/chatservice/internal/infra/web"
	"github.com/rodrigonpaiva/fclx/chatservice/internal/infra/web/webserver"
	"github.com/rodrigonpaiva/fclx/chatservice/internal/usecase/chatcompletionstream"
	"github.com/sashabaranov/go-openai"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(fmt.Errorf("failed to load configs: %w", err))
	}

	conn, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	repo := repository.NewChatRepositoryMySQL(conn)
	client := openai.NewClient(configs.OpenAIApiKey)

	chatConfig := chatcompletionstream.ChatCompletionConfigInputDTO{
		Model:                configs.Model,
		ModelMaxTokens:       configs.ModelMaxTokens,
		Temperature:          float32(configs.Temperature),
		TopP:                 float32(configs.TopP),
		N:                    configs.N,
		Stop:                 configs.Stop,
		MaxTokens:            configs.MaxTokens,
		InitialSystemMessage: configs.InitialChatMessage,
	}
	streamChannel := make(chan chatcompletionstream.ChatCompletionOutputDTO)

	// Consume stream channel to avoid deadlock
	go func() {
		for range streamChannel {
			// Messages are consumed here
			// In a real application, you could log or process streaming updates
		}
	}()

	usecaseStream := chatcompletionstream.NewChatCompletionUseCase(repo, client, streamChannel)

	webServer := webserver.NewWebServer(":" + configs.WebServerPort)
	webServerChatHandler := web.NewChatGPTHandler(*usecaseStream, chatConfig, configs.AuthToken)
	webServer.AddHandler("/chat", webServerChatHandler.Handle)

	fmt.Println("Starting web server on port", configs.WebServerPort)
	if err := webServer.Start(); err != nil {
		panic(err)
	}
}
