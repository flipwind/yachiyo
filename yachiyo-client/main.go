package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-utils/logger"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var sourcename = "Yachiyo.Client"

func main(){
	globalMode := os.Getenv("YACHIYO_GLOBAL_MODE")

	logger.Warn(sourcename, "Since some assets are private, this project may be unavailable at present.")
	logger.Success(sourcename, "Hello, Yachiyo!")
	logger.Info(sourcename, "Running Yachiyo Client in %s mode.", globalMode)

	conn, err := grpc.NewClient(":16800", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error(sourcename, "Failed to connect: %v", err)
		return
	}
	defer conn.Close()
	client := chat.NewChatServiceClient(conn)

	req := &chat.ChatRequest{
		Id: uuid.New().String(),
		Model: "deepseek-v4-flash",
		Messages: []*chat.Message{
			{
				Role: "user",
				Content: "晚上好~",
				Timestamp: time.Now().Unix(),
			},
		},
	}

	stream, err := client.GetChatStream(context.Background(), req)
	if err != nil {
		logger.Error(sourcename, "Creating stream Failed: %v", err)
		return
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			fmt.Println()
			break
		}

		if err != nil {
			logger.Error(sourcename, "Streaming Failed: %v", err)
			break
		}

		fmt.Print(response.Delta.Content)
	}
}