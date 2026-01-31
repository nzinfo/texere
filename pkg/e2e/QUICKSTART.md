# E2E 测试框架快速开始

## 安装依赖

项目使用标准的 [gorilla/websocket](https://github.com/gorilla/websocket) 实现。如果网络受限，可以通过代理安装：

```bash
export https_proxy=http://127.0.0.1:7890
export http_proxy=http://127.0.0.1:7890
export all_proxy=socks5://127.0.0.1:7890

go get github.com/gorilla/websocket
```

## 快速开始

### 1. 基础并发编辑测试

```go
package e2e

import (
    "testing"
    "time"
    "github.com/coreseekdev/texere/pkg/session"
)

func TestBasicConcurrentEditing(t *testing.T) {
    // 创建测试框架
    framework := NewTestFramework()
    defer framework.StopServer()

    // 启动 WebSocket 服务器
    if err := framework.StartServer(":8080", "websocket"); err != nil {
        t.Fatalf("Failed to start server: %v", err)
    }

    // 创建测试会话
    _, err := framework.CreateSession("test-doc", "Hello World", session.DocTypeString)
    if err != nil {
        t.Fatalf("Failed to create session: %v", err)
    }

    // 定义测试规格：2 个客户端同时编辑
    testSpec := &TestSpec{
        DocID:          "test-doc",
        InitialContent: "Hello World",
        DocType:        session.DocTypeString,
        TransportType:  "websocket",
        VerifyConsistency: true,
        Clients: []*ClientSpec{
            {
                ID: "user1",
                Operations: []*ClientOperation{
                    {
                        Type:     OpInsert,
                        Position: 5,
                        Content:  " Alice",
                    },
                },
            },
            {
                ID: "user2",
                Operations: []*ClientOperation{
                    {
                        Type:     OpInsert,
                        Position: 5,
                        Content:  " Bob",
                        Delay:    50 * time.Millisecond,
                    },
                },
            },
        },
    }

    // 运行测试
    result := framework.RunConcurrentTest(testSpec)

    // 验证结果
    if !result.Success() {
        t.Errorf("Test failed: %s", result.String())
    }

    t.Logf("Test completed: %s", result.String())
}
```

### 2. 运行测试

```bash
# 运行所有 e2e 测试
go test ./pkg/e2e/... -v

# 运行特定测试
go test ./pkg/e2e/... -run TestBasicConcurrentEditing -v

# 运行基准测试
go test ./pkg/e2e/... -bench=. -benchmem

# 短模式（跳过耗时测试）
go test ./pkg/e2e/... -short
```

## 测试结果示例

```
TestResult{
    Duration: 150ms,
    Clients: 2,
    Success: 2,
    Failures: 0,
    Consistent: true
}
```

## 支持的操作

### 插入操作
```go
{
    Type:     OpInsert,
    Position: 5,
    Content:  " World",
}
```

### 删除操作
```go
{
    Type:     OpDelete,
    Position: 0,
    Length:   5,  // 删除 5 个字符
}
```

### 带延迟的操作
```go
{
    Type:     OpInsert,
    Position: 10,
    Content:  " Delayed",
    Delay:    100 * time.Millisecond,
}
```

## 核心组件

### TestFramework
测试框架核心，管理服务器和会话。

### SimulatedClient
模拟客户端，可以连接到服务器并发送操作。

### ClientPool
客户端池，用于管理多个模拟客户端。

### TestResult
测试结果，包含成功/失败统计和一致性检查。

## 高级特性

### 1. 竞争条件测试
多个客户端同时编辑同一位置，验证 OT 转换的正确性。

### 2. 大规模并发测试
支持数百个客户端同时连接和编辑。

### 3. 客户端池管理
统一管理多个测试客户端的生命周期。

### 4. 一致性验证
自动验证所有客户端的文档状态是否一致。

## 架构优势

- ✅ 使用 gorilla/websocket 标准实现
- ✅ 轻量级 goroutine 并发
- ✅ 真实的 WebSocket 通信
- ✅ 完整的 OT 操作转换
- ✅ 自动化一致性检查

## 参考资源

- [完整文档](./README.md)
- [gorilla/websocket 文档](https://pkg.go.dev/github.com/gorilla/websocket)
