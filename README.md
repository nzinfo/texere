# Texere

> **Weave Knowledge Together** - 编织知识，连接智慧

Texere 是一个基于 Operational Transformation 和 AI 的文档协作与生成引擎。

## 🧵 核心理念

Texere 将文档视为"织物"，通过编织多源内容来创建完整的文档：

- **协同编辑**：编织多人的实时编辑（OT）
- **AI 生成**：编织 LLM 的智能创作
- **知识融合**：编织多源信息（RAG）
- **文档合成**：编织最终的知识产物

## ✨ 特性

- **Operational Transformation (OT)**：基于 `concordia` 包的高效操作转换算法
- **实时协作**：基于 WebSocket 的低延迟同步
- **AI 集成**：支持 LLM 流式生成和文档合成
- **可扩展架构**：模块化设计，易于扩展
- **高性能**：使用 Rope/Piece Table 优化文本操作

## 🚀 快速开始

### 安装

```bash
go get github.com/coreseekdev/texere
```

### 基础使用

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/concordia"
    "github.com/coreseekdev/texere/pkg/weave"
)

func main() {
    // 创建编织引擎
    engine := weave.NewEngine()

    // 添加操作
    op1 := concordia.NewInsert(0, "Hello ")
    op2 := concordia.NewInsert(6, "World")

    engine.Weave(op1)
    engine.Weave(op2)

    // 获取文档
    doc := engine.Document()
    fmt.Println(doc.String()) // "Hello World"
}
```

### 运行服务器

```bash
# 构建
make build

# 运行
./bin/texere-server --port 8080
```

## 📦 包结构

```
texere/
├── pkg/concordia/   # OT 核心算法（操作转换与协调）
├── pkg/unio/        # 统一与排序（时间、版本）
├── pkg/textor/      # 文本处理（Rope/Piece Table）
├── pkg/fabric/      # 文档织物结构
├── pkg/ai/          # AI 集成
├── pkg/weave/       # 编织引擎
├── pkg/flux/        # 数据流与同步
└── pkg/store/       # 持久化存储
```

## 🏛️ 命名说明（全部使用拉丁语）

- **Texere** (拉丁语)：编织 - 主项目名
- **Concordia** (拉丁语)：和谐 - OT 操作协调
- **Unio** (拉丁语)：统一 - 时间与版本统一
- **Textor** (拉丁语)：编织者 - 文本处理
- **Fabric** (拉丁语)：织物 - 文档结构
- **Weave** (英语)：编织 - 核心引擎
- **Flux** (拉丁语)：流动 - 数据流与同步

## 📚 文档

- [架构文档](docs/architecture/README.md)
- [API 文档](docs/api/README.md)
- [使用指南](docs/guides/getting-started.md)
- [研究文档](docs/research/README.md)

## 🤝 贡献

欢迎贡献！请参阅 [CONTRIBUTING.md](CONTRIBUTING.md)

## 📄 许可证

MIT License

---

**Texere - Weave Knowledge Together** 🧵✨
