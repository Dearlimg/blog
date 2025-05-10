package reply

import "time"

type ReplyMessage struct {
	Id        int32     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Content   string    `json:"content"`
	Create_at time.Time `json:"create_at"`
}
