# E2E Testing Framework for Collaborative Editing

这是一个为协作编辑系统设计的端到端（E2E）测试框架，支持模拟多用户并发编辑场景。

## 功能特性

- ✅ **WebSocket 客户端模拟**：模拟多个客户端连接到 WebSocket 服务器
- ✅ **并发编辑测试**：支持多个客户端同时编辑同一文档
- ✅ **一致性验证**：自动验证所有客户端的文档状态是否一致
- ✅ **操作类型支持**：支持插入、删除、保留等 OT 操作
- ✅ **延迟控制**：可为每个操作设置延迟，模拟真实场景
- ✅ **客户端池管理**：统一管理多个测试客户端
- ✅ **性能测试**：支持基准测试和压力测试

## 架构设计

### 核心组件

```
pkg/e2e/
├── framework.go          # 测试框架核心
├── client.go             # 模拟客户端实现
└── example_test.go       # 使用示例
```

### 测试流程

1. **创建测试框架** - `NewTestFramework()`
2. **启动服务器** - `StartServer(addr, transportType)`
3. **创建会话** - `CreateSession(docID, content, docType)`
4. **定义测试规格** - `TestSpec`
5. **运行测试** - `RunConcurrentTest(testSpec)`
6. **验证结果** - `TestResult.Success()`

## 使用示例

### 基础示例：3 个客户端并发编辑

```go
package e2e

import (
    "testing"
    "time"

    "github.com/coreseekdev/texere/pkg/session"
)

func TestConcurrentEditing(t *testing.T) {
    // 1. 创建测试框架
    framework := NewTestFramework()
    defer framework.StopServer()

    // 2. 启动 WebSocket 服务器
    if err := framework.StartServer(":8080", "websocket"); err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }

    // 3. 创建测试会话
    _, err := framework.CreateSession("test-doc", "Hello World", session.DocTypeString)
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }

    // 4. 定义测试规格
    testSpec := &TestSpec{
        DocID:          "test-doc",
        InitialContent: "Hello World",
        DocType:        session.DocTypeString,
        TransportType:  "websocket",
        VerifyConsistency: true,
        Timeout:        30 * time.Second,
        Clients: []*ClientSpec{
            {
                ID: "client-1",
                Operations: []*ClientOperation{
                    {
                        Type:     OpInsert,
                        Position: 5,
                        Content:  " Beautiful",
                    },
                },
            },
            {
                ID: "client-2",
                Operations: []*ClientOperation{
                    {
                        Type:     OpInsert,
                        Position: 11,
                        Content:  " Wonderful",
                        Delay:    50 * time.Millisecond,
                    },
                },
            },
        },
    }

    // 5. 运行测试
    result := framework.RunConcurrentTest(testSpec)

    // 6. 验证结果
    if !result.Success() {
        t.Errorf("Test failed: %s", result.String())
    }
}
```

### 操作类型

#### 插入操作

```go
{
    Type:     OpInsert,
    Position: 5,
    Content:  " Hello",
}
```

#### 删除操作

```go
{
    Type:     OpDelete,
    Position: 0,
    Length:   5,
}
```

#### 带延迟的操作

```go
{
    Type:     OpInsert,
    Position: 10,
    Content:  " World",
    Delay:    100 * time.Millisecond, // 操作前延迟 100ms
}
```

### 客户端池管理

```go
pool := NewClientPool()
defer pool.Close()

// 添加客户端
client := NewSimulatedClient("client-1", "doc-1")
pool.Add(client)

// 获取客户端
client, ok := pool.Get("client-1")

// 获取所有客户端
clients := pool.GetAll()

// 移除客户端
pool.Remove("client-1")

// 广播操作到所有客户端
op := &ClientOperation{
    Type:     OpInsert,
    Position: 0,
    Content:  "Broadcast",
}
pool.Broadcast(op)
```

## 高级测试场景

### 1. 竞争条件测试

```go
func TestRaceCondition(t *testing.T) {
    framework := NewTestFramework()
    defer framework.StopServer()

    framework.StartServer(":8080", "websocket")
    framework.CreateSession("race-doc", "ABC", session.DocTypeString)

    // 所有客户端同时尝试在同一位置插入
    testSpec := &TestSpec{
        DocID:          "race-doc",
        InitialContent: "ABC",
        DocType:        session.DocTypeString,
        TransportType:  "websocket",
        VerifyConsistency: true,
        Clients: []*ClientSpec{
            {ID: "client-1", Operations: []*ClientOperation{{
                Type: OpInsert, Position: 0, Content: "1",
            }}},
            {ID: "client-2", Operations: []*ClientOperation{{
                Type: OpInsert, Position: 0, Content: "2",
            }}},
            {ID: "client-3", Operations: []*ClientOperation{{
                Type: OpInsert, Position: 0, Content: "3",
            }}},
        },
    }

    result := framework.RunConcurrentTest(testSpec)

    if !result.ConsistencyCheck {
        t.Error("Documents are not synchronized!")
    }
}
```

### 2. 大规模并发测试

```go
func TestMassiveConcurrency(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }

    framework := NewTestFramework()
    defer framework.StopServer()

    framework.StartServer(":8080", "websocket")
    framework.CreateSession("massive-doc", "Initial", session.DocTypeString)

    // 创建 100 个并发客户端
    numClients := 100
    clients := make([]*ClientSpec, numClients)

    for i := 0; i < numClients; i++ {
        clients[i] = &ClientSpec{
            ID: fmt.Sprintf("client-%d", i),
            Operations: []*ClientOperation{{
                Type:     OpInsert,
                Position: 7,
                Content:  fmt.Sprintf(" [%d]", i),
                Delay:    time.Duration(i) * time.Millisecond,
            }},
        }
    }

    testSpec := &TestSpec{
        DocID:          "massive-doc",
        InitialContent: "Initial",
        DocType:        session.DocTypeString,
        TransportType:  "websocket",
        VerifyConsistency: true,
        Timeout:        60 * time.Second,
        Clients:        clients,
    }

    result := framework.RunConcurrentTest(testSpec)

    t.Logf("Result: %s", result.String())
}
```

### 3. 基准测试

```go
func BenchmarkConcurrentInsert(b *testing.B) {
    framework := NewTestFramework()
    defer framework.StopServer()

    framework.StartServer(":8080", "websocket")

    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        docID := fmt.Sprintf("bench-doc-%d", i)
        framework.CreateSession(docID, "Start", session.DocTypeString)

        testSpec := &TestSpec{
            DocID:          docID,
            InitialContent: "Start",
            DocType:        session.DocTypeString,
            TransportType:  "websocket",
            VerifyConsistency: false,
            Clients: []*ClientSpec{{
                ID: "bench-client",
                Operations: []*ClientOperation{{
                    Type:     OpInsert,
                    Position: 5,
                    Content:  " Benchmark",
                }},
            }},
        }

        result := framework.RunConcurrentTest(testSpec)
        if !result.Success() {
            b.Errorf("Benchmark failed: %s", result.String())
        }
    }
}
```

## 测试结果

`TestResult` 结构体包含详细的测试信息：

```go
type TestResult struct {
    StartTime           time.Time
    EndTime             time.Time
    Duration            time.Duration
    ClientCount         int
    SuccessCount        int32
    FailureCount        int32
    Errors              []error
    ConsistencyCheck    bool      // 所有一致性检查通过
    MessagesSent        int32
    MessagesReceived    int32
    OperationsApplied   int32
}
```

### 结果示例输出

```
TestResult{
    Duration: 150ms,
    Clients: 3,
    Success: 3,
    Failures: 0,
    Consistent: true
}
```

## 运行测试

### 运行所有 e2e 测试

```bash
go test ./pkg/e2e/...
```

### 运行特定测试

```bash
go test ./pkg/e2e/... -run TestConcurrentEditing
```

### 运行基准测试

```bash
go test ./pkg/e2e/... -bench=. -benchmem
```

### 短模式测试（跳过耗时测试）

```bash
go test ./pkg/e2e/... -short
```

### 详细输出

```bash
go test ./pkg/e2e/... -v
```

## 架构优势

### 1. 真实的并发场景
- 使用 goroutines 模拟真实并发
- 支持数千个并发客户端
- 轻量级并发原语（channel、atomic）

### 2. 完整的 OT 支持
- 操作转换（OT）支持
- 冲突检测和解决
- 版本一致性验证

### 3. 灵活的测试控制
- 可配置的操作延迟
- 支持多种传输类型（WebSocket、SSE）
- 细粒度的错误收集

### 4. 易于扩展
- 插件化的传输层
- 可自定义的客户端行为
- 支持自定义断言

## 最佳实践

1. **使用 defer 清理**：始终使用 `defer framework.StopServer()` 确保资源清理

2. **合理的超时设置**：为测试设置合理的超时时间，避免测试挂起

3. **一致性验证**：在并发测试中启用 `VerifyConsistency`

4. **错误处理**：检查 `TestResult.Errors` 了解失败原因

5. **性能测试**：使用 `-short` 标志跳过耗时的测试

## 相关文档

- [OT 包文档](../ot/README.md)
- [Session 包文档](../session/README.md)
- [Transport 包文档](../transport/README.md)

## 参考资源

基于以下主流 Go WebSocket 测试框架设计：

- [gorilla/websocket](https://pkg.go.dev/github.com/gorilla/websocket) - 最流行的 Go WebSocket 库
- [Carrot](https://github.com/gophercarrot/carrot) - 分布式 WebSocket 负载测试框架
- [websocket-tester-go](https://libraries.io/go/github.com%252FSaiNivedh26%252Fwebsocket-tester-go) - 高性能 WebSocket 负载测试工具
- [webstress](https://github.com/d-Rickyy-b/webstress) - WebSocket 压力测试工具
