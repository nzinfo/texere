# Texere

> **Weave Knowledge Together** - 编织知识，连接智慧

Texere 是一个基于 Operational Transformation (OT) 和 Rope 数据结构的文本编辑核心库。

## 🎯 项目概述

Texere 提供了构建实时协作编辑器和文本编辑器所需的核心组件：

- **Operational Transformation (OT)** - 通过 `concordia` 包实现高效的 OT 算法
- **Rope 数据结构** - 通过 `rope` 包实现高性能的文本操作
- **文档抽象** - 通过 `document` 包提供统一的文档接口

## ✨ 核心特性

### OT (Operational Transformation)
- ✅ 完整的操作转换实现（Insert, Delete, Retain）
- ✅ 操作组合 (Compose)
- ✅ 操作转换 (Transform) - 支持并发编辑冲突解决
- ✅ 操作反转 (Invert) - 支持 Undo/Redo
- ✅ 客户端同步 (Client) - 支持客户端-服务器架构
- ✅ 撤销管理器 (UndoManager) - 带时间戳的撤销/重做

### Rope 数据结构
- ✅ 不可变二叉树结构 - 高效的文本操作
- ✅ 快速插入/删除 - O(log n) 时间复杂度
- ✅ 零拷贝切片 - 高效的文本访问
- ✅ UTF-8 支持 - 完整的 Unicode 支持
- ✅ 字节/字符迭代器 - 灵活的文本遍历
- ✅ 性能优化 - InsertOptimized/DeleteOptimized (比标准实现快 17-35%)
- ✅ 事务支持 - 支持原子操作和位置映射

### 性能
- **插入操作**: InsertOptimized 比 ZeroAlloc 快 **17%**
- **删除操作**: DeleteOptimized 与 ZeroAlloc 相当或更快
- **单叶优化**: InsertFast/DeleteFast 快 **4-16x**
- **内存优化**: 移除了 ZeroAlloc (内存开销减少 97%)

## 📦 包结构

```
texere/
├── pkg/ot/   # OT 核心算法
│   ├── operation.go     # 操作定义和实现
│   ├── builder.go       # 操作构建器
│   ├── transform.go     # 操作转换
│   ├── compose.go       # 操作组合
│   ├── client.go        # 客户端同步
│   └── undo_manager.go  # 撤销管理器
├── pkg/rope/        # Rope 数据结构
│   ├── rope.go          # 核心 Rope 实现
│   ├── insert.go        # 插入操作
│   ├── delete.go        # 删除操作
│   ├── split.go         # 分割操作
│   ├── concat.go        # 拼接操作
│   └── balance.go       # 重新平衡
├── pkg/concordia/    # 文档接口
│   ├── document.go      # Document 接口定义
│   └── string_document.go # String 实现
├── QUICKSTART.md    # OT 快速入门
└── ROPE_QUICKSTART.md  # Rope 快速入门
```

## 🚀 快速开始

### 安装

```bash
go get github.com/coreseekdev/texere
```

### OT 基础使用

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/ot"
)

func main() {
    // 创建插入操作
    op := ot.NewBuilder().
        Insert("Hello, World!").
        Build()

    // 应用到文档
    result, err := op.Apply("")
    if err != nil {
        panic(err)
    }

    fmt.Println(result) // "Hello, World!"
}
```

### Rope 基础使用

```go
package main

import (
    "fmt"
    "github.com/coreseekdev/texere/pkg/rope"
)

func main() {
    // 创建 Rope
    r := rope.New("Hello, World!")

    // 插入文本
    r = r.InsertFast(7, "Beautiful ")

    // 删除文本
    r = r.DeleteFast(16, 25)

    // 获取结果
    fmt.Println(r.String()) // "Hello, Beautiful!"
}
```

## 📚 文档

- **[OT 快速入门](QUICKSTART.md)** - 5 分钟上手 Concordia OT 库
- **[Rope 快速入门](ROPE_QUICKSTART.md)** - Rope 数据结构使用指南
- **[Concordia API](pkg/ot/README.md)** - OT API 文档
- **[Rope API](pkg/rope/README.md)** - Rope API 文档

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行 OT 测试
go test ./pkg/ot/... -v

# 运行 Rope 测试
go test ./pkg/rope/... -v

# 带覆盖率
go test ./... -cover
```

## 🔧 构建

项目使用 [just](https://github.com/casey/just) 作为构建工具：

```bash
# 安装 just
cargo install just

# 查看可用命令
just --list

# 运行测试
just test

# 构建项目
just build

# 清理
just clean
```

## 🏗️ 分支结构

- **master** - 主分支（基于 feature/ot + feature/rope 合并）
- **master-legacy** - 归档的旧分支（混合了多种框架）
- **feature/ot** - OT 实现分支
- **feature/rope** - Rope 性能优化分支

## 📊 性能基准

### Insert 操作
| 实现 | 速度 (ns/op) | 内存 (B/op) |
|------|-------------|-------------|
| InsertFast | 144 | 72 |
| InsertOptimized | 1952 | 2864 |
| Insert (Standard) | 2991 | 880 |

### Delete 操作
| 实现 | 速度 (ns/op) | 内存 (B/op) |
|------|-------------|-------------|
| DeleteFast | 174 | 56 |
| DeleteOptimized | 672 | 2864 |
| Delete (Standard) | 922 | 1456 |

## 🤝 贡献

欢迎贡献！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 开启 Pull Request

## 📄 许可证

MIT License

---

**Texere - Weave Knowledge Together** 🧵✨
