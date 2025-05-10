package request

type ParamCreateMessage struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Content string `json:"content"`
}
