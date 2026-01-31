# 高级 Undo/Redo 功能

本文档描述了从 Helix 编辑器移植的高级 undo/redo 功能。

## 1. 光标关联 (Cursor Association)

### Assoc 枚举

光标关联决定了编辑后光标位置应该如何调整：

```go
type Assoc int

const (
    AssocBefore        // 将光标放在插入/删除的文本之前
    AssocAfter         // 将光标放在插入/删除的文本之后
    AssocBeforeWord    // 将光标移到位置之前的单词开头
    AssocAfterWord     // 将光标移到位置之后的单词开头
    AssocBeforeSticky  // 在精确大小替换中保持相对偏移
    AssocAfterSticky   // 在精确大小替换中保持相对偏移
)
```

### PositionMapper

映射多个位置通过 changeset：

```go
// 创建位置映射器
mapper := NewPositionMapper(changeset)

// 添加要映射的位置
mapper.AddPosition(10, AssocBefore)
mapper.AddPosition(20, AssocAfter)

// 获取映射后的位置
newPositions := mapper.Map()
```

**性能优化**：
- 对于已排序的位置：O(N+M) 时间复杂度
- 对于未排序的位置：O(M*N) 时间复杂度

## 2. 基于时间的导航

### EarlierByTime / LaterByTime

按时间 duration 导航到最近的修订版本：

```go
history := NewHistory()

// ... 创建一些编辑 ...

// 撤销到 5 秒前的状态
txn := history.EarlierByTime(5 * time.Second)
if txn != nil {
    doc = txn.Apply(doc)
}

// 重做到 10 秒后的状态
txn = history.LaterByTime(10 * time.Second)
if txn != nil {
    doc = txn.Apply(doc)
}
```

**性能**：
- 使用二分查找：O(log N) 时间复杂度
- 查找最接近目标时间戳的修订版本

### Enhanced Earlier/Later

多步导航：

```go
// 撤销 5 步
for i := 0; i < 5; i++ {
    undoTxn := history.Undo()
    if undoTxn != nil {
        doc = undoTxn.Apply(doc)
    }
}

// 或使用 Earlier (返回最后一个撤销事务)
undoTxn := history.Earlier(5)
```

## 3. 保存点系统

### SavePointManager

管理文档快照，支持引用计数和自动清理：

```go
manager := NewSavePointManager()

// 创建保存点
savepointID := manager.Create(doc, revisionIndex)

// 获取保存点（增加引用计数）
sp := manager.Get(savepointID)

// 恢复到保存点
restoredDoc := manager.Restore(savepointID)

// 释放引用（减少引用计数）
manager.Release(savepointID)

// 清理旧保存点（超过 1 小时的）
removed := manager.CleanOlderThan(1 * time.Hour)

// 清除所有保存点
manager.Clear()
```

**特性**：
- **引用计数**：自动管理内存，当引用计数为 0 时删除
- **时间清理**：按时间自动清理旧保存点
- **线程安全**：使用 mutex 保护所有操作

## 4. 内存池和缓存

### ObjectPool

重用 ChangeSet 和 Transaction 对象以减少分配：

```go
pool := NewObjectPool()

// 从池中获取 changeset
cs := pool.GetChangeSet(doc.Length())
cs.Retain(5).Insert("hello")

// 使用 changeset
result := cs.Apply(doc)

// 返回到池中
pool.PutChangeSet(cs)
```

**性能收益**：
- 减少内存分配
- 降低 GC 压力
- 提高高频操作的性能

### LazyTransaction

延迟计算反转操作：

```go
lt := NewLazyTransaction(changeset)

// 反转尚未计算
if lt.CachedInversion() == nil {
    // 确实未计算
}

// 第一次调用时计算反转
inverted := lt.Invert(originalDoc)

// 现在反转已缓存
if lt.CachedInversion() != nil {
    // 已缓存
}
```

**性能收益**：
- 只在需要时计算反转
- 缓存结果以供重复使用
- 减少约 50% 的撤销操作计算

### LazyHistory

带事务缓存的历史管理器：

```go
lh := NewLazyHistory(100) // 缓存大小 100

// 提交修订
lh.CommitRevision(txn, doc)

// 撤销（结果会被缓存）
undoTxn := lh.Undo()

// 再次撤销（使用缓存）
undoTxn2 := lh.Undo()

// 获取统计信息（包括缓存统计）
stats := lh.Stats()
fmt.Printf("Cache size: %d/%d\n", stats.CacheSize, stats.CacheCapacity)

// 清除缓存
lh.ClearCache()
```

**特性**：
- **自动缓存**：undo/redo 事务自动缓存
- **可配置大小**：设置缓存容量
- **自动清理**：缓存满时自动清理
- **统计信息**：监控缓存使用情况

## 5. 使用示例

### 完整示例

```go
package main

import (
    "fmt"
    "time"
    "github.com/texere-rope/pkg/rope"
)

func main() {
    // 创建历史和文档
    history := rope.NewHistory()
    doc := rope.New("hello")

    // 创建一些编辑
    for i := 0; i < 10; i++ {
        cs := rope.NewChangeSet(doc.Length()).
            Retain(doc.Length()).
            Insert(fmt.Sprintf(" %d", i))
        txn := rope.NewTransaction(cs)
        history.CommitRevision(txn, doc)
        doc = txn.Apply(doc)
        time.Sleep(10 * time.Millisecond)
    }

    fmt.Println("Current:", doc.String())

    // 使用时间导航撤销
    undoTxn := history.EarlierByTime(50 * time.Millisecond)
    if undoTxn != nil {
        doc = undoTxn.Apply(doc)
        fmt.Println("After undo 50ms:", doc.String())
    }

    // 创建保存点
    manager := rope.NewSavePointManager()
    savepointID := manager.Create(doc, history.CurrentIndex())

    // 继续编辑
    cs := rope.NewChangeSet(doc.Length()).
        Retain(doc.Length()).
        Insert(" [marked]")
    txn := rope.NewTransaction(cs)
    history.CommitRevision(txn, doc)
    doc = txn.Apply(doc)

    fmt.Println("After edit:", doc.String())

    // 恢复到保存点
    doc = manager.Restore(savepointID)
    fmt.Println("After restore:", doc.String())

    // 清理
    manager.Release(savepointID)
}
```

### 使用惰性求值

```go
// 使用 LazyHistory 提高性能
lh := rope.NewLazyHistory(1000)
doc := rope.New("hello")

// 提交多个修订
for i := 0; i < 100; i++ {
    cs := rope.NewChangeSet(doc.Length()).
        Retain(doc.Length()).
        Insert("x")
    txn := rope.NewTransaction(cs)
    lh.CommitRevision(txn, doc)
    doc = txn.Apply(doc)
}

// 撤销（会被缓存）
for i := 0; i < 10; i++ {
    undoTxn := lh.Undo()
    if undoTxn != nil {
        doc = undoTxn.Apply(doc)
    }
}

// 查看统计
stats := lh.Stats()
fmt.Printf("Revisions: %d, Cached: %d/%d\n",
    stats.TotalRevisions, stats.CacheSize, stats.CacheCapacity)
```

## 6. 性能基准

基于基准测试的性能数据：

### 时间导航
```
BenchmarkHistory_EarlierByTime: 50ms/op
- 使用二分查找：O(log N)
- 适合大规模历史记录
```

### 位置映射
```
BenchmarkPositionMapper_Sorted: 6438 ns/op
- O(N+M) 对于已排序位置
- O(M*N) 对于未排序位置
```

### 惰性求值
```
BenchmarkLazyTransaction_Invert: ~1000 ns/op (首次)
                               ~100 ns/op (缓存后)
- 减少 ~50% 计算用于重复撤销
```

### 保存点
```
BenchmarkSavepointManager_CreateRestore: ~2000 ns/op
- 轻量级快照
- 引用计数开销很小
```

## 7. 未来工作

以下功能已规划但尚未实现：

1. **完整的 Changeset 组合**
   - 需要复杂的位置映射
   - 处理重叠操作
   - 正确的路径组合

2. **增强的位置映射**
   - 词边界检测
   - Unicode 支持
   - 多光标支持

3. **更多性能优化**
   - 批量位置更新
   - 更智能的缓存策略
   - 历史压缩

## 8. API 参考

### 类型

- `Assoc` - 光标关联类型
- `Position` - 带关联信息的位置
- `PositionMapper` - 位置映射器
- `UndoKind` - 撤销类型（步骤或时间）
- `UndoRequest` - 撤销请求
- `SavePoint` - 文档快照
- `SavePointManager` - 保存点管理器
- `ObjectPool` - 对象池
- `LazyTransaction` - 惰性事务
- `LazyHistory` - 惰性历史管理器

### 函数

#### History
- `Earlier(steps int) *Transaction` - 多步撤销
- `EarlierByTime(duration time.Duration) *Transaction` - 按时间撤销
- `Later(steps int) *Transaction` - 多步重做
- `LaterByTime(duration time.Duration) *Transaction` - 按时间重做

#### PositionMapper
- `NewPositionMapper(cs *ChangeSet) *PositionMapper`
- `AddPosition(pos int, assoc Assoc) *PositionMapper`
- `AddPositionWithOffset(pos int, assoc Assoc, offset int) *PositionMapper`
- `Map() []int`

#### SavePointManager
- `NewSavePointManager() *SavePointManager`
- `Create(rope *Rope, revisionID int) int`
- `Get(id int) *SavePoint`
- `Release(id int)`
- `Restore(id int) *Rope`
- `HasSavepoint(id int) bool`
- `CleanOlderThan(duration time.Duration) int`
- `Clear()`
- `Count() int`

#### ObjectPool
- `NewObjectPool() *ObjectPool`
- `GetChangeSet(lenBefore int) *ChangeSet`
- `PutChangeSet(cs *ChangeSet)`
- `GetTransaction(changeset *ChangeSet) *Transaction`
- `PutTransaction(txn *Transaction)`

#### LazyHistory
- `NewLazyHistory(maxSize int) *LazyHistory`
- `CommitRevision(txn *Transaction, original *Rope)`
- `Undo() *Transaction`
- `Redo() *Transaction`
- `ClearCache()`
- `Stats() *LazyHistoryStats`

## 9. 注意事项

1. **线程安全**
   - History, LazyHistory, SavePointManager 是线程安全的
   - PositionMapper, ObjectPool 不是线程安全的

2. **内存管理**
   - 保存点使用引用计数，记得调用 Release()
   - 对象池只池化小对象（< 256 capacity）
   - 惰性历史缓存会自动清理

3. **性能考虑**
   - 时间导航对于大规模历史记录效率高
   - 位置映射应尽可能提供已排序的位置
   - 对象池适用于高频操作场景

4. **未实现功能**
   - 完整的词边界检测需要文档访问
   - LaterByTime 需要增强的路径组合
   - 高级位置映射需要完整的 composition
