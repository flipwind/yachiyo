package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-core/prompt"
	"yachiyo/yachiyo-core/storage"
	"yachiyo/yachiyo-server/config"
	"yachiyo/yachiyo-server/llmprovider"
	"yachiyo/yachiyo-server/plugin"
	"yachiyo/yachiyo-utils/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

var sourcename string = "Yachiyo.Server"

var (
	globalProvider *llmprovider.OpenAIProvider
	systemPrompt   prompt.SystemPrompt
	globalStorage  storage.MemoryStorage

	globalMode string
	charPath   string
)

type ChatServer struct {
	chat.UnimplementedChatServiceServer
}

func (s *ChatServer) CreateSession(ctx context.Context, req *chat.CreateSessionRequest) (*chat.CreateSessionResponse, error) {
	session_id := uuid.NewString()
	timestamp := time.Now().Unix()

	systemPrompt = prompt.LoadSystemPrompt(charPath)
	systemPrompt = prompt.ProcessSystemPrompt(systemPrompt, "cli")

	message := chat.BuildMessage(session_id, "system", systemPrompt.Content, timestamp)

	globalStorage.SaveMessage(ctx, message)

	return &chat.CreateSessionResponse{
		Uuid: session_id,
	}, nil
}

func (s *ChatServer) ChatStream(req *chat.ChatRequest, stream chat.ChatService_ChatStreamServer) error {
	logger.Info(sourcename, "Received %v", req.SessionId)
	session_id := req.SessionId
	timestamp := time.Now().Unix()

	var messages []*chat.Message

	history, err := globalStorage.GetHistory(stream.Context(), session_id, -1)
	if err != nil {
		logger.Error(sourcename, "Getting history failed: %v", err)
		return err
	}

	for _, msg := range history {
		messages = append(messages, msg)
	}

	user_message := chat.BuildMessage(session_id, "user", req.Content, timestamp)

	logger.Debug(sourcename, "%s", chat.ContentLoggerFormat(user_message))

	messages = append(messages, user_message)

	output, err := globalProvider.ChatStream(stream.Context(), messages)
	if err != nil {
		logger.Error(sourcename, "Chatstreaming Failed: %v", err)
		return err
	}

	response := ""

	for word := range output {
		response += word
		err := stream.Send(&chat.ChatResponse{
			Id: session_id,
			Delta: &chat.Message{
				Role:    "assistant",
				Content: word,
			},
		})
		if err != nil {
			logger.Error(sourcename, "Stream Failed: %v", err)
			return err
		}
	}

	assistant_message := chat.BuildMessage(session_id, "assistant", response, timestamp)
	logger.Debug(sourcename, "%s", chat.ContentLoggerFormat(assistant_message))

	timestamp = time.Now().Unix()
	err = globalStorage.SaveMessage(stream.Context(), user_message)
	if err != nil {
		logger.Error(sourcename, "Saving user message error: %v", err)
	}

	err = globalStorage.SaveMessage(stream.Context(), assistant_message)
	if err != nil {
		logger.Error(sourcename, "Saving assistant message error: %v", err)
	}

	return nil
}

func main() {
	// Since this project is under active development, you may find this unavailable.
	// Some assets like prompts are currently private.

	globalMode = os.Getenv("YACHIYO_GLOBAL_MODE")
	charPath = os.Getenv("YACHIYO_CHAR_PATH")

	logger.Warn(sourcename, "Since some assets are private, this project may be unavailable at present.")
	logger.Success(sourcename, "Hello, Yachiyo!")
	logger.Info(sourcename, "Running Yachiyo Server in %s mode.", globalMode)
	logger.Info(sourcename, "Now character package reading the path {%s}", charPath)

	// Config
	configManager := config.NewConfigManager()
	err := configManager.Load("config.yaml")
	if err != nil {
		logger.Error(sourcename, "Reading Config Failed.")
		return
	}

	cfg := configManager.CurrentConfig
	currentKey := cfg.LLM.Key[0]
	globalProvider = llmprovider.CreateOpenAIProvider(currentKey.BaseUrl, currentKey.Secret, currentKey.ModelName)

	// Storage
	globalStorage, err = storage.NewMemoryStorage("sqlite3", "memory.db")
	if err != nil {
		logger.Error(sourcename, "Storage creation failed: %v", err)
	}
	defer globalStorage.Close()

	// Service
	port := cfg.Runtime.Port

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error(sourcename, "Listening port failed: %v", err)
		return
	}

	grpcServer := grpc.NewServer()
	chat.RegisterChatServiceServer(grpcServer, &ChatServer{})
	logger.Success(sourcename, "Yachiyo Server listening at port %v", port)

	// Plugins
	pluginDriver := plugin.NewPluginDriver(cfg.Plugin.Path, fmt.Sprintf(":%v", port))
	pluginDriver.Init()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error(sourcename, "Failed to serve: %v", err)
			return
		}
	}()

	<-ctx.Done()
	logger.Info(sourcename, "Yachiyo logging out...")

	pluginDriver.Close()
	grpcServer.GracefulStop()

	logger.Success(sourcename, "Bye Yachiyo ^-^")
}
