# 多文档 WebSocket 传输层

## 概述

`MultiDocWebSocketTransport` 允许通过**单个 WebSocket 连接**处理多个文档的实时协作，极大地提高了资源利用效率。

## 对比

### 旧方案：每文档一个 WebSocket

```
Browser ─┬─ WebSocket 1 ─→ doc1.txt
         ├─ WebSocket 2 ─→ doc2.txt
         ├─ WebSocket 3 ─→ doc3.txt
         └─ WebSocket N ─→ docN.txt
```

**缺点**：
- 资源开销大（N 个连接）
- 连接管理复杂
- 服务器负担重

### 新方案：单 WebSocket 多文档

```
Browser ── WebSocket ──→ Server
                        ├─ doc1.txt
                        ├─ doc2.txt
                        ├─ doc3.txt
                        └─ docN.txt
```

**优点**：
- 只有一个 WebSocket 连接
- 资源占用少
- 更简单的客户端代码
- 与服务器设计一致

## API 使用

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/coreseekdev/texere/pkg/transport"
)

func main() {
    ctx := context.Background()

    // 1. 创建多文档传输
    transport := transport.NewMultiDocWebSocketTransport(
        "client-123",
        "ws://localhost:8080/ws",
    )
    defer transport.Close()

    // 2. 连接到服务器
    if err := transport.Connect(ctx); err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }

    // 3. 订阅多个文档
    doc1Sub, _ := transport.Subscribe("/doc1.txt")
    doc2Sub, _ := transport.Subscribe("/doc2.txt")
    doc3Sub, _ := transport.Subscribe("/doc3.txt")

    // 4. 为每个文档设置消息处理器
    setupDoc1Handlers(doc1Sub)
    setupDoc2Handlers(doc2Sub)
    setupDoc3Handlers(doc3Sub)

    // 5. 发送操作（指定文档）
    transport.SendOperation("/doc1.txt", []interface{}{5, "Hello"})
    transport.SendOperation("/doc2.txt", []interface{}{0, "World"})

    // 6. 发送心跳（包含多个 session）
    transport.SendHeartbeat([]string{"session-1", "session-2", "session-3"})

    // 保持运行...
    select {}
}

func setupDoc1Handlers(sub *transport.DocumentSubscription) {
    // 接收快照
    sub.OnSnapshot(func(data *transport.SnapshotData) {
        fmt.Printf("Doc1 snapshot: %s (v%d)\n",
            data.Content, data.Revision)
    })

    // 接收其他客户端的操作
    sub.OnRemoteOperation(func(data *transport.RemoteOperationData) {
        fmt.Printf("Doc1 remote op from %s: %v\n",
            data.ClientID, data.Operation)
    })

    // 用户加入
    sub.OnUserJoined(func(data *transport.UserJoinedData) {
        fmt.Printf("User %s joined Doc1\n", data.ClientID)
    })

    // 用户离开
    sub.OnUserLeft(func(data *transport.UserLeftData) {
        fmt.Printf("User %s left Doc1\n", data.ClientID)
    })

    // Session 更新
    sub.OnSessionInfo(func(data *transport.SessionInfoData) {
        fmt.Printf("Doc1 session: %d readers, %d writers\n",
            data.ReaderCount, data.WriterCount)
    })

    // 错误处理
    sub.OnError(func(data *transport.ErrorData) {
        fmt.Printf("Doc1 error: %s - %s\n", data.Code, data.Message)
    })
}
```

### 订阅/取消订阅

```go
// 订阅文档
sub, err := transport.Subscribe("/new-doc.txt")
if err != nil {
    log.Printf("Failed to subscribe: %v", err)
    return
}

// 设置为只读
sub.ReadOnly = true

// 发送订阅消息（自动发送到服务器）

// 取消订阅
err = transport.Unsubscribe("/new-doc.txt")
if err != nil {
    log.Printf("Failed to unsubscribe: %v", err)
}
```

### 查询订阅状态

```go
// 列出所有订阅的文档
docs := transport.ListSubscriptions()
fmt.Println("Subscribed documents:", docs)
// 输出: [/doc1.txt /doc2.txt /doc3.txt]

// 检查是否订阅了特定文档
if sub, exists := transport.GetSubscription("/doc1.txt"); exists {
    fmt.Printf("Found subscription for %s (session: %s)\n",
        sub.DocPath, sub.SessionID)
}

// 检查连接状态
if transport.IsConnected() {
    fmt.Println("WebSocket is connected")
}
```

### 发送操作

```go
// 发送 OT 操作
err := transport.SendOperation("/doc1.txt", []interface{}{
    5,      // Retain 5 characters
    "Hello", // Insert "Hello"
})
if err != nil {
    log.Printf("Failed to send operation: %v", err)
}

// 带上下文发送
err = transport.SendOperationWithContext(ctx, "/doc1.txt", []interface{}{
    0,  // Insert at position 0
    "World",
})
```

## 高级用法

### 动态管理文档订阅

```go
// 文档管理器
type DocumentManager struct {
    transport *transport.MultiDocWebSocketTransport
    docs      map[string]*transport.DocumentSubscription
}

func NewDocumentManager(endpoint string) *DocumentManager {
    return &DocumentManager{
        transport: transport.NewMultiDocWebSocketTransport("client-1", endpoint),
        docs:      make(map[string]*transport.DocumentSubscription),
    }
}

func (dm *DocumentManager) OpenDocument(ctx context.Context, docPath string) error {
    // 订阅文档
    sub, err := dm.transport.Subscribe(docPath)
    if err != nil {
        return err
    }

    // 设置处理器
    sub.OnSnapshot(func(data *transport.SnapshotData) {
        // 加载文档内容到编辑器
        fmt.Printf("Loaded %s: %s\n", docPath, data.Content)
    })

    sub.OnRemoteOperation(func(data *transport.RemoteOperationData) {
        // 应用远程操作
        dm.applyRemoteOperation(docPath, data)
    })

    dm.docs[docPath] = sub
    return nil
}

func (dm *DocumentManager) CloseDocument(docPath string) error {
    delete(dm.docs, docPath)
    return dm.transport.Unsubscribe(docPath)
}

func (dm *DocumentManager) applyRemoteOperation(docPath string, data *transport.RemoteOperationData) {
    // 应用操作到文档
    fmt.Printf("Apply operation to %s: %v\n", docPath, data.Operation)
}
```

### 多标签页编辑器

```go
// 编辑器标签页管理
type TabManager struct {
    transport   *transport.MultiDocWebSocketTransport
    activeTabs map[string]*EditorTab
}

type EditorTab struct {
    DocPath   string
    Content   string
    Modified  bool
    ClientID  string
    SessionID string
    Users     int // 连接用户数
}

func NewTabManager(endpoint string) *TabManager {
    return &TabManager{
        transport:   transport.NewMultiDocWebSocketTransport("editor", endpoint),
        activeTabs: make(map[string]*EditorTab),
    }
}

func (tm *TabManager) OpenTab(docPath string) (*EditorTab, error) {
    // 检查是否已打开
    if tab, exists := tm.activeTabs[docPath]; exists {
        return tab, nil
    }

    // 订阅文档
    sub, err := tm.transport.Subscribe(docPath)
    if err != nil {
        return nil, err
    }

    // 创建标签
    tab := &EditorTab{
        DocPath:  docPath,
        Content:  "",
        Modified: false,
        ClientID: "editor-client",
    }

    // 设置事件处理
    sub.OnSnapshot(func(data *transport.SnapshotData) {
        tab.Content = data.Content
        tab.Modified = false
        tab.SessionID = data.SessionID
        fmt.Printf("[Tab] %s loaded: %d chars\n", docPath, len(data.Content))
    })

    sub.OnRemoteOperation(func(data *transport.RemoteOperationData) {
        // 远程操作：应用并标记为已修改
        tab.Modified = true
        fmt.Printf("[Tab] %s modified by %s\n", docPath, data.ClientID)
    })

    sub.OnUserJoined(func(data *transport.UserJoinedData) {
        tab.Users++
        fmt.Printf("[Tab] %s: %d users\n", docPath, tab.Users)
    })

    sub.OnUserLeft(func(data *transport.UserLeftData) {
        tab.Users--
        fmt.Printf("[Tab] %s: %d users\n", docPath, tab.Users)
    })

    tm.activeTabs[docPath] = tab
    return tab, nil
}

func (tm *TabManager) CloseTab(docPath string) error {
    delete(tm.activeTabs, docPath)
    return tm.transport.Unsubscribe(docPath)
}

func (tm *TabManager) SaveTab(docPath string) error {
    tab := tm.activeTabs[docPath]

    // 发送保存操作
    return tm.transport.SendOperation(docPath, []interface{}{
        0, "SAVE: " + tab.Content,
    })
}
```

### React 组件示例（伪代码）

```jsx
function useMultiDocTransport(endpoint) {
    const [transport, setTransport] = useState(null);
    const [documents, setDocuments] = useState({});

    useEffect(() => {
        const t = new MultiDocWebSocketTransport("client", endpoint);

        // 连接
        t.connect().catch(err => {
            console.error("Connection failed:", err);
        });

        setTransport(t);

        return () => {
            t.close();
        };
    }, [endpoint]);

    const openDocument = useCallback(async (docPath) => {
        if (!transport) return;

        // 订阅文档
        const sub = await transport.subscribe(docPath);

        // 设置处理器
        sub.onSnapshot((data) => {
            setDocuments(prev => ({
                ...prev,
                [docPath]: {
                    ...prev[docPath],
                    content: data.content,
                    revision: data.revision,
                }
            }));
        });

        sub.onRemoteOperation((data) => {
            console.log(`Remote op on ${docPath}:`, data.operation);
            // 应用到编辑器
        });

        return sub;
    }, [transport]);

    const closeDocument = useCallback((docPath) => {
        transport?.unsubscribe(docPath);
        setDocuments(prev => {
            const {[docPath]: _, ...rest} = prev;
            return rest;
        });
    }, [transport]);

    const sendOperation = useCallback((docPath, operation) => {
        transport?.sendOperation(docPath, operation);
    }, [transport]);

    return { transport, documents, openDocument, closeDocument, sendOperation };
}

// 使用示例
function Editor() {
    const { documents, openDocument, closeDocument, sendOperation } =
        useMultiDocTransport("ws://localhost:8080/ws");

    const handleOpen = async (path) => {
        await openDocument(path);
    };

    const handleEdit = (path, op) => {
        sendOperation(path, op);
    };

    return (
        <div>
            <button onClick={() => handleOpen("/doc1.txt")}>Open Doc 1</button>
            <button onClick={() => handleOpen("/doc2.txt")}>Open Doc 2</button>

            {Object.entries(documents).map(([path, doc]) => (
                <DocumentEditor
                    key={path}
                    path={path}
                    content={doc.content}
                    onEdit={(op) => handleEdit(path, op)}
                />
            ))}
        </div>
    );
}
```

## 性能对比

| 指标 | 单文档模式 | 多文档模式 |
|------|-----------|-----------|
| WebSocket 连接数 | N（文档数） | 1 |
| 内存占用 | ~N × 10KB | ~10KB |
| 服务器连接数 | N | 1 |
| 网络开销 | N × 心跳 | 1 × 心跳 |
| 适用场景 | 少量文档 | 多文档协作 |

## 服务器兼容性

服务器端已经支持多文档单连接（见 `handler.go`）：

- ✅ **Session 管理**：每个文档独立的 `EditSession`
- ✅ **消息路由**：通过 `SessionID` 路由到对应文档
- ✅ **广播机制**：`broadcastToSession()` 只广播给订阅了该文档的客户端
- ✅ **Heartbeat**：支持多个 session 的心跳

## 注意事项

1. **线程安全**：`MultiDocWebSocketTransport` 是线程安全的，可以并发调用
2. **连接管理**：单个连接断开会影响所有订阅的文档
3. **错误处理**：建议为每个文档设置独立的错误处理器
4. **资源清理**：关闭 transport 会自动关闭所有订阅
5. **SessionID 映射**：服务器会在响应中返回 `SessionID`，会自动映射到订阅

## 迁移指南

### 从单文档迁移到多文档

**旧代码（单文档）**：
```go
// 为每个文档创建独立连接
doc1Conn := NewWebSocketTransport("transport-1", "client-1", "/doc1.txt")
doc2Conn := NewWebSocketTransport("transport-2", "client-1", "/doc2.txt")
doc3Conn := NewWebSocketTransport("transport-3", "client-1", "/doc3.txt")

doc1Conn.Connect(ctx)
doc2Conn.Connect(ctx)
doc3Conn.Connect(ctx)
```

**新代码（多文档）**：
```go
// 单个连接处理多个文档
multiDocConn := NewMultiDocWebSocketTransport("client-1", "ws://localhost:8080/ws")
multiDocConn.Connect(ctx)

multiDocConn.Subscribe("/doc1.txt")
multiDocConn.Subscribe("/doc2.txt")
multiDocConn.Subscribe("/doc3.txt")
```

## 总结

`MultiDocWebSocketTransport` 提供了高效的多文档协作支持：

✅ **资源高效** - 单 WebSocket 连接处理多文档
✅ **API 简洁** - 清晰的订阅/取消订阅接口
✅ **类型安全** - 每个文档独立的处理器
✅ **线程安全** - 支持并发访问
✅ **完全测试** - 覆盖所有核心功能

适用于：多标签页编辑器、协作编辑平台、文档管理系统等场景。
