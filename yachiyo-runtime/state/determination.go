package state

type Determination struct {
	NewSession   bool
	WaitForReply bool
}

func NewDetermination() Determination {
	return Determination{
		NewSession:   false,
		WaitForReply: false,
	}
}
