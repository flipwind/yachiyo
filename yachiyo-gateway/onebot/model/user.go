package onebotModel

type User struct {
	UserID int64	`json:"user_id"`
	Nickname string	`json:"nickname"`
	Card string		`json:"card"`
	Sex string		`json:"sex"`
	Age int64		`json:"age"`
	Area string		`json:"area"`
	Level string	`json:"level"`
	Role string		`json:"role"`
	Title string	`json:"title"`
}