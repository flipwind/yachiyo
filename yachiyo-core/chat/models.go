package chat

type Message struct{
	// Role including system, user, assistant
	// TODO: tool
	Role string
	Content string
}