package logic

import (
	"blog/dao"
	"blog/global"
	"blog/model/entity"
	"blog/model/reply"
	"blog/model/request"
	"blog/pkg/cache"
	"blog/pkg/errcode"
	"blog/pkg/logger"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type eino struct{}

// OllamaService Ollama 服务
type OllamaService struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaService 创建 Ollama 服务实例
func NewOllamaService() *OllamaService {
	config := global.Config.Ollama
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434" // 默认值
	}

	modelName := config.Model
	if modelName == "" {
		modelName = "deepseek-r1:8b" // 默认值
	}

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second // 默认值
	}

	return &OllamaService{
		baseURL: baseURL,
		model:   modelName,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// ChatRequest Ollama 聊天请求
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message 消息结构
type Message struct {
	Role    string `json:"role"`    // "user" 或 "assistant"
	Content string `json:"content"` // 消息内容
}

// ChatResponse Ollama 聊天响应
type ChatResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// Chat 发送聊天请求
func (o *OllamaService) Chat(messages []Message) (string, error) {
	url := fmt.Sprintf("%s/api/chat", o.baseURL)

	reqBody := ChatRequest{
		Model:    o.model,
		Messages: messages,
		Stream:   false, // 非流式响应
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	logger.Debug("Sending request to Ollama",
		logger.String("url", url),
		logger.String("model", o.model),
		logger.Int("message_count", len(messages)),
	)

	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Debug("Received response from Ollama",
		logger.String("model", chatResp.Model),
		logger.Bool("done", chatResp.Done),
	)

	return chatResp.Message.Content, nil
}

// ChatStream 发送流式聊天请求（返回通道）
func (o *OllamaService) ChatStream(messages []Message) (<-chan string, <-chan error) {
	contentChan := make(chan string, 10)
	errChan := make(chan error, 1)

	go func() {
		defer close(contentChan)
		defer close(errChan)

		url := fmt.Sprintf("%s/api/chat", o.baseURL)

		reqBody := ChatRequest{
			Model:    o.model,
			Messages: messages,
			Stream:   true, // 流式响应
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			errChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			errChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := o.client.Do(req)
		if err != nil {
			errChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errChan <- fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
			return
		}

		decoder := json.NewDecoder(resp.Body)
		for {
			var chatResp ChatResponse
			if err := decoder.Decode(&chatResp); err != nil {
				if err == io.EOF {
					break
				}
				errChan <- fmt.Errorf("failed to decode response: %w", err)
				return
			}

			if chatResp.Message.Content != "" {
				contentChan <- chatResp.Message.Content
			}

			if chatResp.Done {
				break
			}
		}
	}()

	return contentChan, errChan
}

// OllamaChatModel Ollama ChatModel 实现
type OllamaChatModel struct {
	ollamaService *OllamaService
	modelName     string
}

// NewOllamaChatModel 创建 Ollama ChatModel 实例
func NewOllamaChatModel() model.BaseChatModel {
	config := global.Config.Ollama
	modelName := config.Model
	if modelName == "" {
		modelName = "deepseek-r1:8b" // 默认值
	}

	return &OllamaChatModel{
		ollamaService: GetOllamaService(),
		modelName:     modelName,
	}
}

// Generate 实现 BaseChatModel 接口
func (o *OllamaChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 转换消息格式
	ollamaMessages := make([]Message, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		if msg.Role == schema.Assistant {
			role = "assistant"
		}

		ollamaMessages = append(ollamaMessages, Message{
			Role:    role,
			Content: msg.Content, // Content 是 string 类型
		})
	}

	logger.Debug("Invoking Ollama ChatModel",
		logger.Int("message_count", len(ollamaMessages)),
	)

	// 调用 Ollama 服务
	response, err := o.ollamaService.Chat(ollamaMessages)
	if err != nil {
		return nil, fmt.Errorf("ollama chat failed: %w", err)
	}

	// 返回 schema.Message（没有 toolCalls，传 nil）
	return schema.AssistantMessage(response, nil), nil
}

// Stream 实现流式 BaseChatModel 接口
func (o *OllamaChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 转换消息格式
	ollamaMessages := make([]Message, 0, len(messages))
	for _, msg := range messages {
		role := "user"
		if msg.Role == schema.Assistant {
			role = "assistant"
		}

		ollamaMessages = append(ollamaMessages, Message{
			Role:    role,
			Content: msg.Content, // Content 是 string 类型
		})
	}

	logger.Debug("Streaming Ollama ChatModel",
		logger.Int("message_count", len(ollamaMessages)),
	)

	// 获取流式通道
	contentChan, errChan := o.ollamaService.ChatStream(ollamaMessages)

	// 创建 StreamReader 和 StreamWriter
	streamReader, streamWriter := schema.Pipe[*schema.Message](10)

	// 在 goroutine 中处理流式数据
	go func() {
		defer streamWriter.Close()

		for {
			select {
			case content, ok := <-contentChan:
				if !ok {
					return
				}
				// 发送增量消息（没有 toolCalls，传 nil）
				msg := schema.AssistantMessage(content, nil)
				closed := streamWriter.Send(msg, nil)
				if closed {
					logger.Debug("Stream writer closed")
					return
				}
			case err := <-errChan:
				if err != nil {
					logger.Error("Stream error",
						logger.ErrorField(err),
					)
				}
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return streamReader, nil
}

// GetOllamaService 获取 Ollama 服务实例（单例模式）
var ollamaService *OllamaService

func GetOllamaService() *OllamaService {
	if ollamaService == nil {
		ollamaService = NewOllamaService()
	}
	return ollamaService
}

// GetEinoChatModel 获取 Eino ChatModel 实例（单例模式）
var einoChatModel model.BaseChatModel

func GetEinoChatModel() model.BaseChatModel {
	if einoChatModel == nil {
		einoChatModel = NewOllamaChatModel()
	}
	return einoChatModel
}

// buildSystemPrompt 构建系统提示词
func buildSystemPrompt(personality, background string) string {
	var parts []string

	if personality != "" {
		parts = append(parts, fmt.Sprintf("性格设定：%s", personality))
	}

	if background != "" {
		parts = append(parts, fmt.Sprintf("背景设定：%s", background))
	}

	if len(parts) > 0 {
		return strings.Join(parts, "\n\n")
	}

	return "你是一个友好的AI助手。"
}

// CreateChatbot 创建聊天机器人
func (e *eino) CreateChatbot(ctx *gin.Context, param *request.ParamCreateChatbot) (*reply.ChatbotResponse, errcode.Err) {
	// 构建系统提示词
	systemPrompt := buildSystemPrompt(param.Personality, param.Background)

	// 创建实体
	chatbot := &entity.Chatbot{
		Name:         param.Name,
		Personality:  param.Personality,
		Background:   param.Background,
		SystemPrompt: systemPrompt,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存到数据库
	if err := dao.Database.Chatbot.CreateChatbot(ctx, chatbot); err != nil {
		logger.ErrorWithCtx(ctx, "CreateChatbot failed",
			logger.ErrorField(err),
		)
		return nil, errcode.ErrServer.WithDetails(err.Error())
	}

	logger.InfoWithCtx(ctx, "Chatbot created successfully",
		logger.String("name", param.Name),
		logger.Int32("id", chatbot.ID),
	)

	return &reply.ChatbotResponse{
		ID:           chatbot.ID,
		Name:         chatbot.Name,
		Personality:  chatbot.Personality,
		Background:   chatbot.Background,
		SystemPrompt: chatbot.SystemPrompt,
		CreatedAt:    chatbot.CreatedAt,
		UpdatedAt:    chatbot.UpdatedAt,
	}, nil
}

// GetChatbot 获取聊天机器人
func (e *eino) GetChatbot(ctx *gin.Context, id int32) (*reply.ChatbotResponse, errcode.Err) {
	chatbot, err := dao.Database.Chatbot.GetChatbot(ctx, id)
	if err != nil {
		logger.ErrorWithCtx(ctx, "GetChatbot failed",
			logger.ErrorField(err),
			logger.Int32("id", id),
		)
		return nil, errcode.ErrResourceNotFound.WithDetails("chatbot not found")
	}

	return &reply.ChatbotResponse{
		ID:           chatbot.ID,
		Name:         chatbot.Name,
		Personality:  chatbot.Personality,
		Background:   chatbot.Background,
		SystemPrompt: chatbot.SystemPrompt,
		CreatedAt:    chatbot.CreatedAt,
		UpdatedAt:    chatbot.UpdatedAt,
	}, nil
}

// ListChatbots 获取所有聊天机器人
func (e *eino) ListChatbots(ctx *gin.Context) (*reply.ChatbotListResponse, errcode.Err) {
	chatbots, err := dao.Database.Chatbot.ListChatbots(ctx)
	if err != nil {
		logger.ErrorWithCtx(ctx, "ListChatbots failed",
			logger.ErrorField(err),
		)
		return nil, errcode.ErrServer.WithDetails(err.Error())
	}

	list := make([]reply.ChatbotResponse, 0, len(chatbots))
	for _, cb := range chatbots {
		list = append(list, reply.ChatbotResponse{
			ID:           cb.ID,
			Name:         cb.Name,
			Personality:  cb.Personality,
			Background:   cb.Background,
			SystemPrompt: cb.SystemPrompt,
			CreatedAt:    cb.CreatedAt,
			UpdatedAt:    cb.UpdatedAt,
		})
	}

	return &reply.ChatbotListResponse{
		List: list,
	}, nil
}

// UpdateChatbot 更新聊天机器人
func (e *eino) UpdateChatbot(ctx *gin.Context, id int32, param *request.ParamUpdateChatbot) (*reply.ChatbotResponse, errcode.Err) {
	// 获取现有聊天机器人
	chatbot, err := dao.Database.Chatbot.GetChatbot(ctx, id)
	if err != nil {
		return nil, errcode.ErrResourceNotFound.WithDetails("chatbot not found")
	}

	// 更新字段
	if param.Name != "" {
		chatbot.Name = param.Name
	}
	if param.Personality != "" {
		chatbot.Personality = param.Personality
	}
	if param.Background != "" {
		chatbot.Background = param.Background
	}

	// 重新构建系统提示词
	chatbot.SystemPrompt = buildSystemPrompt(chatbot.Personality, chatbot.Background)
	chatbot.UpdatedAt = time.Now()

	// 保存更新
	if err := dao.Database.Chatbot.UpdateChatbot(ctx, chatbot); err != nil {
		logger.ErrorWithCtx(ctx, "UpdateChatbot failed",
			logger.ErrorField(err),
		)
		return nil, errcode.ErrServer.WithDetails(err.Error())
	}

	return &reply.ChatbotResponse{
		ID:           chatbot.ID,
		Name:         chatbot.Name,
		Personality:  chatbot.Personality,
		Background:   chatbot.Background,
		SystemPrompt: chatbot.SystemPrompt,
		CreatedAt:    chatbot.CreatedAt,
		UpdatedAt:    chatbot.UpdatedAt,
	}, nil
}

// DeleteChatbot 删除聊天机器人
func (e *eino) DeleteChatbot(ctx *gin.Context, id int32) errcode.Err {
	if err := dao.Database.Chatbot.DeleteChatbot(ctx, id); err != nil {
		logger.ErrorWithCtx(ctx, "DeleteChatbot failed",
			logger.ErrorField(err),
			logger.Int32("id", id),
		)
		return errcode.ErrServer.WithDetails(err.Error())
	}

	logger.InfoWithCtx(ctx, "Chatbot deleted successfully",
		logger.Int32("id", id),
	)

	return nil
}

// ChatbotChat 聊天机器人对话
func (e *eino) ChatbotChat(ctx *gin.Context, chatbotID int32, param *request.ParamChatbotChat) (*reply.ChatResponse, errcode.Err) {
	startTime := time.Now()

	// 获取聊天机器人
	chatbot, err := dao.Database.Chatbot.GetChatbot(ctx, chatbotID)
	if err != nil {
		return nil, errcode.ErrResourceNotFound.WithDetails("chatbot not found")
	}

	// 获取最新的对话历史（用于聊天上下文）
	maxHistory := global.Config.Eino.MaxHistory
	if maxHistory == 0 {
		maxHistory = 20 // 默认值
	}
	history, err := dao.Database.Chatbot.GetLatestChatHistory(ctx, chatbotID, maxHistory)
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get chat history",
			logger.ErrorField(err),
		)
		history = []*entity.ChatHistory{}
	}

	// 构建消息列表
	messages := make([]*schema.Message, 0, len(history)+2)

	// 添加系统提示词
	if chatbot.SystemPrompt != "" {
		messages = append(messages, schema.SystemMessage(chatbot.SystemPrompt))
	}

	// 添加历史对话
	for _, h := range history {
		if h.Role == "user" {
			messages = append(messages, schema.UserMessage(h.Content))
		} else if h.Role == "assistant" {
			messages = append(messages, schema.AssistantMessage(h.Content, nil))
		}
	}

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(param.Message))

	// 调用 Eino ChatModel
	chatModel := GetEinoChatModel()
	response, err := chatModel.Generate(context.Background(), messages)
	if err != nil {
		logger.ErrorWithCtx(ctx, "ChatbotChat failed",
			logger.ErrorField(err),
		)
		return nil, errcode.ErrServer.WithDetails(err.Error())
	}

	responseContent := response.Content
	duration := time.Since(startTime).Milliseconds()

	// 保存对话历史
	requestID := logger.GetRequestID(ctx)
	go func() {
		// 保存用户消息
		userHistory := &entity.ChatHistory{
			ChatbotID: chatbotID,
			Role:      "user",
			Content:   param.Message,
			CreatedAt: time.Now(),
		}
		dao.Database.Chatbot.AddChatHistory(context.Background(), userHistory)

		// 保存AI回复
		assistantHistory := &entity.ChatHistory{
			ChatbotID: chatbotID,
			Role:      "assistant",
			Content:   responseContent,
			CreatedAt: time.Now(),
		}
		dao.Database.Chatbot.AddChatHistory(context.Background(), assistantHistory)

		// 删除该聊天机器人的对话历史缓存（使缓存失效）
		if err := deleteChatHistoryCache(chatbotID); err != nil {
			logger.Warn("Failed to delete chat history cache",
				logger.ErrorField(err),
				logger.Int32("chatbot_id", chatbotID),
				logger.String("request_id", requestID),
			)
		} else {
			logger.Debug("Chat history cache invalidated successfully",
				logger.Int32("chatbot_id", chatbotID),
				logger.String("request_id", requestID),
			)
		}
	}()

	logger.InfoWithCtx(ctx, "ChatbotChat completed",
		logger.Int32("chatbot_id", chatbotID),
		logger.Int64("duration_ms", duration),
	)

	modelName := global.Config.Ollama.Model
	if modelName == "" {
		modelName = "deepseek-r1:8b"
	}

	return &reply.ChatResponse{
		Message:   responseContent,
		Duration:  duration,
		Timestamp: time.Now(),
	}, nil
}

// ChatbotChatStream 聊天机器人流式对话
func (e *eino) ChatbotChatStream(ctx *gin.Context, chatbotID int32, param *request.ParamChatbotChat) (*schema.StreamReader[*schema.Message], error) {
	// 获取聊天机器人
	chatbot, err := dao.Database.Chatbot.GetChatbot(ctx, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("chatbot not found: %w", err)
	}

	// 获取最新的对话历史（用于聊天上下文）
	maxHistory := global.Config.Eino.MaxHistory
	if maxHistory == 0 {
		maxHistory = 20 // 默认值
	}
	history, err := dao.Database.Chatbot.GetLatestChatHistory(ctx, chatbotID, maxHistory)
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get chat history",
			logger.ErrorField(err),
		)
		history = []*entity.ChatHistory{}
	}

	// 构建消息列表
	messages := make([]*schema.Message, 0, len(history)+2)

	// 添加系统提示词
	if chatbot.SystemPrompt != "" {
		messages = append(messages, schema.SystemMessage(chatbot.SystemPrompt))
	}

	// 添加历史对话
	for _, h := range history {
		if h.Role == "user" {
			messages = append(messages, schema.UserMessage(h.Content))
		} else if h.Role == "assistant" {
			messages = append(messages, schema.AssistantMessage(h.Content, nil))
		}
	}

	// 添加当前用户消息
	messages = append(messages, schema.UserMessage(param.Message))

	// 调用 Eino ChatModel 流式接口
	chatModel := GetEinoChatModel()
	streamReader, err := chatModel.Stream(context.Background(), messages)
	if err != nil {
		logger.ErrorWithCtx(ctx, "ChatbotChatStream failed",
			logger.ErrorField(err),
		)
		return nil, err
	}

	// 注意：流式响应中，历史记录的保存应该在 controller 层处理
	// 因为 StreamReader 只能读取一次，不能在这里读取后再返回给 controller

	return streamReader, nil
}

// SaveChatHistory 保存对话历史（辅助方法）
func (e *eino) SaveChatHistory(ctx context.Context, chatbotID int32, role, content string) {
	history := &entity.ChatHistory{
		ChatbotID: chatbotID,
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}
	if err := dao.Database.Chatbot.AddChatHistory(ctx, history); err != nil {
		logger.Warn("Failed to save chat history",
			logger.ErrorField(err),
			logger.Int32("chatbot_id", chatbotID),
			logger.String("role", role),
		)
		return
	}

	// 删除该聊天机器人的对话历史缓存（使缓存失效）
	go func() {
		if err := deleteChatHistoryCache(chatbotID); err != nil {
			logger.Warn("Failed to delete chat history cache",
				logger.ErrorField(err),
				logger.Int32("chatbot_id", chatbotID),
			)
		}
	}()
}

// 缓存相关常量和方法
const (
	// 对话历史缓存 key 前缀
	chatHistoryCacheKeyPrefix = "cache:chatbot:history"
	// 单条消息缓存 key 前缀
	chatMessageCacheKeyPrefix = "cache:chatbot:message"
	// 消息ID列表缓存 key 前缀
	chatMessageIdsCacheKeyPrefix = "cache:chatbot:message:ids"
)

// getChatHistoryCacheKey 获取对话历史缓存 key
func getChatHistoryCacheKey(chatbotID int32, page, pageSize int) string {
	return fmt.Sprintf("%s:%d:page:%d:size:%d", chatHistoryCacheKeyPrefix, chatbotID, page, pageSize)
}

// getChatMessageCacheKey 获取单条消息缓存 key
func getChatMessageCacheKey(messageID int32) string {
	return fmt.Sprintf("%s:%d", chatMessageCacheKeyPrefix, messageID)
}

// getChatMessageIdsCacheKey 获取消息ID列表缓存 key
func getChatMessageIdsCacheKey(chatbotID int32) string {
	return fmt.Sprintf("%s:%d", chatMessageIdsCacheKeyPrefix, chatbotID)
}

// deleteChatHistoryCache 删除对话历史缓存（缓存失效，删除指定聊天机器人的所有分页缓存）
// 如果 Redis 不可用，静默失败（不影响主流程）
func deleteChatHistoryCache(chatbotID int32) error {
	if !cache.IsCacheEnabled() {
		return nil
	}

	// 使用缓存工具包删除该聊天机器人的所有分页缓存
	pattern := fmt.Sprintf("%s:%d:*", chatHistoryCacheKeyPrefix, chatbotID)
	if err := cache.DeleteCacheByPattern(context.Background(), pattern); err != nil {
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// GetChatHistory 获取对话历史（分页，带缓存）
func (e *eino) GetChatHistory(ctx *gin.Context, chatbotID int32, param *request.ParamGetChatHistory) (*reply.ChatHistoryResponse, *reply.PageInfo, errcode.Err) {
	page := param.GetPage()
	pageSize := param.GetPageSize()

	logger.DebugWithCtx(ctx, "GetChatHistory with pagination",
		logger.Int32("chatbot_id", chatbotID),
		logger.Int("page", page),
		logger.Int("page_size", pageSize),
	)

	// 尝试使用Redis有序集合实现高效分页
	if cache.IsCacheEnabled() {
		if result, pageInfo, err := e.getChatHistoryFromRedis(ctx, chatbotID, page, pageSize); err == nil {
			return result, pageInfo, nil
		}
	}

	// 缓存未命中或Redis不可用，从数据库查询
	history, err := dao.Database.Chatbot.GetChatHistory(ctx, chatbotID, page, pageSize)
	if err != nil {
		logger.ErrorWithCtx(ctx, "GetChatHistory failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer.WithDetails(err.Error())
	}

	// 统计总数
	total, err := dao.Database.Chatbot.CountChatHistory(ctx, chatbotID)
	if err != nil {
		logger.ErrorWithCtx(ctx, "CountChatHistory failed",
			logger.ErrorField(err),
		)
		return nil, nil, errcode.ErrServer.WithDetails(err.Error())
	}

	// 转换数据格式
	list := make([]reply.ChatHistoryItem, 0, len(history))
	for _, h := range history {
		list = append(list, reply.ChatHistoryItem{
			ID:        h.ID,
			Role:      h.Role,
			Content:   h.Content,
			CreatedAt: h.CreatedAt,
		})

		// 异步将单条消息写入缓存
		if cache.IsCacheEnabled() {
			go e.cacheChatMessage(ctx, h)
		}
	}

	// 构建分页信息
	pageInfo := reply.NewPageInfo(page, pageSize, total)

	// 构建响应
	result := &reply.ChatHistoryResponse{
		List: list,
	}

	// 更新Redis有序集合（异步）
	if cache.IsCacheEnabled() {
		go e.updateChatMessageIdsCache(ctx, chatbotID, history)
	}

	logger.DebugWithCtx(ctx, "GetChatHistory success",
		logger.Int("count", len(list)),
		logger.Int64("total", total),
	)

	return result, pageInfo, nil
}

// getChatHistoryFromRedis 从Redis获取聊天历史
func (e *eino) getChatHistoryFromRedis(ctx *gin.Context, chatbotID int32, page, pageSize int) (*reply.ChatHistoryResponse, *reply.PageInfo, error) {
	// 1. 获取消息ID列表（使用有序集合的分页功能）
	idKey := getChatMessageIdsCacheKey(chatbotID)
	start := int64((page - 1) * pageSize)
	stop := int64(page*pageSize - 1)

	// ZRevRange：按时间倒序获取消息ID
	messageIDs, err := global.RedisClient.ZRevRange(ctx, idKey, start, stop).Result()
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get message IDs from Redis",
			logger.ErrorField(err),
			logger.String("key", idKey),
		)
		return nil, nil, err
	}

	if len(messageIDs) == 0 {
		// 没有消息，返回空结果
		return &reply.ChatHistoryResponse{List: []reply.ChatHistoryItem{}}, reply.NewPageInfo(page, pageSize, 0), nil
	}

	// 2. 获取消息总数
	total, err := global.RedisClient.ZCard(ctx, idKey).Result()
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get message count from Redis",
			logger.ErrorField(err),
			logger.String("key", idKey),
		)
		// 总数获取失败，从数据库查询
		if total, err = dao.Database.Chatbot.CountChatHistory(ctx, chatbotID); err != nil {
			return nil, nil, err
		}
	}

	// 3. 批量获取消息内容（使用Pipeline减少网络开销）
	pipeline := global.RedisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(messageIDs))

	for i, id := range messageIDs {
		messageKey := fmt.Sprintf("%s:%s", chatMessageCacheKeyPrefix, id)
		cmds[i] = pipeline.Get(ctx, messageKey)
	}

	_, err = pipeline.Exec(ctx)
	if err != nil {
		logger.WarnWithCtx(ctx, "Failed to get messages from Redis pipeline",
			logger.ErrorField(err),
		)
		return nil, nil, err
	}

	// 4. 组装结果
	list := make([]reply.ChatHistoryItem, 0, len(messageIDs))
	for i, cmd := range cmds {
		if data, err := cmd.Result(); err == nil {
			var message entity.ChatHistory
			if err := json.Unmarshal([]byte(data), &message); err == nil {
				list = append(list, reply.ChatHistoryItem{
					ID:        message.ID,
					Role:      message.Role,
					Content:   message.Content,
					CreatedAt: message.CreatedAt,
				})
			}
		} else if err != redis.Nil {
			logger.WarnWithCtx(ctx, "Failed to get message from Redis",
				logger.ErrorField(err),
				logger.String("id", messageIDs[i]),
			)
		}
	}

	// 如果缓存的消息数量不足，返回失败，让主函数从数据库查询
	if len(list) < len(messageIDs) {
		return nil, nil, fmt.Errorf("insufficient cached messages")
	}

	// 构建分页信息
	pageInfo := reply.NewPageInfo(page, pageSize, total)

	logger.DebugWithCtx(ctx, "GetChatHistory from Redis",
		logger.Int("count", len(list)),
		logger.Int64("total", total),
	)

	return &reply.ChatHistoryResponse{List: list}, pageInfo, nil
}

// cacheChatMessage 缓存单条聊天消息
func (e *eino) cacheChatMessage(ctx context.Context, message *entity.ChatHistory) {
	if !cache.IsCacheEnabled() {
		return
	}

	messageKey := getChatMessageCacheKey(message.ID)
	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.Warn("Failed to marshal chat message",
			logger.ErrorField(err),
			logger.Int32("message_id", message.ID),
		)
		return
	}

	// 使用随机过期时间
	expire := cache.GetRandomExpiration()
	if err := global.RedisClient.Set(ctx, messageKey, jsonData, expire).Err(); err != nil {
		logger.Warn("Failed to cache chat message",
			logger.ErrorField(err),
			logger.String("cache_key", messageKey),
		)
	}
}

// updateChatMessageIdsCache 更新聊天消息ID列表缓存
func (e *eino) updateChatMessageIdsCache(ctx context.Context, chatbotID int32, messages []*entity.ChatHistory) {
	if !cache.IsCacheEnabled() || len(messages) == 0 {
		return
	}

	idKey := getChatMessageIdsCacheKey(chatbotID)
	pipe := global.RedisClient.Pipeline()

	for _, msg := range messages {
		// 分数使用消息ID（假设ID是递增的）或创建时间戳
		score := float64(msg.ID)
		pipe.ZAdd(ctx, idKey, redis.Z{Score: score, Member: msg.ID})
	}

	// 设置过期时间
	pipe.Expire(ctx, idKey, cache.GetHotDataExpiration())

	if _, err := pipe.Exec(ctx); err != nil {
		logger.Warn("Failed to update chat message IDs cache",
			logger.ErrorField(err),
			logger.String("cache_key", idKey),
		)
	}
}
