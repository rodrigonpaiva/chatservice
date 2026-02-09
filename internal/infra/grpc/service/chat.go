package service

import (
	"github.com/rodrigonpaiva/fclx/chatservice/internal/infra/grpc/pb"
	"github.com/rodrigonpaiva/fclx/chatservice/internal/usecase/chatcompletionstream"
)

type ChatService struct {
	pb.UnimplementedChatServiceServer
	ChatCompletionStreamUseCase chatcompletionstream.ChatCompletionUseCase
	ChatConfigStream            chatcompletionstream.ChatCompletionConfigInputDTO
	StreamChannel               chan chatcompletionstream.ChatCompletionOutputDTO
}

func NewChatService(chatCompletionStreamUseCase chatcompletionstream.ChatCompletionUseCase, chatConfigStream chatcompletionstream.ChatCompletionConfigInputDTO, streamChannel chan chatcompletionstream.ChatCompletionOutputDTO) *ChatService {
	return &ChatService{
		ChatCompletionStreamUseCase: chatCompletionStreamUseCase,
		ChatConfigStream:            chatConfigStream,
		StreamChannel:               streamChannel,
	}
}

func (s *ChatService) ChatStream(req *pb.ChatRequest, stream pb.ChatService_ChatStreamServer) error {
	input := chatcompletionstream.ChatCompletionInputDTO{
		ChatID:  req.GetChatId(),
		UserID:  req.GetUserId(),
		Message: req.GetUserMessage(),
		Config:  s.ChatConfigStream,
	}

	ctx := stream.Context()

	go func() {
		for msg := range s.StreamChannel {
			stream.Send(&pb.ChatResponse{
				ChatId:  msg.ChatID,
				UserId:  msg.UserID,
				Content: msg.Content,
			})
		}
	}()

	_, err := s.ChatCompletionStreamUseCase.Execute(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
