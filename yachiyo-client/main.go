package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"yachiyo/yachiyo-core/chat"
	"yachiyo/yachiyo-utils/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var sourcename = "Yachiyo.Client"

func main() {
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

	session_response, err := client.CreateSession(context.Background(), &chat.CreateSessionRequest{})
	session_id := session_response.GetUuid()

	if err != nil {
		logger.Error(sourcename, "Getting session ID failed: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nUser > ")

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		input = strings.TrimSpace(input)

		req := &chat.ChatRequest{
			SessionId:      session_id,
			Model:   "deepseek-v4-flash",
			Content: input,
		}

		stream, err := client.ChatStream(context.Background(), req)
		if err != nil {
			logger.Error(sourcename, "Creating stream failed: %v", err)
			return
		}

		fmt.Print("\nYachiyo > ")

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

	if err := scanner.Err(); err != nil {
		logger.Error(sourcename, "Scanned error: %v", err)
	}
}
