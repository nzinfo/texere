# 架构重组分析：编辑功能的组织

> **日期**: 2026-01-31
> **问题**: undo/redo, selection, history, savepoint 等编辑功能应该放在哪里？

---

## 📊 当前架构分析

### 现有分层

```
pkg/ot/          ← OT 算法层
├── operation.go
├── transform.go
├── compose.go
├── client.go
└── undo_manager.go  ← 🔴 编辑功能

pkg/rope/        ← 数据结构层
├── rope.go
├── insert.go
├── delete.go
├── selection.go      ← 🔴 编辑功能
├── history.go        ← 🔴 编辑功能
└── savepoint*.go    ← 🔴 编辑功能

pkg/concordia/  ← 文档接口层
├── document.go
└── string_document.go
```

### 问题识别

1. **编辑功能分散**:
   - UndoManager 在 `pkg/ot`
   - Selection 在 `pkg/rope`
   - History 在 `pkg/rope`
   - SavePoint 在 `pkg/rope`

2. **职责混乱**:
   - `pkg/ot` 混合了算法和编辑功能
   - `pkg/rope` 混合了数据结构和编辑功能
   - `pkg/concordia` 只有接口，没有具体实现

3. **依赖关系不清晰**:
   - UndoManager (ot) 依赖 Operation (ot)
   - Selection (rope) 依赖 PositionMapper (rope)
   - History (rope) 依赖 Transaction (rope)

---

## 🎯 架构重组方案

### 方案 A: 三层架构（推荐）⭐⭐⭐

```
pkg/ot/          ← 纯算法层
├── operation.go      # Operation 定义
├── transform.go      # Transform 算法
├── compose.go        # Compose 算法
└── client.go         # Client 算法

pkg/concordia/  ← 编辑/协作层（统一管理编辑功能）
├── document.go               # Document 接口
├── string_document.go        # String 实现
├── rope_document.go          # Rope 适配器
├── undo_manager.go           # 从 ot 移入
├── selection.go              # 从 rope 移入
├── history.go                # 从 rope 移入
├── savepoint.go              # 从 rope 移入
├── savepoint_enhanced.go     # 从 rope 移入
├── collaborative_document.go # 新增：协作文档
└── resolver.go               # 新增：冲突解决

pkg/rope/        ← 纯数据结构层
├── rope.go                  # 核心 Rope
├── node.go                  # 节点定义
├── insert.go                # 插入操作
├── delete.go                # 删除操作
├── split.go                 # 分割
├── concat.go                # 拼接
├── balance.go               # 平衡
├── position_mapper.go       # 位置映射
└── transaction.go           # 事务
```

**职责划分**:
- `pkg/ot`: 纯粹的 OT 算法，无状态
- `pkg/concordia`: 编辑功能（使用 ot + rope），有状态
- `pkg/rope`: 纯数据结构，无编辑语义

**优点**:
- ✅ 职责清晰：算法、编辑、数据分离
- ✅ concordia 成为真正的"协作层"
- ✅ rope 保持纯粹的数据结构
- ✅ 符合"从 OT 到 Rope"的依赖方向

**缺点**:
- ❌ 需要大量文件移动
- ❌ 破坏现有代码组织

---

### 方案 B: 两层架构

```
pkg/ot/          ← 算法 + 基础编辑
├── operation.go
├── transform.go
├── compose.go
├── client.go
└── undo_manager.go    # 保持不动

pkg/concordia/  ← 高级编辑功能（基于 rope）
├── document.go
├── string_document.go
├── rope_document.go
├── selection.go        # 从 rope 移入
├── history.go          # 从 rope 移入
├── savepoint.go        # 从 rope 移入
└── collaborative.go    # 新增

pkg/rope/        ← 数据结构 + 基础工具
├── rope.go
├── node.go
├── insert.go
├── delete.go
├── split.go
├── concat.go
├── balance.go
├── position_mapper.go
└── transaction.go
```

**优点**:
- ✅ 变动较小（UndoManager 不移动）
- ✅ concordia 获得高级编辑功能

**缺点**:
- ❌ undo_manager 在 ot，但其他编辑功能在 concordia，不一致
- ❌ ot 仍然包含"编辑"语义

---

### 方案 C: 保持现状 + 重新命名

```
pkg/ot/          ← 改名 pkg/algorithm？
pkg/concordia/  ← 改名 pkg/editor？
pkg/rope/        ← 保持不动
```

**优点**:
- ✅ 无需移动文件

**缺点**:
- ❌ 只是改名，没有解决架构问题
- ❌ 名称不能准确反映职责

---

## 🏆 推荐：方案 A（三层架构）

### 理由

1. **单一职责原则**:
   - `pkg/ot` - 纯算法，无状态
   - `pkg/concordia` - 编辑逻辑，有状态
   - `pkg/rope` - 纯数据结构

2. **依赖方向清晰**:
   ```
   concordia (编辑层)
       ↓ 使用
   rope (数据层)
       ↓ 使用
   ot (算法层)
   ```

3. **符合命名语义**:
   - **Concordia** = 协调/和谐 → 协作编辑
   - **OT** = 操作转换 → 纯算法
   - **Rope** = 绳索 → 数据结构

### 具体迁移计划

#### 第一阶段：移动 UndoManager (简单)
```
pkg/ot/undo_manager.go → pkg/concordia/undo_manager.go
```
**原因**: UndoManager 是编辑功能，不是纯算法
**影响**:
- `import "github.com/coreseekdev/texere/pkg/ot"` → `pkg/concordia`
- 更新所有引用

#### 第二阶段：移动 Selection (中等)
```
pkg/rope/selection.go → pkg/concordia/selection.go
```
**原因**: Selection 是编辑状态，不是数据结构
**依赖**: Selection 依赖 Rope 和 PositionMapper
**影响**:
- 需要在 concordia 中导入 rope
- 保持 Selection 的 Rope 适配器

#### 第三阶段：移动 History (复杂)
```
pkg/rope/history.go → pkg/concordia/history.go
```
**原因**: History 是编辑历史，不是数据结构
**依赖**: History 依赖 Transaction (rope)
**影响**:
- 需要在 concordia 中导入 rope
- 可能需要抽象 History 的存储接口

#### 第四阶段：移动 SavePoint (中等)
```
pkg/rope/savepoint.go → pkg/concordia/savepoint.go
pkg/rope/savepoint_enhanced.go → pkg/concordia/savepoint_enhanced.go
```
**原因**: SavePoint 是编辑快照，不是数据结构
**依赖**: SavePoint 依赖 Rope
**影响**:
- 需要在 concordia 中导入 rope
- 保持 RopeDocument 适配器

---

## 📋 文件清单

### 需要移动的文件

| 当前位置 | 目标位置 | 代码行数 | 复杂度 |
|---------|---------|---------|--------|
| `pkg/ot/undo_manager.go` | `pkg/concordia/` | 355 | 🟢 低 |
| `pkg/rope/selection.go` | `pkg/concordia/` | 316 | 🟡 中 |
| `pkg/rope/history.go` | `pkg/concordia/` | 851 | 🔴 高 |
| `pkg/rope/savepoint.go` | `pkg/concordia/` | 184 | 🟡 中 |
| `pkg/rope/savepoint_enhanced.go` | `pkg/concordia/` | 724 | 🟡 中 |
| **总计** | - | **2430** | - |

### 保持不动的文件

| 文件 | 位置 | 原因 |
|------|------|------|
| `operation.go`, `transform.go`, `compose.go` | `pkg/ot/` | 纯算法 |
| `rope.go`, `insert.go`, `delete.go`, ... | `pkg/rope/` | 纯数据结构 |
| `document.go`, `string_document.go` | `pkg/concordia/` | 已在正确位置 |

---

## 🔄 依赖关系图

### 迁移前
```
应用层
  ↓
pkg/rope (Selection, History, SavePoint)
  ↓               ↓
pkg/ot (UndoManager) ─── rope.go
```

### 迁移后（方案 A）
```
应用层
  ↓
pkg/concordia
  ├── UndoManager
  ├── Selection      ───┐
  ├── History        ───┤
  ├── SavePoint      ───┤
  └── Document      ───┘
  ↓                  ↓
pkg/rope ───────────────┘
  ↓
pkg/ot (Operation, Transform, Compose)
```

---

## ⚖️ 利弊权衡

### 重组的优点

1. **职责清晰**:
   - `pkg/ot`: 无状态的算法
   - `pkg/concordia`: 有状态的编辑
   - `pkg/rope`: 无状态的数据结构

2. **语义准确**:
   - Concordia = 协作编辑层
   - 包含所有编辑相关功能

3. **易于扩展**:
   - 未来添加 CollaborativeDocument
   - 统一的编辑接口

### 重组的缺点

1. **破坏性变更**:
   - 大量 import 路径更改
   - API 用户需要更新代码

2. **循环依赖风险**:
   - concordia 导入 rope（数据结构）
   - 需要避免 rope 导入 concordia

3. **工作量**:
   - 移动 ~2430 行代码
   - 更新数百个 import
   - 完整的测试验证

---

## 💡 我的建议

### 短期（立即执行）
1. **保持现状** - 当前架构虽然不完美，但功能完整
2. **添加文档** - 在 README 中说明各包的职责
3. **添加示例** - 展示如何在应用层组合使用

### 中期（可选）
如果项目还在早期阶段，可以考虑重组：
1. **先做实验** - 创建新分支尝试方案 A
2. **评估影响** - 评估破坏性变更的影响
3. **征求反馈** - 如果有其他使用者，征求他们的意见

### 长期（理想状态）
实现方案 A 的三层架构，但要：
1. 保留旧的 import 作为别名（向后兼容）
2. 分阶段迁移，不是一次性
3. 充分的测试覆盖

---

## 🎯 总结

### 当前架构的问题
- ✅ 功能完整，但组织分散
- ❌ 编辑功能跨三个包
- ❌ 职责边界模糊

### 重组的收益
- ✅ pkg/concordia 成为真正的"协作层"
- ✅ 职责清晰：算法、编辑、数据分离
- ✅ 更好的可维护性

### 重组的成本
- ❌ 破坏性变更（API 用户受影响）
- ❌ 大量文件移动和 import 更新
- ❌ 需要完整的回归测试

### 最终建议
**建议**: 如果这是**个人项目**或**早期阶段**，考虑重组（方案 A）。
**建议**: 如果已有**外部用户**或**生产使用**，保持现状，在文档中说明架构。

---

**报告版本**: 1.0
**创建日期**: 2026-01-31
**状态**: 等待决策
