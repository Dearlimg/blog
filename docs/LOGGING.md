# 日志系统使用指南

## 为什么需要专业的日志系统？

### 1. **fmt.Println 的问题**

```go
// ❌ 不好的做法
fmt.Println("User logged in")
fmt.Println("Error:", err)
fmt.Printf("Processing %d items\n", count)
```

**问题：**
- 没有日志级别（无法区分 info、error、debug）
- 无法控制输出位置（控制台、文件、远程服务）
- 没有结构化数据（难以解析和搜索）
- 无法过滤和查询
- 生产环境难以管理

### 2. **专业日志系统的优势**

✅ **日志级别**：Debug、Info、Warn、Error、Fatal  
✅ **结构化日志**：JSON 格式，易于解析  
✅ **多输出目标**：控制台、文件、远程服务  
✅ **日志轮转**：自动管理日志文件大小和数量  
✅ **性能优化**：异步写入，不阻塞业务  
✅ **上下文信息**：自动记录时间、调用位置等  

---

## Zap 日志库简介

**Zap** 是 Uber 开发的高性能结构化日志库，特点：
- 🚀 **高性能**：比标准库快 4-10 倍
- 📦 **结构化**：支持 JSON 和文本格式
- 🔧 **可配置**：丰富的配置选项
- 🎯 **零分配**：关键路径零内存分配

---

## 日志级别说明

### Debug（调试）
用于开发调试，记录详细的执行流程。

```go
logger.Debug("Cache hit",
    logger.String("key", "user:123"),
    logger.Int("ttl", 3600),
)
```

### Info（信息）
记录重要的业务事件，如服务启动、用户操作等。

```go
logger.Info("User logged in",
    logger.String("user_id", "123"),
    logger.String("ip", "192.168.1.1"),
)
```

### Warn（警告）
记录潜在问题，但不影响服务运行。

```go
logger.Warn("Redis connection failed, using fallback",
    logger.ErrorField(err),
)
```

### Error（错误）
记录错误信息，需要关注和修复。

```go
logger.Error("Failed to create user",
    logger.ErrorField(err),
    logger.String("email", email),
)
```

### Fatal（致命）
严重错误，程序无法继续运行，会退出。

```go
logger.Fatal("Database connection failed",
    logger.ErrorField(err),
)
```

---

## 使用方法

### 基本用法

```go
import "blog/pkg/logger"

// 简单日志
logger.Info("Server started")
logger.Error("Operation failed", logger.ErrorField(err))

// 带字段的日志
logger.Info("User created",
    logger.String("user_id", "123"),
    logger.String("email", "user@example.com"),
    logger.Int("age", 25),
)
```

### 常用字段类型

```go
// 字符串
logger.String("key", "value")

// 整数
logger.Int("count", 100)
logger.Int32("id", 123)
logger.Int64("timestamp", 1234567890)

// 浮点数
logger.Float64("price", 99.99)

// 布尔值
logger.Bool("enabled", true)

// 任意类型
logger.Any("data", someStruct)

// 错误（最常用）
logger.ErrorField(err)
```

### 实际使用示例

#### 1. 服务启动/关闭

```go
// main.go
logger.Info("Server starting",
    logger.String("name", global.Config.App.Name),
    logger.String("port", global.Config.App.Port),
)

// 关闭时
logger.Info("Server shutting down...")
defer logger.Sync() // 确保日志缓冲区刷新
```

#### 2. 业务逻辑

```go
// 成功操作
logger.Info("Message created successfully",
    logger.String("name", param.Name),
    logger.String("email", param.Email),
)

// 失败操作
logger.Error("CreateMessage failed",
    logger.ErrorField(err),
    logger.String("name", param.Name),
)
```

#### 3. 缓存操作

```go
// 缓存命中
logger.Debug("Cache hit",
    logger.Int("count", len(messages)),
)

// 缓存未命中
logger.Debug("Cache miss, querying database")

// 缓存错误
logger.Warn("Failed to set cache (non-blocking)",
    logger.ErrorField(err),
)
```

#### 4. 数据库操作

```go
logger.Info("Database connected successfully",
    logger.String("dsn", dsn),
)

logger.Error("Query failed",
    logger.ErrorField(err),
    logger.String("query", sql),
)
```

---

## 配置说明

在 `config/app/config.yaml` 中配置：

```yaml
log:
  level: info          # 日志级别: debug, info, warn, error
  format: json         # 日志格式: json, text
  output: both        # 输出位置: stdout, file, both
  file_path: logs/blog.log
  max_size: 100       # 单个日志文件最大大小（MB）
  max_backups: 10     # 保留的备份文件数量
  max_age: 30         # 保留日志文件的天数
  compress: true      # 是否压缩旧日志文件
```

### 配置项说明

- **level**: 日志级别，低于此级别的日志不会输出
- **format**: 
  - `json`: JSON 格式，适合生产环境和日志收集系统
  - `text`: 文本格式，适合开发环境，可读性好
- **output**: 
  - `stdout`: 只输出到控制台
  - `file`: 只输出到文件
  - `both`: 同时输出到控制台和文件
- **file_path**: 日志文件路径
- **max_size**: 单个日志文件最大大小（MB），超过会自动轮转
- **max_backups**: 保留的备份文件数量
- **max_age**: 保留日志文件的天数，超过会自动删除
- **compress**: 是否压缩旧日志文件（节省空间）

---

## 日志格式示例

### JSON 格式（生产环境推荐）

```json
{
  "level": "info",
  "timestamp": "2025-01-27T10:30:45.123Z",
  "caller": "logic/message.go:25",
  "msg": "Message created successfully",
  "name": "John",
  "email": "john@example.com"
}
```

### 文本格式（开发环境推荐）

```
2025-01-27T10:30:45.123Z    INFO    logic/message.go:25    Message created successfully    {"name": "John", "email": "john@example.com"}
```

---

## 最佳实践

### ✅ 推荐做法

1. **使用合适的日志级别**
   ```go
   logger.Debug("Detailed debug info")  // 开发调试
   logger.Info("Important business event")  // 业务事件
   logger.Warn("Potential issue")  // 警告
   logger.Error("Error occurred")  // 错误
   ```

2. **添加上下文信息**
   ```go
   logger.Error("Failed to process order",
       logger.ErrorField(err),
       logger.String("order_id", orderID),
       logger.String("user_id", userID),
   )
   ```

3. **使用结构化字段**
   ```go
   // ✅ 好
   logger.Info("User logged in",
       logger.String("user_id", userID),
       logger.String("ip", ip),
   )
   
   // ❌ 不好
   logger.Info(fmt.Sprintf("User %s logged in from %s", userID, ip))
   ```

4. **程序退出前同步日志**
   ```go
   defer logger.Sync()  // 确保日志缓冲区刷新
   ```

### ❌ 避免的做法

1. **不要使用 fmt.Println**
   ```go
   // ❌ 不好
   fmt.Println("Error:", err)
   
   // ✅ 好
   logger.Error("Operation failed", logger.ErrorField(err))
   ```

2. **不要记录敏感信息**
   ```go
   // ❌ 不好
   logger.Info("User logged in",
       logger.String("password", password),  // 不要记录密码！
   )
   ```

3. **不要过度使用 Debug**
   ```go
   // ❌ 不好（生产环境会产生大量日志）
   logger.Debug("Processing item", logger.Int("index", i))  // 在循环中
   
   // ✅ 好
   logger.Debug("Starting batch processing", logger.Int("total", len(items)))
   ```

---

## 日志查询和分析

### 使用 jq 查询 JSON 日志

```bash
# 查看所有错误日志
cat logs/blog.log | jq 'select(.level == "error")'

# 查看特定用户的日志
cat logs/blog.log | jq 'select(.user_id == "123")'

# 查看最近1小时的日志
cat logs/blog.log | jq 'select(.timestamp > "2025-01-27T09:00:00Z")'
```

### 使用 grep 查询文本日志

```bash
# 查看所有错误日志
grep "ERROR" logs/blog.log

# 查看特定关键词
grep "user_id.*123" logs/blog.log
```

---

## 性能考虑

Zap 是高性能日志库，但仍有注意事项：

1. **避免在热路径中使用 Debug**
   ```go
   // 如果日志级别是 info，debug 日志不会执行，但仍有函数调用开销
   logger.Debug("Processing...")  // 在循环中避免
   ```

2. **使用结构化字段而不是字符串拼接**
   ```go
   // ✅ 好（零分配）
   logger.Info("User", logger.String("id", userID))
   
   // ❌ 不好（有内存分配）
   logger.Info(fmt.Sprintf("User %s", userID))
   ```

---

## 总结

专业的日志系统带来的好处：

1. ✅ **可观测性**：了解系统运行状态
2. ✅ **问题排查**：快速定位和解决问题
3. ✅ **性能监控**：分析系统性能瓶颈
4. ✅ **业务分析**：分析用户行为和业务指标
5. ✅ **合规要求**：满足审计和合规需求

从 `fmt.Println` 迁移到专业日志系统，是后端服务成熟度的重要标志！

