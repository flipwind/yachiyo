package state

type Determination struct {
	NewSession   bool `json:"should_open_a_new_session"`
	WaitForReply bool `json:"should_wait_for_reply"`
}

func NewDetermination() Determination{
	return Determination{
		NewSession: false,
		WaitForReply: false,
	}
}
