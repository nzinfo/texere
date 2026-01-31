# History Service 接口设计

## 概述

History Service 提供了一个抽象接口，类似于 Content 和 Auth 接口，用于管理文档的版本历史。

## 架构设计

```
┌─────────────────────────────────────────────────────────┐
│                  SessionManager                         │
│  ┌───────────────────────────────────────────────────┐  │
│  │               EditSession                         │  │
│  │  - 1 snapshot + 200 recent changes (memory)      │  │
│  │  - Forwards to HistoryService for long-term       │  │
│  └──────────────┬────────────────────────────────────┘  │
│                 │                                        │
│                 ▼                                        │
│  ┌───────────────────────────────────────────────────┐  │
│  │         HistoryService (interface)                │  │
│  │  - OnSnapshot()                                    │  │
│  │  - OnOperation()                                   │  │
│  │  - GetSnapshot()                                   │  │
│  │  - ReconstructSnapshot()                           │  │
│  └──────┬─────────────────────────────────┬──────────┘  │
│         │                                 │              │
│    ┌────▼────────┐              ┌────────▼──────┐       │
│    │ Redis        │              │ Memory         │       │
│    │ History      │              │ History        │       │
│    │ Service      │              │ Service        │       │
│    └──────────────┘              └───────────────┘       │
└───────────────────────────────────────────────────────────┘
```

## 接口定义

```go
// HistoryService provides version history storage and retrieval.
type HistoryService interface {
    // OnSnapshot handles snapshot events from edit sessions
    OnSnapshot(event *HistoryEvent) error

    // OnOperation handles operation events from edit sessions
    OnOperation(event *HistoryEvent) error

    // GetSnapshot retrieves a specific snapshot from storage
    GetSnapshot(ctx context.Context, sessionID string, versionID int64) (*HistoryEvent, error)

    // GetSessionHistory retrieves history for a session
    GetSessionHistory(ctx context.Context, sessionID string, limit int64) ([]*HistoryEvent, error)

    // ReconstructSnapshot reconstructs content for a version (patch mode)
    ReconstructSnapshot(ctx context.Context, sessionID string, targetVersionID int64) (string, error)

    // ListSnapshots lists all snapshots for a session
    ListSnapshots(ctx context.Context, sessionID string) ([]*SnapshotInfo, error)

    // Close closes the history service
    Close() error
}
```

## 实现方式

### 1. RedisHistoryService（Redis 存储）

适合生产环境的分布式部署：

```go
// 创建 Redis history service（默认：完整内容模式）
redisClient := NewMiniRedis() // 或真实 Redis 客户端
historySvc := NewRedisHistoryService(redisClient)

// 或使用 patch 模式（节省存储空间）
historySvc := NewRedisHistoryServiceWithOpts(redisClient, true)
```

**特点**：
- 支持完整内容模式（简单）
- 支持 Patch 模式（节省 70-90% 空间）
- 支持分布式部署
- 异步事件处理
- 死锁安全

### 2. MemoryHistoryService（内存存储）

适合测试和单实例部署：

```go
// 创建内存 history service
historySvc := NewMemoryHistoryService(false) // 完整内容模式
historySvc := NewMemoryHistoryService(true)  // Patch 模式
```

**特点**：
- 纯内存实现
- 快速访问
- 适合测试
- 无需外部依赖

### 3. 使用工厂函数

通过配置选项创建：

```go
historySvc := NewHistoryService(&HistoryOptions{
    StorageBackend: "redis",    // 或 "memory"
    UsePatchMode:    true,      // 启用 patch 模式
    MaxChangesBeforeSnapshot: 200,
    MaxSnapshotInterval: 300,
    RedisAddr: "localhost:6379",
})
```

## 使用示例

### 方式 1：直接创建 SessionManager

```go
package main

import (
    "github.com/coreseekdev/texere/pkg/transport"
)

func main() {
    // 创建 Redis history service
    redisClient := transport.NewMiniRedis()
    historySvc := transport.NewRedisHistoryService(redisClient)
    defer historySvc.Close()

    // 创建带 history 的 session manager
    sm := transport.NewSessionManagerWithHistory(historySvc)

    // 使用 session manager
    session, _ := sm.GetOrCreateSession("/test.txt")
    session.SetContent("Hello World")
    session.AddOperation([]interface{}{5, " World"}, "client-1")
}
```

### 方式 2：使用工厂函数（推荐）

```go
package main

import (
    "github.com/coreseekdev/texere/pkg/transport"
)

func main() {
    // 使用工厂函数创建 history service
    historySvc := transport.NewHistoryService(&transport.HistoryOptions{
        StorageBackend: "redis",
        UsePatchMode:    true,  // 启用 patch 模式节省空间
    })
    defer historySvc.Close()

    // 创建 session manager
    sm := transport.NewSessionManagerWithHistory(historySvc)

    // 正常使用...
}
```

### 方式 3：运行时切换

```go
// 根据配置选择不同的 history service
var historySvc transport.HistoryService

if config.UseRedis {
    historySvc = transport.NewRedisHistoryServiceWithOpts(redisClient, config.UsePatchMode)
} else {
    historySvc = transport.NewMemoryHistoryService(config.UsePatchMode)
}

sm := transport.NewSessionManagerWithHistory(historySvc)
```

### 方式 4：自定义实现

实现 HistoryService 接口：

```go
type DatabaseHistoryService struct {
    db *sql.DB
}

func (s *DatabaseHistoryService) OnSnapshot(event *transport.HistoryEvent) error {
    // 保存到数据库
    _, err := s.db.Exec("INSERT INTO snapshots ...")
    return err
}

func (s *DatabaseHistoryService) GetSnapshot(ctx context.Context, sessionID string, versionID int64) (*transport.HistoryEvent, error) {
    // 从数据库读取
    // ...
}

// 实现其他接口方法...

// 使用自定义实现
historySvc := &DatabaseHistoryService{db: myDB}
sm := transport.NewSessionManagerWithHistory(historySvc)
```

## 版本回滚示例

```go
// 获取历史版本列表
snapshots, _ := historySvc.ListSnapshots(context.Background(), sessionID)

// 回滚到指定版本
targetVersion := int64(5)

// 方式 1：直接获取快照
snapshot, err := historySvc.GetSnapshot(context.Background(), sessionID, targetVersion)
if err == nil {
    fmt.Printf("Version %d content: %s\n", targetVersion, snapshot.Content)
}

// 方式 2：使用 ReconstructSnapshot（patch 模式）
content, err := historySvc.ReconstructSnapshot(context.Background(), sessionID, targetVersion)
if err == nil {
    fmt.Printf("Reconstructed version %d: %s\n", targetVersion, content)

    // 恢复内容到会话
    session.SetContent(content)
}
```

## 设计优势

1. **接口抽象**：使用 HistoryService 接口，易于测试和替换实现
2. **灵活配置**：可以选择不同的存储后端
3. **兼容性好**：RedisHistoryService 同时实现 HistoryListener 接口
4. **易于扩展**：可以轻松添加新的实现（如 Database、S3 等）
5. **与现有架构一致**：遵循 Content、Auth 接口的设计模式

## 性能对比

| 实现 | 读取速度 | 写入速度 | 存储空间 | 分布式 |
|------|---------|---------|---------|--------|
| MemoryHistoryService | 最快 | 快 | 受限 | 否 |
| RedisHistoryService (完整内容) | 快 | 快 | 大 | 是 |
| RedisHistoryService (Patch模式) | 中等 | 中等 | 小 (70-90%节省) | 是 |
| DatabaseHistoryService (自定义) | 慢 | 慢 | 大 | 是 |

## 配置建议

**开发/测试环境**：
```go
historySvc := NewMemoryHistoryService(false) // 简单快速
```

**生产环境（小文档）**：
```go
historySvc := NewRedisHistoryService(redisClient) // 完整内容模式
```

**生产环境（大文档 > 1KB）**：
```go
historySvc := NewRedisHistoryServiceWithOpts(redisClient, true) // Patch 模式
```

**混合模式**：
```go
// 根据文档大小动态选择
func SmartHistoryService(redisClient RedisClient, docSize int) HistoryService {
    if docSize > 1024 {
        return NewRedisHistoryServiceWithOpts(redisClient, true) // Patch
    }
    return NewRedisHistoryService(redisClient) // 完整内容
}
```
