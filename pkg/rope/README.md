# Rope 包 - 高效文本数据结构

[![Go Documentation](https://pkg.go.dev/badge/github.com/coreseekdev/texere/pkg/rope.svg)](https://pkg.go.dev/github.com/coreseekdev/texere/pkg/rope)

Rope 是一种专为大型文本编辑设计的高效字符串数据结构。它使用平衡二叉树（B-tree）来表示字符串，针对频繁插入、删除操作进行了优化。

## 特性

- **不可变性 (Immutable)**: 所有操作返回新的 Rope，原对象保持不变
- **高效性**: 插入、删除、切片操作均为 O(log n) 复杂度
- **内存优化**: 树结构最小化内存拷贝
- **线程安全**: 不可变特性天然支持并发访问
- **Unicode 支持**: 完整支持 UTF-8、字素簇、单词边界

## 快速开始

```go
import "github.com/coreseekdev/texere/pkg/rope"

// 创建 Rope
r := rope.New("Hello, World!")

// 插入文本
r = r.Insert(13, " Have a nice day!")

// 删除文本
r = r.Delete(0, 7)

// 获取内容
fmt.Println(r.String()) // "World! Have a nice day!"
```

## 目录结构

### 核心文件

| 文件 | 说明 |
|------|------|
| `rope.go` | 核心 Rope 数据结构和基本操作 |
| `changeset.go` | 内部编辑表示 (ChangeSet, Operation) |
| `edits.go` | 编辑操作类型 (EditOperation, Deletion) |
| `selection.go` | 选择范围管理 |
| `composition.go` | ChangeSet 组合逻辑 |
| `position.go` | 光标位置映射和关联 |

### 操作实现

| 文件 | 说明 |
|------|------|
| `char_ops.go` | 字符操作 |
| `chunk_ops.go` | 块操作 |
| `line_ops.go` | 行操作 |

### 迭代器

| 文件 | 说明 |
|------|------|
| `iterator.go` | 字符迭代器 (Rune Iterator) |
| `bytes_iter.go` | 字节迭代器 (Bytes Iterator) |
| `runes_iter.go` | Rune 迭代器 |
| `reverse_iter.go` | 反向迭代器 |

### 优化

| 文件 | 说明 |
|------|------|
| `cow_optimization.go` | 写时复制优化 |
| `insert_optimized.go` | 插入优化 |
| `micro_optimizations.go` | 微优化 |
| `byte_cache.go` | 字节位置缓存 |
| `pools.go` | 对象池 |
| `hash.go` | 哈希工具 |

### 文本处理

| 文件 | 说明 |
|------|------|
| `graphemes.go` | 字素簇处理 |
| `word_boundary.go` | 单词边界检测 |
| `utf16.go` | UTF-16 支持 |
| `crlf.go` | 换行符处理 |

### 工具

| 文件 | 说明 |
|------|------|
| `builder.go` | Rope 构建器 |
| `str_utils.go` | 字符串工具 |
| `balance.go` | 树平衡 |
| `rope_concat.go` | 拼接操作 |
| `rope_split.go` | 分割操作 |
| `rope_io.go` | I/O 操作 |

### 测试文件

| 文件 | 说明 |
|------|------|
| `core_test.go` | 核心功能测试 |
| `property_test.go` | 基于属性的测试 |
| `tree_integrity_test.go` | 树完整性测试 |
| `chunk_test.go` | 块操作测试 |
| `*_test.go` | 其他功能测试 |
| `*_bench_test.go` | 基准测试 |

### 文档

| 文件 | 说明 |
|------|------|
| `README.md` | 本文件 |
| `USAGE.md` | 详细使用指南 |
| `doc/FILE_CLASSIFICATION.md` | 文件分类说明 |
| `doc/OPTIMIZATION_EVALUATION.md` | 优化评估 |

## 适用场景

### 适合使用 Rope 的场景

- 大型文本编辑器（如 Helix、Kakoune）
- 需要频繁插入/删除的文本处理
- 需要高效撤销/重做功能的应用
- 多线程文本处理场景
- 需要频繁切片操作的文本

### 不适合使用 Rope 的场景

- 小型字符串（Go 原生 `string` 更高效）
- 只读文本（无需 Rope 的复杂结构）
- 频繁转换为字符串的场景

## 性能

基于当前实现的性能数据（Go 1.23）：

| 操作 | 复杂度 | 说明 |
|------|--------|------|
| `Length()` | O(1) | 缓存的长度 |
| `Slice()` | O(log n) | 需要遍历树 |
| `Insert()` | O(log n) | 创建新节点 |
| `Delete()` | O(log n) | 创建新节点 |
| `String()` | O(n) | 需要遍历所有节点 |
| `Iterator` | O(1) 每字符 | 块迭代器 |

详细性能数据请参考 `USAGE.md` 中的性能基准章节。

## 架构设计

### 依赖关系

```
                    ┌──────────────┐
                    │  concordia   │  (OT 集成层)
                    │   包         │
                    └──────┬───────┘
                           │ 依赖 ot 包
                           ▼
                    ┌────────────────┐
                    │   rope 包      │  (核心层)
                    │   (无 ot 依赖)  │
                    └────────────────┘
```

### 核心概念

1. **不可变性**: 所有操作返回新 Rope，原 Rope 保持不变
2. **树结构**: 使用平衡二叉树（B-tree）存储文本
3. **叶子节点**: 包含实际文本内容
4. **内部节点**: 包含左右子树和缓存信息

## 更多资源

- [USAGE.md](USAGE.md) - 详细使用指南和示例
- [doc/FILE_CLASSIFICATION.md](doc/FILE_CLASSIFICATION.md) - 文件分类和依赖说明
- [GoDoc 文档](https://pkg.go.dev/github.com/coreseekdev/texere/pkg/rope)

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
