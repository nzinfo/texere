# Texere - 文档编织引擎

> Weave Knowledge Together - 编织知识，连接智慧

Texere 是一个基于 Operational Transformation 和 AI 的文档协作与生成引擎。

## 🧵 核心概念

Texere 将文档视为"织物"，通过编织多源内容来创建完整的文档：

- **协同编辑**：编织多人的实时编辑（OT）
- **AI 生成**：编织 LLM 的智能创作
- **知识融合**：编织多源信息（RAG）
- **文档合成**：编织最终的知识产物

## 📦 项目结构

```
texere/
├── cmd/                        # 命令行工具和服务器入口
│   ├── texere-server/          # 主服务器
│   │   └── main.go
│   ├── texere-cli/             # CLI 工具
│   │   └── main.go
│   └── texere-migrate/         # 数据迁移工具
│       └── main.go
│
├── pkg/                        # 公共库（可被外部导入）
│   ├── syntaxis/               # 📝 Operational Transformation 核心包
│   │   ├── operation/          # 操作定义（Insert/Delete/Retain）
│   │   ├── transform/          # 转换算法（Include/Exclude）
│   │   ├── compose/            # 操作组合
│   │   ├── history/            # 撤销/重做历史
│   │   └── syntaxis.go         # 公共 API
│   │
│   ├── concordia/              # 🤝 协作状态管理（替代方案）
│   │   ├── document/           # 文档状态
│   │   ├── session/            # 会话管理
│   │   ├── awareness/          # 用户感知（光标、选择）
│   │   └── consensus/          # 分布式共识
│   │
│   ├── ordo/                   # ⏱️ 时间与排序（替代方案）
│   │   ├── clock/              # 逻辑时钟（Lamport）
│   │   ├── vector/             # 向量时钟
│   │   ├── ordering/           # 全局排序
│   │   └── version/            # 版本管理
│   │
│   ├── textor/                 # 📝 文本处理
│   │   ├── rope/               # Rope 数据结构
│   │   ├── piecetable/         # Piece Table 实现
│   │   ├── cursor/             # 光标操作
│   │   └── selection/          # 文本选择
│   │
│   ├── fabric/                 # 🧵 文档织物结构
│   │   ├── document/           # 文档模型
│   │   ├── block/              # 文档块
│   │   ├── delta/              # 增量变更
│   │   └── patch/              # 补丁应用
│   │
│   ├── ai/                     # 🤖 AI 集成
│   │   ├── llm/                # LLM 抽象层
│   │   ├── prompt/             # 提示工程
│   │   ├── stream/             # 流式生成
│   │   └── template/           # 模板引擎
│   │
│   ├── weave/                  # 🧶 核心编织引擎
│   │   ├── engine/             # 主引擎
│   │   ├── pipeline/           # 编织流水线
│   │   ├── transformer/        # 内容转换
│   │   └── merger/             # 内容合并
│   │
│   ├── flux/                   # 🌊 数据流与同步
│   │   ├── transport/          # 传输层抽象
│   │   ├── websocket/          # WebSocket 实现
│   │   ├── webrtc/             # WebRTC 实现
│   │   └── sync/               # 同步协议
│   │
│   └── store/                  # 💾 持久化存储
│       ├── database/           # 数据库抽象
│       ├── repository/         # 仓库模式
│       ├── snapshot/           # 快照管理
│       └── cache/              # 缓存层
│
├── internal/                   # 内部实现（不对外暴露）
│   ├── server/                 # 服务器核心
│   │   ├── http/               # HTTP API
│   │   ├── ws/                 # WebSocket 处理
│   │   ├── rpc/                # RPC 服务
│   │   └── middleware/         # 中间件
│   │
│   ├── client/                 # 客户端 SDK
│   │   ├── go/                 # Go 客户端
│   │   └── protocol/           # 协议定义
│   │
│   ├── config/                 # 配置管理
│   │   ├── loader/             # 配置加载
│   │   └── validator/          # 配置验证
│   │
│   └── logger/                 # 日志系统
│       ├── format/             # 日志格式化
│       └── rotate/             # 日志轮转
│
├── api/                        # API 定义
│   ├── openapi/                # OpenAPI 规范
│   │   └── texere.yaml
│   ├── graphql/                # GraphQL schema
│   │   └── schema.graphql
│   └── proto/                  # Protocol Buffers
│       └── texere.proto
│
├── web/                        # Web 前端（可选）
│   ├── src/                    # 源码
│   ├── public/                 # 静态资源
│   └── package.json
│
├── deployments/                # 部署配置
│   ├── docker/                 # Docker 配置
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/             # K8s 配置
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── terraform/              # Terraform 配置
│       └── main.tf
│
├── scripts/                    # 构建和部署脚本
│   ├── build.sh                # 构建脚本
│   ├── test.sh                 # 测试脚本
│   ├── release.sh              # 发布脚本
│   └── deploy.sh               # 部署脚本
│
├── docs/                       # 文档
│   ├── architecture/           # 架构文档
│   │   ├── ot-algorithm.md
│   │   ├── ai-integration.md
│   │   └── data-structures.md
│   ├── api/                    # API 文档
│   │   ├── rest-api.md
│   │   └── websocket-protocol.md
│   ├── guides/                 # 使用指南
│   │   ├── getting-started.md
│   │   └── advanced-usage.md
│   └── research/               # 研究文档
│       ├── ot-survey.md
│       ├── llm-integration.md
│       └── benchmarks.md
│
├── examples/                   # 示例代码
│   ├── simple-editor/          # 简单编辑器示例
│   ├── ai-assistant/           # AI 助手示例
│   └── real-time-collab/       # 实时协作示例
│
├── test/                       # 测试
│   ├── unit/                   # 单元测试
│   ├── integration/            # 集成测试
│   ├── benchmark/              # 基准测试
│   └── e2e/                    # 端到端测试
│
├── tools/                      # 开发工具
│   ├── mockgen/                # Mock 生成
│   ├── protoc/                 # Proto 编译
│   └── lint/                   # 代码检查
│
├── .github/                    # GitHub 配置
│   ├── workflows/              # CI/CD
│   │   ├── ci.yml
│   │   └── release.yml
│   ├── ISSUE_TEMPLATE/         # Issue 模板
│   └── PULL_REQUEST_TEMPLATE.md
│
├── .gitignore
├── .golangci.yml               # Go linter 配置
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── CONTRIBUTING.md
├── LICENSE
└── CHANGELOG.md
```

## 🏛️ 命名体系

### 主项目：Texere
- **含义**：编织（拉丁语）
- **定位**：文档编织引擎
- **范围**：整个平台

### 核心子包命名（拉丁/希腊语词根）

#### 1. **syntaxis** - OT 协同（推荐）✅
- **词源**：希腊语 *σύνταξις* (协调、排列)
- **职责**：Operational Transformation 核心
- **包路径**：`pkg/syntaxis/`

```go
import "github.com/coreseekdev/texere/pkg/syntaxis"

op := syntaxis.NewInsert(0, "Hello")
transformed := syntaxis.Transform(op1, op2)
```

#### 2. **concordia** - 协作状态管理（备选）
- **词源**：拉丁语 *concordia* (和谐、共识)
- **职责**：会话管理、用户感知、分布式共识
- **包路径**：`pkg/concordia/`

```go
import "github.com/coreseekdev/texere/pkg/concordia"

session := concordia.NewSession()
users := session.GetAwareness()
```

#### 3. **ordo** - 时间与排序（备选）
- **词源**：拉丁语 *ordo* (顺序、秩序)
- **职责**：逻辑时钟、向量时钟、版本管理
- **包路径**：`pkg/ordo/`

```go
import "github.com/coreseekdev/texere/pkg/ordo"

clock := ordo.NewLamportClock()
timestamp := clock.Tick()
```

#### 4. **textor** - 文本处理
- **词源**：拉丁语 *textor* (编织者、文本者)
- **职责**：Rope、Piece Table、光标操作
- **包路径**：`pkg/textor/`

```go
import "github.com/coreseekdev/texere/pkg/textor"

rope := textor.NewRope("Hello World")
```

#### 5. **fabric** - 文档织物
- **词源**：英语 *fabric* (织物、结构)
- **职责**：文档模型、块结构、增量变更
- **包路径**：`pkg/fabric/`

```go
import "github.com/coreseekdev/texere/pkg/fabric"

doc := fabric.NewDocument()
```

#### 6. **weave** - 编织引擎
- **词源**：英语 *weave* (编织)
- **职责**：核心编织引擎、流水线
- **包路径**：`pkg/weave/`

```go
import "github.com/coreseekdev/texere/pkg/weave"

engine := weave.NewEngine()
engine.Weave(&humanEdit, &aiGeneration)
```

#### 7. **flux** - 数据流
- **词源**：拉丁语 *fluxus* (流动)
- **职责**：WebSocket、传输、同步
- **包路径**：`pkg/flux/`

```go
import "github.com/coreseekdev/texere/pkg/flux"

transport := flux.NewWebSocket()
```

## 📦 OT 协同包命名方案对比

| 方案 | 包名 | 词源 | 含义 | 推荐度 |
|------|------|------|------|--------|
| **方案 A** | `pkg/syntaxis` | 希腊语 | 协调、排列 | ⭐⭐⭐⭐⭐ |
| **方案 B** | `pkg/concordia` | 拉丁语 | 和谐、共识 | ⭐⭐⭐⭐ |
| **方案 C** | `pkg/ordo` | 拉丁语 | 顺序、秩序 | ⭐⭐⭐ |
| **方案 D** | `pkg/unio` | 拉丁语 | 统一 | ⭐⭐⭐ |

### 方案 A：syntaxis（最推荐）✅

```go
import "github.com/coreseekdev/texere/pkg/syntaxis"

// 使用示例
op1 := syntaxis.NewInsert(0, "Hello")
op2 := syntaxis.NewRetain(5)
op3 := syntaxis.NewInsert(" World")

composed := syntaxis.Compose(op1, op2, op3)
transformed := syntaxis.Transform(composed, otherOp)
```

**优点**：
- ✅ 准确描述 OT 的核心功能
- ✅ 技术圈熟悉（类似 syntax）
- ✅ 与 Texere 的拉丁语形成优雅对比

### 方案 B：concordia

```go
import "github.com/coreseekdev/texere/pkg/concordia"

// 使用示例
session := concordia.NewSession(docID)
concordia.JoinSession(session, user)
concordia.Broadcast(session, operation)
```

**优点**：
- ✅ 强调协作和谐
- ✅ 适合会话管理
- ⚠️ 但对 OT 算法本身描述不够精准

### 方案 C：ordo

```go
import "github.com/coreseekdev/texere/pkg/ordo"

// 使用示例
clock := ordo.NewLamportClock()
ordo.ClockIn(clock, operation)
```

**优点**：
- ✅ 简短易记
- ✅ 适合时间排序
- ⚠️ 但不够描述 OT 的转换特性

## 🎯 最终推荐

### 主项目：Texere ✅
- 编织文档的引擎

### OT 核心包：syntaxis ✅
- 操作转换的核心算法

### 协作状态包：concordia
- 会话和用户感知

### 时间排序包：ordo
- 逻辑时钟和版本

这样形成一个完整且语义清晰的命名体系：

```
Texere (编织引擎)
├── syntaxis (协调操作 - OT 核心)
├── concordia (协作和谐 - 会话管理)
├── ordo (时间秩序 - 版本控制)
├── textor (文本处理 - Rope/PT)
├── fabric (文档织物 - 文档模型)
├── weave (编织引擎 - AI + 人工)
└── flux (数据流动 - WebSocket)
```

## 📝 快速开始

```bash
# 克隆仓库
git clone https://github.com/coreseekdev/texere.git
cd texere

# 构建
make build

# 运行服务器
./bin/texere-server

# 运行测试
make test

# 运行基准测试
make benchmark
```

## 🧵 示例：编织文档

```go
package main

import (
    "github.com/coreseekdev/texere/pkg/syntaxis"
    "github.com/coreseekdev/texere/pkg/weave"
    "github.com/coreseekdev/texere/pkg/ai"
)

func main() {
    // 创建编织引擎
    engine := weave.NewEngine()

    // 添加人工编辑
    humanEdit := syntaxis.NewInsert(0, "Hello ")
    engine.WeaveHuman(humanEdit)

    // 添加 AI 生成
    aiGen := ai.NewLLMRequest("续写这段文本")
    engine.WeaveAI(aiGen)

    // 获取编织好的文档
    doc := engine.Document()
    println(doc.String()) // "Hello World"
}
```

## 📄 License

MIT License

## 🤝 Contributing

欢迎贡献！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)

---

**Texere - Weave Knowledge Together** 🧵✨
