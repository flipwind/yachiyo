package main

import (
	"net"
	"os"
	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-core/prompt"
	"yachiyo/yachiyo-server/config"
	"yachiyo/yachiyo-server/llmprovider"
	"yachiyo/yachiyo-utils/logger"

	"google.golang.org/grpc"
)

var sourcename string = "Yachiyo.Server"

var (
	globalProvider *llmprovider.OpenAIProvider
	systemPrompt prompt.SystemPrompt
)

type ChatServer struct {
	chat.UnimplementedChatServiceServer
}

func (s *ChatServer) GetChatStream(req *chat.ChatRequest, stream chat.ChatService_GetChatStreamServer) error {
	logger.Info(sourcename, "Received %v, %v", req.Id, req.Messages)

	var messages []chat.Message
	messages = append(messages, chat.Message{
		Role: "system",
		Content: systemPrompt.Content,
	})

	for _, msg := range req.Messages {
		messages = append(messages, chat.Message{
			Role: msg.Role,
			Content: msg.Content,
		})
	}
	
	output, err := globalProvider.ChatStream(stream.Context(), messages)
	if err != nil {
		logger.Error(sourcename, "Chatstreaming Failed: %v", err)
		return err
	}

	for word := range output {
		err := stream.Send(&chat.ChatResponse{
			Id: req.Id,
			Delta: &chat.Message{
				Role: "assistant",
				Content: word,
			},
		})
		if err != nil {
			logger.Error(sourcename, "Stream Failed: %v", err)
			return err
		}
	}

	return nil
}

func main() {
	// Since this project is under active development, you may find this unavailable.
	// Some assets like prompts are currently private.

	globalMode := os.Getenv("YACHIYO_GLOBAL_MODE")
	charPath := os.Getenv("YACHIYO_CHAR_PATH")

	logger.Warn(sourcename, "Since some assets are private, this project may be unavailable at present.")
	logger.Success(sourcename, "Hello, Yachiyo!")
	logger.Info(sourcename, "Running Yachiyo Server in %s mode.", globalMode)
	logger.Info(sourcename, "Now character package reading the path {%s}", charPath)

	systemPrompt = prompt.LoadSystemPrompt(charPath)
	systemPrompt = prompt.ProcessSystemPrompt(systemPrompt, "cli")

	configManager := config.NewConfigManager()
	err := configManager.Load("config.yaml")
	if err != nil {
		logger.Error(sourcename, "Reading Config Failed.")
		return
	}

	cfg := configManager.CurrentConfig
	currentKey := cfg.LLM.Key[0]
	globalProvider = llmprovider.CreateOpenAIProvider(currentKey.BaseUrl, currentKey.Secret, currentKey.ModelName)

	lis, err := net.Listen("tcp", ":16800")
	if err != nil {
		logger.Error(sourcename, "Listening port failed: %v", err)
		return
	}

	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, &ChatServer{})
	logger.Success(sourcename, "Yachiyo Server listening at port 16800")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Error(sourcename, "Failed to serve: %v", err)
		return
	}
}
