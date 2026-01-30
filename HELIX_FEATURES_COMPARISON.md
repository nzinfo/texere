# Helix Rope/Transaction 功能对比分析

## 核心数据结构

### Helix
- `Operation` 枚举：Retain, Delete, Insert
- `ChangeSet` 结构：包含 operations, len, len_after
- `Transaction` 结构：包含 changeset 和 selection
- `Assoc` 枚举：Before, After, AfterWord, BeforeWord, BeforeSticky, AfterSticky
- `ChangeIterator`：用于迭代 changeset 的操作

### Go 实现
- ✅ `Operation` 结构：OpRetain, OpDelete, OpInsert
- ✅ `ChangeSet` 结构：包含 operations, lenBefore, lenAfter
- ✅ `Transaction` 结构：包含 changeset 和 timestamp
- ✅ `Assoc` 类型：包含所有 6 种模式（甚至更多）
- ❌ `ChangeIterator`：未实现

## ChangeSet 核心功能

### Helix ChangeSet API
| 函数 | 状态 | 说明 |
|------|------|------|
| `new(doc)` | ✅ | 已实现为 `NewChangeSet(len)` |
| `with_capacity(cap)` | ✅ | 已实现 |
| `changes()` | ✅ | 返回 operations 切片 |
| `compose(other)` | ✅ | 完整实现，所有测试通过 |
| `map(other)` | N/A | Helix 标记为 unimplemented! |
| `invert(original)` | ✅ | 已实现 |
| `apply(text)` | ✅ | 已实现 |
| `is_empty()` | ✅ | 已实现 |
| `update_positions()` | ✅ | 通过 PositionMapper 实现 |
| `map_pos()` | ✅ | 已实现为 MapPosition |
| `changes_iter()` | ❌ | **缺失：ChangeIterator** |
| `len_chars()` | N/A | Go 使用 Length() 方法 |

### 高级 Position Mapping 功能

#### Helix 实现
- `update_positions()` - 批量映射位置
- `map_pos()` - 单个位置映射
- 支持所有 Assoc 模式的精确语义
- 处理未排序位置的回溯逻辑
- 处理 replace 操作（Insert + Delete 组合）

#### Go 实现
- ✅ `MapPosition(pos, assoc)` - 单个位置映射
- ✅ `MapPositions(positions, assocs)` - 批量位置映射
- ✅ `PositionMapper` - 完整的 mapper 实现
- ✅ 支持所有 6 种 Assoc 模式
- ⚠️ 可能缺少：未排序位置的处理
- ⚠️ 可能缺少：replace 的特殊处理（需要验证）

## Transaction 功能

### Helix Transaction API
| 函数 | 状态 | 说明 |
|------|------|------|
| `new(doc)` | ✅ | 已实现为 `NewTransaction(changeset)` |
| `changes()` | ✅ | 已实现 |
| `selection()` | ⚠️ | 返回 `time.Time` 而非 Selection |
| `apply(doc)` | ✅ | 已实现 |
| `invert(original)` | ✅ | 已实现 |
| `compose(other)` | ✅ | 已实现 |
| `with_selection(sel)` | ❌ | **缺失：selection 支持** |
| `change_ignore_overlapping()` | ❌ | **缺失** |
| `change(doc, changes)` | ❌ | **缺失** |
| `delete(doc, deletions)` | ❌ | **缺失** |
| `insert_at_eof(text)` | ❌ | **缺失** |
| `change_by_selection()` | ❌ | **缺失** |
| `delete_by_selection()` | ❌ | **缺失** |
| `insert(doc, selection, text)` | ❌ | **缺失** |

## 缺失的关键功能

### 1. Selection 支持 ⚠️ 重要
**Helix:**
```rust
pub struct Transaction {
    changes: ChangeSet,
    selection: Option<Selection>,  // 支持 multiple cursors
}
```

**Go:**
```go
type Transaction struct {
    changeset  *ChangeSet
    timestamp  time.Time  // ❌ 只有时间戳，没有 selection
}
```

**影响:**
- 无法跟踪光标位置
- 无法处理多重选择
- 无法实现基于选择的操作

### 2. ChangeIterator ⚠️ 重要
**Helix:**
```rust
pub fn changes_iter(&self) -> ChangeIterator<'_>
```

**影响:**
- 无法高效迭代 changeset 的操作
- 某些高级功能可能需要此功能

### 3. 便捷构造方法 ⚠️ 中等重要

#### change() - 从一组变化创建 Transaction
```rust
pub fn change<I>(doc: &Rope, changes: I) -> Self
where I: Iterator<Item = Change>
// Change = (from, to, Option<Tendril>)
```

#### delete() - 从一组删除创建 Transaction
```rust
pub fn delete<I>(doc: &Rope, deletions: I) -> Self
where I: Iterator<Item = Deletion>
// Deletion = (usize, usize)
```

#### change_by_selection() - 为每个选择范围应用变化
```rust
pub fn change_by_selection<F>(doc: &Rope, selection: &Selection, f: F) -> Self
where F: FnMut(&Range) -> Change
```

#### insert() - 在所有选择位置插入文本
```rust
pub fn insert(doc: &Rope, selection: &Selection, text: Tendril)
```

#### delete_by_selection() - 从选择删除
```rust
pub fn delete_by_selection<F>(doc: &Rope, selection: &Selection, f: F) -> Self
```

### 4. change_ignore_overlapping() ⚠️ 低优先级
处理可能重叠的变化，自动忽略重叠部分。

## 高级功能对比

### History/Undo-Redo
| 功能 | Helix | Go |
|------|-------|-----|
| Undo/Redo 栈 | ✅ | ✅ LazyHistory |
| Savepoints | ✅ | ✅ SavepointManager |
| 分支历史 | ✅ | ✅ |
| 时间导航 | ✅ | ✅ Earlier/Later |

### 性能优化
| 功能 | Helix | Go |
|------|-------|-----|
| Operation Fusion | ✅ | ✅ |
| Lazy Inversion | ✅ | ✅ LazyTransaction |
| Object Pooling | ✅ | ✅ ObjectPool |

## 推荐的优先级

### P0 - 必须实现（核心功能缺失）
1. **Selection 支持** - 与光标位置和多重选择相关
   - 添加 `Selection` 类型
   - Transaction 添加 `selection` 字段
   - 实现 `with_selection()` 方法

2. **change() / delete() 便捷方法** - 创建 Transaction 的常用方式
   - `Change(changes []Change) Transaction`
   - `Delete(deletions []Delete) Transaction`

### P1 - 重要但非紧急
1. **ChangeIterator** - 某些高级功能需要
2. **change_by_selection()** - 基于选择的批量操作
3. **insert()** - 多重选择插入

### P2 - 锦上添花
1. **change_ignore_overlapping()** - 边界情况处理
2. **delete_by_selection()** - 基于选择的删除

## 验证项

需要确认的 Go 实现：
- [ ] PositionMapper 是否正确处理未排序位置
- [ ] PositionMapper 是否正确处理 Insert+Delete (replace)
- [ ] Assoc 的 BeforeWord/AfterWord 是否与 Helix 语义完全一致
- [ ] Assoc 的 BeforeSticky/AfterSticky 是否正确实现

## 总结

### 已完成 ✅
- 核心的 Changeset 组合算法（compose）
- 完整的 invert 功能
- 基础的 position mapping
- 所有 6 种 Assoc 模式
- History/Undo-Redo/Savepoints
- 性能优化（fusion, lazy evaluation, object pooling）

### 缺失但重要 ❌
- **Selection 支持** - 这是最大的缺失
- 便捷的 Transaction 构造方法
- ChangeIterator

### 可选功能
- map() - Helix 自己也没实现
- change_ignore_overlapping() - 边界情况
