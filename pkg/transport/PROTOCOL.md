# WebSocket 协议规范 (Collaborative Editing)

> **版本**: 1.0
> **协议类型**: WebSocket
> **编码格式**: JSON
> **OT 格式**: 基于 ot.js 的数组格式

---

## 概述

这是一个用于实时协作编辑的 WebSocket 协议，支持：
- ✅ 多文件同时关注
- ✅ 读写分离计数
- ✅ OT 操作转换
- ✅ 会话生命周期管理
- ✅ 只读订阅（可用 SSE 优化）

---

## 核心概念

### Edit Session (编辑会话)

每个文件对应一个 **Edit Session**，使用 **UUID** 标识：

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 引用计数

每个会话维护两个计数器：
- **ReaderCount**: 只读订阅者数量
- **WriterCount**: 活跃编辑者数量

只有当 `ReaderCount == 0 && WriterCount == 0` 时，会话才被销毁。

---

## 消息格式

所有消息使用 JSON 格式：

```json
{
  "type": "message_type",
  "session_id": "optional-session-uuid",
  "timestamp": 1706745600,
  "data": { ... },
  "metadata": { ... }
}
```

---

## 客户端 → 服务器消息

### 1. 关注文件 (subscribe)

关注一个文件，即使该文件未被编辑。

```json
{
  "type": "subscribe",
  "timestamp": 1706745600,
  "data": {
    "file_path": "/path/to/file.txt",
    "read_only": true,
    "use_sse": true
  }
}
```

**字段说明**:
- `file_path`: 文件路径
- `read_only`: `true` = 只读订阅（可用 SSE），`false` = 准备编辑
- `use_sse`: `true` = 优先使用 SSE 推送变更（仅 read_only 时有效）

**服务器响应**:
- 如果文件已在编辑：发送 `snapshot` + 最近操作
- 如果文件未编辑：发送 `snapshot` + 空内容

---

### 2. 取消关注 (unsubscribe)

取消关注一个文件。

```json
{
  "type": "unsubscribe",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

### 3. 开始编辑 (start_editing)

开始编辑一个文件（自动订阅）。

```json
{
  "type": "start_editing",
  "timestamp": 1706745600,
  "data": {
    "file_path": "/path/to/file.txt",
    "content_type": "text",
    "initial_text": "Optional initial content if file doesn't exist"
  }
}
```

**效果**:
- 创建或加入编辑会话
- `WriterCount++`
- 接收完整快照

---

### 4. 停止编辑 (stop_editing)

停止编辑一个文件。

```json
{
  "type": "stop_editing",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**效果**:
- `WriterCount--`
- 其他用户可以看到该用户离开

**注意**: 停止编辑 ≠ 销毁会话。只有所有编辑者和订阅者都离开，会话才销毁。

---

### 5. 发送操作 (operation)

发送 OT 操作到服务器。

```json
{
  "type": "operation",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "revision": 42,
    "operation": [5, "Hello", 10, -3],
    "selection": {
      "position": 15,
      "selection_end": 15
    }
  }
}
```

**OT 操作格式** (基于 ot.js):
- 数组格式: `[retain, insert, retain, delete]`
- `5` - 保留前 5 个字符
- `"Hello"` - 插入 "Hello"
- `10` - 保留 10 个字符
- `-3` - 删除 3 个字符

**服务器响应**: `ack` 消息

---

### 6. 光标位置 (cursor)

更新光标/选择位置（可选）。

```json
{
  "type": "cursor",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "position": 100,
    "selection_end": 105
  }
}
```

---

### 7. 心跳 (heartbeat)

保持连接活跃。

```json
{
  "type": "heartbeat",
  "timestamp": 1706745600,
  "data": {
    "session_ids": [
      "550e8400-e29b-41d4-a716-446655440000",
      "660e8400-e29b-41d4-a716-446655440001"
    ]
  }
}
```

---

## 服务器 → 客户端消息

### 1. 欢迎消息 (welcome)

连接成功时发送。

```json
{
  "type": "welcome",
  "timestamp": 1706745600,
  "data": {
    "client_id": "client-123",
    "server_id": "server-abc",
    "timestamp": 1706745600
  }
}
```

---

### 2. 文档快照 (snapshot)

发送文档完整内容 + 最近操作。

```json
{
  "type": "snapshot",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "file_path": "/path/to/file.txt",
    "content": "Hello World",
    "revision": 42,
    "created_at": 1706745000,
    "updated_at": 1706745500,
    "operations": [[5, " Alice"], [11, " Bob"]],
    "clients": [
      {
        "client_id": "client-1",
        "is_editing": true,
        "selection": {"position": 10, "selection_end": 10},
        "updated_at": 1706745590
      }
    ],
    "read_only": false
  }
}
```

---

### 3. 远程操作 (remote_operation)

其他客户端发送的操作。

```json
{
  "type": "remote_operation",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "client_id": "client-2",
    "revision": 43,
    "operation": [10, "Beautiful"],
    "selection": {"position": 19, "selection_end": 19}
  }
}
```

**客户端应**:
1. 应用 OT 转换（如果需要）
2. 应用操作到本地文档
3. 更新 UI

---

### 4. 操作确认 (ack)

服务器确认收到操作。

```json
{
  "type": "ack",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "revision": 43,
    "timestamp": 1706745600
  }
}
```

---

### 5. 错误消息 (error)

操作失败时发送。

```json
{
  "type": "error",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "code": "operation_failed",
    "message": "Failed to apply operation",
    "details": {}
  }
}
```

**错误码**:
- `invalid_subscribe_data` - 订阅数据无效
- `invalid_unsubscribe_data` - 取消订阅数据无效
- `invalid_start_editing_data` - 开始编辑数据无效
- `invalid_stop_editing_data` - 停止编辑数据无效
- `invalid_operation_data` - 操作数据无效
- `invalid_operation` - OT 操作无效
- `operation_failed` - 操作应用失败
- `session_not_found` - 会话不存在

---

### 6. 用户加入 (user_joined)

新用户加入会话。

```json
{
  "type": "user_joined",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "client_id": "client-3",
    "client": {
      "client_id": "client-3",
      "name": "Alice",
      "color": "#ff0000",
      "is_editing": true,
      "selection": {"position": 0, "selection_end": 0},
      "updated_at": 1706745600
    }
  }
}
```

---

### 7. 用户离开 (user_left)

用户离开会话。

```json
{
  "type": "user_left",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "client_id": "client-3"
  }
}
```

---

### 8. 会话信息 (session_info)

会话状态更新。

```json
{
  "type": "session_info",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": 1706745600,
  "data": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "file_path": "/path/to/file.txt",
    "reader_count": 2,
    "writer_count": 1,
    "clients": [...],
    "is_editing": true
  }
}
```

---

## 完整工作流示例

### 场景 1: 用户开始编辑一个新文件

```
Client                                    Server
  |                                         |
  |--- start_editing ---------------------->|
  |     {file_path: "/doc.txt"}            |
  |                                         |
  |                                         | 创建会话
  |                                         | WriterCount = 1
  |                                         |
  |<---------------- snapshot ---------------|
  |     {content: "", revision: 0}          |
  |                                         |
  |--- operation -------------------------->|
  |     {revision: 0, op: [6, "Hello"]}    |
  |                                         |
  |                                         | 应用操作
  |                                         | Revision = 1
  |                                         |
  |<---------------- ack ---------------------|
  |     {revision: 1}                       |
```

---

### 场景 2: 两个用户同时编辑

```
Alice Client                              Server                           Bob Client
    |                                         |                                  |
    |--- start_editing ---------------------->|                                  |
    |                                         |                                  |
    |<---------------- snapshot ---------------|                                  |
    |     {content: "Hello"}                   |                                  |
    |                                         |                                  |
    |                                         |--- snapshot ------------------>|
    |                                         |     {content: "Hello"}          |
    |                                         |                                  |
    |--- operation -------------------------->|                                  |
    |     {revision: 0, op: [5, " Alice"]}   |                                  |
    |                                         |                                  |
    |                                         | 应用操作                           |
    |                                         | Revision = 1                       |
    |                                         |                                  |
    |<---------------- ack -------------------|                                  |
    |                                         |--- remote_operation ------------>|
    |                                         |     {op: [5, " Alice"]}           |
    |                                         |                                  |
    |                                         |<--- operation ------------------|
    |                                         |     {revision: 0, op: [11, " Bob"]}
    |                                         |                                  |
    |                                         | OT 转换                            |
    |                                         | [11, " Bob"] -> [16, " Bob"]       |
    |                                         |                                  |
    |<---------------- remote_operation -------|                                  |
    |     {op: [16, " Bob"]}                  |                                  |
    |                                         |<---------------- ack -------------|
    |                                         |     {revision: 2}                |
```

**最终文档内容**: "Hello Alice Bob"

---

### 场景 3: 只读订阅（使用 SSE）

```
Reader Client                             Server                           Editor Client
    |                                         |                                  |
    |--- subscribe -------------------------->|                                  |
    |     {read_only: true, use_sse: true}   |                                  |
    |                                         |                                  |
    |                                         |--- SSE snapshot -------------->|
    |                                         |     {content: "Hello"}          |
    |                                         |                                  |
    |                                         |<--- operation ------------------|
    |                                         |     {revision: 0, op: [5, " World"]}
    |                                         |                                  |
    |                                         | 应用操作                           |
    |                                         |                                  |
    |                                         |--- SSE remote_operation ------>|
    |                                         |     {op: [5, " World"]}          |
```

---

## SSE 优化

对于只读订阅 (`read_only: true, use_sse: true`)，服务器应使用 **SSE** 推送变更：

```
Event: snapshot
Data: {"session_id":"...", "content":"Hello World", ...}

Event: remote_operation
Data: {"operation":[5, " Beautiful"], ...}

Event: remote_operation
Data: {"operation":[16, " Day"], ...}
```

**优势**:
- 减少服务器连接数
- 浏览器原生支持
- 自动重连

---

## OT 操作格式详解

### 基础类型

| 类型 | 格式 | 示例 | 说明 |
|------|------|------|------|
| **Retain** | `number` | `5` | 保留前 N 个字符，光标前进 |
| **Insert** | `string` | `"Hello"` | 在当前位置插入文本 |
| **Delete** | `-number` | `-3` | 删除 N 个字符 |

### 示例操作

#### 插入 "Hello" 到位置 5

```json
[5, "Hello"]
```

#### 在位置 5 删除 3 个字符，然后插入 "World"

```json
[5, -3, "World"]
```

#### 复杂操作

```json
[10, "Alice", 5, -2, 15, " Bob"]
```

解释：
1. 保留前 10 个字符
2. 插入 "Alice"
3. 保留 5 个字符
4. 删除 2 个字符
5. 保留 15 个字符
6. 插入 " Bob"

---

## 多文件订阅

客户端可以同时订阅多个文件：

```json
{
  "type": "subscribe",
  "data": {"file_path": "/doc1.txt", "read_only": true}
}

{
  "type": "subscribe",
  "data": {"file_path": "/doc2.txt", "read_only": false}
}
```

**服务器响应**:
每个订阅发送独立的 `snapshot` 消息，包含对应的 `session_id`。

**客户端处理**:
根据 `session_id` 区分不同文件的变更。

---

## 错误处理

### 客户端应处理

1. **网络断开**: 自动重连，重新订阅所有会话
2. **版本不匹配**: 请求完整快照
3. **操作失败**: 显示错误，不更新本地文档

### 服务器应处理

1. **无效操作**: 返回 `error` 消息
2. **会话不存在**: 返回 `session_not_found`
3. **权限错误**: 返回 `permission_denied`

---

## 性能优化

### 1. 操作批处理

客户端可以缓存多个操作，批量发送：

```json
{
  "type": "operation",
  "data": {
    "operation": [[5, "A"], [6, "B"], [7, "C"]],
    "revision": 10
  }
}
```

服务器会先 `compose` 这些操作，然后应用。

### 2. 增量同步

只发送变更部分，不是完整文档。

### 3. SSE 推送

只读订阅使用 SSE，减少 WebSocket 连接数。

---

## 安全考虑

### 1. 认证

建议在 WebSocket 连接建立时进行认证：

```
ws://server/ws?token=xxx
```

### 2. 权限

服务器应验证：
- 客户端是否有权限访问文件
- 客户端是否有编辑权限（vs 只读）

### 3. 会话隔离

不同会话之间数据隔离，使用 `session_id` (UUID) 区分。

---

## 总结

这个协议提供了：

✅ **完整的多文件协作支持**
✅ **读写分离的会话管理**
✅ **基于 OT 的冲突解决**
✅ **SSE 优化只读订阅**
✅ **清晰的错误处理**
✅ **易于扩展**

基于 **ot.js** 的 OT 格式，与现有生态系统兼容。
