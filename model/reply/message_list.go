package reply

// MessageListResponse 消息列表响应（带分页）
type MessageListResponse struct {
	List []ReplyMessage `json:"list"` // 消息列表
}
