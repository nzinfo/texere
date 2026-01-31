# Rope 使用指南

## 目录

- [简介](#简介)
- [核心概念](#核心概念)
- [基础操作](#基础操作)
- [修改操作](#修改操作)
- [迭代器](#迭代器)
- [构建器](#构建器)
- [行操作](#行操作)
- [性能优化建议](#性能优化建议)
- [完整示例](#完整示例)

---

## 简介

Rope 是一种高效的字符串数据结构，专为大型文本编辑设计。它使用平衡二叉树（B-tree）来表示字符串，提供优于传统字符串的操作性能。

### 主要特性

- **不可变性 (Immutable)**: 所有操作返回新的 Rope，原对象保持不变
- **高效性**: 插入、删除、切片操作均为 O(log n) 复杂度
- **内存优化**: 树结构最小化内存拷贝
- **线程安全**: 不可变特性天然支持并发访问

### 适用场景

- 大型文本编辑器（如 Helix、Kakoune）
- 需要频繁插入/删除的文本处理
- 需要高效撤销/重做功能的应用
- 多线程文本处理场景

### 不适用场景

- 小型字符串（Go 原生 string 更高效）
- 只读文本（无需 Rope 的复杂结构）
- 频繁转换为字符串的场景

---

## 核心概念

### Rope 结构

```go
type Rope struct {
    root   RopeNode  // 树根节点
    length int       // 字符总数（Unicode 码点）
    size   int       // 字节总数
}
```

### 字符位置 vs 字节位置

Rope 中所有位置索引均使用**字符位置**（Unicode 码点），而非字节位置。这是因为 UTF-8 编码中一个字符可能占 1-4 字节。

```go
r := rope.New("Hello 世界")
// 字符位置:   012345 6 7
// 字符:      H e l l o   世 界

// "世" 的字符位置是 6
r.CharAt(6)  // 返回 '世'

// "世" 的字节位置可能是 6（实际需要计算）
r.CharToByte(6)  // 返回 6
```

### 不可变语义

所有修改操作都返回新的 Rope，原 Rope 保持不变：

```go
r1 := rope.New("Hello")
r2 := r1.Insert(5, " World")  // r2 = "Hello World"

// r1 仍然是 "Hello"，未改变
fmt.Println(r1.String())  // "Hello"
fmt.Println(r2.String())  // "Hello World"
```

---

## 基础操作

### 创建 Rope

```go
// 从字符串创建
r := rope.New("Hello World")

// 创建空 Rope
empty := rope.Empty()

// 从切片创建
text := []string{"Line1", "Line2", "Line3"}
r := rope.New(strings.Join(text, "\n"))
```

### 查询信息

```go
r := rope.New("Hello 世界")

// 字符数（Unicode 码点数）
length := r.Length()  // 8

// 字节数
size := r.Size()  // 12 (5 + 1 + 2*3)

// 转换为字符串
s := r.String()  // "Hello 世界"

// 转换为字节切片
b := r.Bytes()  // []byte("Hello 世界")
```

### 字符访问

```go
r := rope.New("Hello")

// 获取指定位置的字符
ch := r.CharAt(1)  // 'e'

// 获取指定字节位置的字节
b := r.ByteAt(0)  // 'H'
```

### 子串操作

```go
r := rope.New("Hello World")

// 获取子串 [start, end)
// 注意：end 是字符位置，不是字节位置
sub := r.Slice(0, 5)  // "Hello"
sub := r.Slice(6, 11) // "World"

// 字符位置超出范围会 panic
// sub := r.Slice(0, 100) // panic: slice bounds out of range
```

### 搜索操作

```go
r := rope.New("Hello World Hello")

// 检查是否包含子串
contains := r.Contains("World")  // true

// 查找子串首次出现位置（字符位置）
idx := r.Index("World")  // 6
idx := r.Index("xyz")    // -1 (未找到)

// 查找子串最后出现位置
lastIdx := r.LastIndex("Hello")  // 12

// 比较两个 Rope
cmp := r.Compare(otherRope)  // -1, 0, 或 1

// 检查内容是否相同
equal := r.Equals(otherRope)  // true 或 false
```

---

## 修改操作

### 插入文本

```go
r := rope.New("Hello World")

// 在指定字符位置插入文本
r2 := r.Insert(5, " Beautiful")  // "Hello Beautiful World"
r3 := r.Insert(0, "Say: ")       // "Say: Hello World"

// 插入到末尾
r4 := r.Insert(r.Length(), "!")  // "Hello World!"

// 位置超出范围会 panic
// r5 := r.Insert(100, "x")  // panic: insert position out of range
```

### 删除文本

```go
r := rope.New("Hello World")

// 删除 [start, end) 范围的字符
r2 := r.Delete(0, 6)   // "World" (删除 "Hello ")
r3 := r.Delete(5, 6)   // "HelloWorld" (删除 " ")
r4 := r.Delete(6, 11)  // "Hello " (删除 "World")

// 删除单个字符
r5 := r.Delete(5, 6)   // "HelloWorld"

// 范围超出会 panic
// r6 := r.Delete(0, 100)  // panic: delete range out of bounds
```

### 替换文本

```go
r := rope.New("Hello World")

// 替换 [start, end) 范围的文本
r2 := r.Replace(6, 11, "Go")      // "Hello Go"
r3 := r.Replace(0, 5, "Hi")       // "Hi World"

// Replace = Delete + Insert
r4 := r.Delete(6, 11).Insert(6, "Go")  // 等价于上面的 r2
```

### 分割与连接

```go
r := rope.New("Hello World")

// 在指定位置分割为两个 Rope
left, right := r.Split(5)
// left = "Hello"
// right = " World"

// 连接两个 Rope
r2 := left.Concat(right)  // "Hello World"
r3 := r.Concat(rope.New("!!!"))  // "Hello World!!!"

// Clone 返回自身（不可变，无需拷贝）
r4 := r.Clone()  // r4 和 r 是同一个对象
```

---

## 迭代器

### 字符迭代器 (Iterator)

正向遍历 Rope 中的所有字符：

```go
r := rope.New("Hello")

// 创建迭代器
it := r.NewIterator()

// 遍历所有字符
for it.Next() {
    ch := it.Current()  // 当前字符
    pos := it.Position() // 当前位置（字符位置）
    fmt.Printf("%c at %d\n", ch, pos)
}

// 输出:
// H at 1
// e at 2
// l at 3
// l at 4
// o at 5
```

#### 从指定位置开始迭代

```go
r := rope.New("Hello World")

// 从指定字符位置开始
it := r.IteratorAt(6)
for it.Next() {
    fmt.Printf("%c", it.Current())
}
// 输出: World
```

#### 迭代器导航

```go
it := r.NewIterator()

// 前进/后退
it.Next()      // 前进到下一个字符
it.Previous()  // 后退到上一个字符

// 跳过 n 个字符
it.Skip(5)  // 跳过 5 个字符

// 定位到指定位置
it.Seek(10)  // 定位到字符位置 10

// 检查状态
it.HasNext()      // 是否还有下一个字符
it.HasPrevious()  // 是否还有上一个字符
it.IsAtStart()    // 是否在起始位置
it.IsAtEnd()      // 是否在结束位置
it.Remaining()    // 剩余字符数

// 重置迭代器
it.Reset()  // 回到起始位置
```

#### 预览操作

```go
it := r.NewIterator()

// 预览当前字符（不移动迭代器）
ch, ok := it.Peek()  // 查看当前字符
if ok {
    fmt.Printf("Next char: %c\n", ch)
}

// 预览下一个字符
nextCh, ok := it.PeekNext()
```

#### 收集操作

```go
r := rope.New("Hello")

it := r.NewIterator()

// 收集剩余字符为字符串
s := it.Collect()  // "Hello"

// 收集为 rune 切片
runes := it.CollectToSlice()  // []rune{'H','e','l','l','o'}
```

### 字节迭代器 (BytesIterator)

遍历 Rope 的字节：

```go
r := rope.New("Hello")

it := r.NewBytesIterator()

for it.Next() {
    b := it.Current()      // 当前字节
    pos := it.Position()   // 当前字节位置
    fmt.Printf("%c at byte %d\n", b, pos)
}
```

### 反向迭代器 (ReverseIterator)

从后往前遍历字符：

```go
r := rope.New("Hello")

it := r.NewReverseIterator()

for it.Next() {
    ch := it.Current()  // 'o', 'l', 'l', 'e', 'H'
    fmt.Printf("%c", ch)
}
// 输出: olleH
```

### 块迭代器 (ChunksIterator)

遍历 Rope 的叶子节点（文本块）：

```go
r := rope.New("Hello World")

it := r.Chunks()

for it.Next() {
    chunk := it.Current()  // 当前块的文本
    info := it.Info()      // 当前块的详细信息
    fmt.Printf("Chunk: %s (bytes: %d, chars: %d)\n",
        chunk, info.ByteLen, info.CharLen)
}
```

### 函数式操作

```go
r := rope.New("Hello World")

// ForEach: 对每个字符执行函数
r.ForEach(func(ch rune) {
    fmt.Printf("%c", ch)
})

// ForEachWithIndex: 带索引的遍历
r.ForEachWithIndex(func(idx int, ch rune) {
    fmt.Printf("Char %d: %c\n", idx, ch)
})

// Map: 转换每个字符
upper := r.Map(func(ch rune) rune {
    if ch >= 'a' && ch <= 'z' {
        return ch - 32  // 转大写
    }
    return ch
})
// upper = "HELLO WORLD"

// Filter: 过滤字符
noSpace := r.Filter(func(ch rune) bool {
    return ch != ' '  // 保留非空格字符
})
// noSpace = "HelloWorld"

// Reduce: 归约操作
count := r.Count(func(ch rune) bool {
    return ch == 'l'  // 统计 'l' 的数量
})
// count = 3

// Any: 是否存在满足条件的字符
hasDigit := r.Any(func(ch rune) bool {
    return ch >= '0' && ch <= '9'
})
// hasDigit = false

// All: 是否所有字符都满足条件
allUpper := r.All(func(ch rune) bool {
    return ch >= 'A' && ch <= 'Z'
})
// allUpper = false
```

---

## 构建器 (RopeBuilder)

RopeBuilder 用于批量构建 Rope，优化多次操作的性能。

### 基础使用

```go
builder := rope.NewBuilder()

// 追加文本
builder.Append("Hello")
builder.Append(" ")
builder.Append("World")

// 构建最终的 Rope
r := builder.Build()
// r = "Hello World"
```

### 插入操作

```go
builder := rope.NewBuilder()

builder.Append("Hello World")
builder.Insert(5, " Beautiful")
builder.Insert(0, "Say: ")

r := builder.Build()
// r = "Say: Hello Beautiful World"
```

### 删除和替换

```go
builder := rope.NewBuilder()

builder.Append("Hello World")
builder.Delete(5, 6)        // 删除空格
builder.Replace(0, 5, "Hi") // 替换 "Hello" 为 "Hi"

r := builder.Build()
// r = "HiWorld"
```

### 便捷方法

```go
builder := rope.NewBuilder()

// 追加单个字符
builder.AppendRune('H')
builder.AppendRune('i')

// 追加单个字节
builder.AppendByte('!')

// 追加一行
builder.AppendLine("Hello")

// 写入接口（io.Writer）
var buf bytes.Buffer
builder.Write([]byte("Test"))
builder.WriteString("String")

r := builder.Build()
```

### 构建器复用

```go
builder := rope.NewBuilder()

// 第一次构建
builder.Append("Hello")
r1 := builder.Build()  // builder 保留内容

// 第二次构建（基于第一次的结果）
builder.Append(" World")
r2 := builder.Build()  // "Hello World"

// 重置构建器
builder.Reset()
builder.Append("New text")
r3 := builder.Build()  // "New text"
```

### 性能优势

```go
// 低效方式（每次 Insert 都创建新 Rope）
r := rope.New("")
for i := 0; i < 1000; i++ {
    r = r.Insert(r.Length(), "a")  // 创建 1000 个 Rope 对象
}

// 高效方式（使用 Builder）
builder := rope.NewBuilder()
for i := 0; i < 1000; i++ {
    builder.Append("a")  // 只在 Build() 时创建最终 Rope
}
r := builder.Build()
```

---

## 行操作

### 行信息查询

```go
r := rope.New("Line1\nLine2\nLine3")

// 行数（0-based）
lineCount := r.LineCount()  // 3

// 获取指定行（不包含换行符）
line := r.Line(0)  // "Line1"
line := r.Line(1)  // "Line2"

// 获取指定行（包含换行符）
lineWithEnding := r.LineWithEnding(0)  // "Line1\n"

// 行的起始和结束位置（字符位置）
start := r.LineStart(1)  // 6
end := r.LineEnd(1)      // 11

// 行长度（不包含换行符）
len := r.LineLength(1)  // 5 ("Line2")
```

### 位置转换

```go
r := rope.New("Line1\nLine2\nLine3")

// 字符位置 → 行号
line := r.LineAtChar(7)  // 1 (第 2 行)

// 字符位置 → 列号
col := r.ColumnAtChar(7)  // 1 (第 2 列)

// 行列 → 字符位置
pos := r.PositionAtLineCol(1, 2)  // 8
```

### 行编辑

```go
r := rope.New("Line1\nLine2\nLine3")

// 在指定行首插入文本
r2 := r.InsertLine(1, ">>> ")  // "Line1\n>>> Line2\nLine3"

// 删除指定行
r3 := r.DeleteLine(1)  // "Line1\nLine3"

// 替换指定行
r4 := r.ReplaceLine(1, "NewLine")  // "Line1\nNewLine\nLine3"

// 追加新行
r5 := r.AppendLine("Line4")  // "Line1\nLine2\nLine3\nLine4"

// 在开头插入新行
r6 := r.PrependLine("Line0")  // "Line0\nLine1\nLine2\nLine3"
```

### 行迭代器

```go
r := rope.New("Line1\nLine2\nLine3")

it := r.LinesIterator()

for it.Next() {
    line := it.Current()           // 当前行（不含换行符）
    lineNum := it.LineNumber()     // 当前行号
    fmt.Printf("Line %d: %s\n", lineNum, line)
}

// 收集所有行
lines := it.ToSlice()  // []string{"Line1", "Line2", "Line3"}
```

### 换行符处理

```go
r := rope.New("Line1\nLine2\r\nLine3")

// 检测换行符类型
ending := r.LineEnding()  // "\n", "\r\n", "\r", 或 ""

// 检查是否有尾部换行符
hasTrailing := r.HasTrailingNewline()  // true/false

// 统一换行符风格
unixStyle := r.NormalizeLineEndings("\n")      // Unix 风格
windowsStyle := r.NormalizeLineEndings("\r\n") // Windows 风格

// 删除尾部换行符
trimmed := r.TrimTrailingNewlines()

// 删除头部换行符
trimmed := r.TrimLeadingNewlines()

// 连接所有行为单行
joined := r.JoinLines()  // "Line1Line2Line3"
```

### 缩进操作

```go
r := rope.New("Line1\n  Line2\nLine3")

// 增加缩进
indented := r.IndentLines("  ")
// "  Line1\n    Line2\n  Line3"

// 删除公共缩进
dedented := r.DedentLines()
// "Line1\n  Line2\nLine3"
```

---

## 性能优化建议

### 1. 使用迭代器而非重复访问

**低效：**
```go
for i := 0; i < r.Length(); i++ {
    ch := r.CharAt(i)  // 每次 O(log n)
    // ...
}
// 总复杂度: O(n log n)
```

**高效：**
```go
it := r.NewIterator()
for it.Next() {
    ch := it.Current()  // 每次 O(1)
    // ...
}
// 总复杂度: O(n)
```

### 2. 批量操作使用 Builder

**低效：**
```go
r := rope.New("")
for i := 0; i < 1000; i++ {
    r = r.Insert(r.Length(), "a")  // 每次创建新 Rope
}
```

**高效：**
```go
builder := rope.NewBuilder()
for i := 0; i < 1000; i++ {
    builder.Append("a")
}
r := builder.Build()
```

### 3. 避免频繁 String() 转换

**低效：**
```go
for i := 0; i < r.Length(); i++ {
    if strings.Contains(r.String(), "pattern") {  // 每次都转换
        // ...
    }
}
```

**高效：**
```go
s := r.String()  // 只转换一次
for i := 0; i < r.Length(); i++ {
    if strings.Contains(s, "pattern") {
        // ...
    }
}
```

### 4. 使用 Slice 而非多次 CharAt

**低效：**
```go
// 逐个字符读取
var chars []rune
for i := 0; i < r.Length(); i++ {
    chars = append(chars, r.CharAt(i))
}
```

**高效：**
```go
// 一次性切片
s := r.Slice(0, r.Length())
chars := []rune(s)
```

### 5. 利用不可变特性缓存

```go
// 缓存字符串表示
type CachedRope struct {
    *rope.Rope
    stringCache string
    cacheValid  bool
}

func (cr *CachedRope) String() string {
    if !cr.cacheValid {
        cr.stringCache = cr.Rope.String()
        cr.cacheValid = true
    }
    return cr.stringCache
}
```

### 6. 选择合适的迭代器

```go
// 字符迭代：使用 Iterator
it := r.NewIterator()
for it.Next() {
    ch := it.Current()
}

// 字节迭代：使用 BytesIterator
it := r.NewBytesIterator()
for it.Next() {
    b := it.Current()
}

// 块迭代：使用 ChunksIterator
it := r.Chunks()
for it.Next() {
    chunk := it.Current()
}

// 反向迭代：使用 ReverseIterator
it := r.NewReverseIterator()
for it.Next() {
    ch := it.Current()
}
```

---

## 撤销和重做 (Undo/Redo)

Rope 包提供了完整的撤销/重做功能，基于 Helix 编辑器的设计模式实现。

### 基础 Undo/Redo

```go
package main

import (
    "fmt"
    "github.com/texere-rope/pkg/rope"
)

func main() {
    // 创建文档和历史记录
    doc := rope.New("Hello World")
    history := rope.NewHistory()

    // 第一次编辑：插入 " Beautiful"
    cs := rope.NewChangeSet(doc.Length()).
        Retain(5).
        Insert(" Beautiful")
    txn1 := rope.NewTransaction(cs)

    // 提交到历史并应用
    history.CommitRevision(txn1, doc)
    doc = txn1.Apply(doc)

    fmt.Println(doc.String()) // "Hello Beautiful World"

    // 第二次编辑：删除 " Beautiful"
    cs = rope.NewChangeSet(doc.Length()).
        Retain(5).
        Delete(10)
    txn2 := rope.NewTransaction(cs)

    history.CommitRevision(txn2, doc)
    doc = txn2.Apply(doc)

    fmt.Println(doc.String()) // "Hello World"

    // 撤销第二次编辑
    undoTxn := history.Undo()
    if undoTxn != nil {
        doc = undoTxn.Apply(doc)
        fmt.Println("After undo:", doc.String()) // "Hello Beautiful World"
    }

    // 重做
    redoTxn := history.Redo()
    if redoTxn != nil {
        doc = redoTxn.Apply(doc)
        fmt.Println("After redo:", doc.String()) // "Hello World"
    }
}
```

### 检查 Undo/Redo 可用性

```go
history := rope.NewHistory()

// 初始状态
fmt.Println("CanUndo:", history.CanUndo()) // false
fmt.Println("CanRedo:", history.CanRedo()) // false

// 做一些编辑...
history.CommitRevision(txn, doc)
doc = txn.Apply(doc)

fmt.Println("CanUndo:", history.CanUndo()) // true
fmt.Println("CanRedo:", history.CanRedo()) // false

history.Undo()

fmt.Println("CanUndo:", history.CanUndo()) // false (已回到根)
fmt.Println("CanRedo:", history.CanRedo()) // true
```

### 分支历史

Rope 的历史记录是基于树的，支持非线性撤销：

```go
history := rope.NewHistory()
doc := rope.New("Hello")

// 编辑 1
cs1 := rope.NewChangeSet(doc.Length()).
    Retain(5).
    Insert(" World")
txn1 := rope.NewTransaction(cs1)
history.CommitRevision(txn1, doc)
doc = txn1.Apply(doc)

// 撤销编辑 1
undoTxn := history.Undo()
doc = undoTxn.Apply(doc)

// 做不同的编辑（创建新分支）
cs2 := rope.NewChangeSet(doc.Length()).
    Retain(5).
    Insert(" Gophers")
txn2 := rope.NewTransaction(cs2)
history.CommitRevision(txn2, doc)
doc = txn2.Apply(doc)

fmt.Println(doc.String()) // "Hello Gophers"

// 注意：此时无法直接 redo 到 "Hello World"
// 因为我们在不同的历史分支上
fmt.Println("CanRedo:", history.CanRedo()) // false
```

### 历史记录统计

```go
history := rope.NewHistory()

// ... 做一些编辑 ...

stats := history.Stats()
fmt.Printf("总修订数: %d\n", stats.TotalRevisions)
fmt.Printf("当前索引: %d\n", stats.CurrentIndex)
fmt.Printf("最大大小: %d\n", stats.MaxSize)
fmt.Printf("可撤销: %v\n", stats.CanUndo)
fmt.Printf("可重做: %v\n", stats.CanRedo)
```

### 时间导航

```go
history := rope.NewHistory()
doc := rope.New("Hello")

// 做多个编辑...
for i := 0; i < 5; i++ {
    cs := rope.NewChangeSet(doc.Length()).
        Retain(doc.Length()).
        Insert(string(rune('a' + i)))
    txn := rope.NewTransaction(cs)
    history.CommitRevision(txn, doc)
    doc = txn.Apply(doc)
}

// 撤销 3 步（注意：目前只支持单步）
earlierTxn := history.Undo()
doc = earlierTxn.Apply(doc)
// 继续撤销...
earlierTxn = history.Undo()
doc = earlierTxn.Apply(doc)
earlierTxn = history.Undo()
doc = earlierTxn.Apply(doc)

fmt.Println(doc.String()) // 回到 3 步之前的状态
```

### 历史记录限制

```go
history := rope.NewHistory()
history.SetMaxSize(100) // 最多保留 100 个修订

// 当达到限制时，最旧的修订会被自动移除
```

### 清除历史

```go
history.Clear()
// 清除所有历史记录
```

### 设计原理

Rope 的 undo/redo 基于 Helix 编辑器的设计：

1. **Transaction（事务）**: 代表一个原子编辑操作
2. **ChangeSet（变更集）**: 可组合、可逆的操作序列
3. **Inversion（反转）**: 预计算的撤销操作
4. **Tree History（树形历史）**: 支持非线性撤销分支

关键特性：
- ✅ **不可变性**: 所有操作返回新的 Rope
- ✅ **线程安全**: 历史记录使用锁保护
- ✅ **高效**: 利用 Rope 的持久化特性
- ✅ **灵活**: 支持复杂的编辑模式

---

## 完整示例

### 示例 1: 文本编辑器基础操作

```go
package main

import (
    "fmt"
    "github.com/yourusername/texere-rope/pkg/rope"
)

func main() {
    // 创建文档
    doc := rope.New("Hello World\n")

    // 在光标位置插入文本
    cursorPos := 6  // "Hello|" 后面
    doc = doc.Insert(cursorPos, "Beautiful ")
    fmt.Println(doc.String())
    // 输出: Hello Beautiful World

    // 删除选中文本
    start, end := 6, 16  // "Beautiful"
    doc = doc.Delete(start, end)
    fmt.Println(doc.String())
    // 输出: Hello  World

    // 替换文本
    doc = doc.Replace(6, 7, "Go")
    fmt.Println(doc.String())
    // 输出: Hello Go World
}
```

### 示例 2: 逐字符处理

```go
package main

import (
    "fmt"
    "github.com/yourusername/texere-rope/pkg/rope"
)

func main() {
    text := rope.New("Hello World")

    // 统计字符频率
    freq := make(map[rune]int)
    it := text.NewIterator()
    for it.Next() {
        ch := it.Current()
        freq[ch]++
    }

    fmt.Printf("字符频率: %v\n", freq)
    // 输出: 字符频率: map[72:1 87:1 100:1 101:1 108:3 111:2 32:1]
}
```

### 示例 3: 行编辑

```go
package main

import (
    "fmt"
    "github.com/yourusername/texere-rope/pkg/rope"
)

func main() {
    doc := rope.New("Line1\nLine2\nLine3\n")

    // 在第 2 行前插入注释
    lineNum := 1
    doc = doc.InsertLine(lineNum, "// TODO: fix this\n")
    fmt.Println(doc.String())
    // 输出:
    // Line1
    // // TODO: fix this
    // Line2
    // Line3

    // 删除最后一行
    doc = doc.DeleteLine(doc.LineCount() - 1)
    fmt.Println(doc.String())
    // 输出:
    // Line1
    // // TODO: fix this
    // Line2

    // 添加缩进
    doc = doc.IndentLines("  ")
    fmt.Println(doc.String())
    // 输出:
    //   Line1
    //   // TODO: fix this
    //   Line2
}
```

### 示例 4: 使用 Builder 构建大文本

```go
package main

import (
    "fmt"
    "github.com/yourusername/texere-rope/pkg/rope"
)

func main() {
    // 构建大型文档
    builder := rope.NewBuilder()

    // 添加标题
    builder.AppendLine("# Document Title")
    builder.AppendLine("")

    // 添加多行内容
    for i := 1; i <= 100; i++ {
        builder.AppendLine(fmt.Sprintf("Line %d: Some content here", i))
    }

    // 构建最终文档
    doc := builder.Build()

    // 统计信息
    fmt.Printf("Lines: %d\n", doc.LineCount())
    fmt.Printf("Characters: %d\n", doc.Length())
    fmt.Printf("Bytes: %d\n", doc.Size())
}
```

### 示例 5: 撤销/重做历史

```go
package main

import (
    "fmt"
    "github.com/yourusername/texere-rope/pkg/rope"
)

// History 简单的撤销/重做历史
type History struct {
    past   []*rope.Rope
    future []*rope.Rope
}

func NewHistory(initial *rope.Rpe) *History {
    return &History{
        past:   []*rope.Rope{initial},
        future: nil,
    }
}

func (h *History) Current() *rope.Rope {
    return h.past[len(h.past)-1]
}

func (h *History) Apply(newRope *rope.Rope) {
    h.past = append(h.past, newRope)
    h.future = nil  // 清空未来历史
}

func (h *History) Undo() *rope.Rope {
    if len(h.past) <= 1 {
        return nil
    }

    current := h.past[len(h.past)-1]
    h.past = h.past[:len(h.past)-1]
    h.future = append([]*rope.Rope{current}, h.future...)

    return h.Current()
}

func (h *History) Redo() *rope.Rope {
    if len(h.future) == 0 {
        return nil
    }

    next := h.future[0]
    h.future = h.future[1:]
    h.past = append(h.past, next)

    return next
}

func main() {
    doc := rope.New("Hello")
    history := NewHistory(doc)

    // 执行操作
    doc = doc.Insert(5, " World")
    history.Apply(doc)

    doc = doc.Append("!")
    history.Apply(doc)

    fmt.Println(doc.String())  // "Hello World!"

    // 撤销
    doc = history.Undo()
    fmt.Println(doc.String())  // "Hello World"

    doc = history.Undo()
    fmt.Println(doc.String())  // "Hello"

    // 重做
    doc = history.Redo()
    fmt.Println(doc.String())  // "Hello World"
}
```

---

## 性能基准

基于当前实现的性能数据（Go 1.23, AMD Ryzen 7 5800X）：

| 操作 | 性能 | 分配 |
|------|------|------|
| Length | 0.9 ns | 0 B, 0 allocs |
| Slice (middle) | 58,776 ns | 0 B, 0 allocs |
| Iterator traversal | 0.13 ms | 384 B, 1 alloc |
| Insert (middle) | 4,766 ns | 3 allocs |
| Delete (small) | 1,488 ns | 1 alloc |
| CharAt | 19,092 ns | 1 alloc |

**注意**:
- 性能数据随硬件和 Go 版本变化
- 微基准测试结果仅供参考
- 实际应用性能取决于具体使用模式

---

## 最佳实践总结

1. **使用迭代器**进行遍历，避免重复 CharAt 调用
2. **批量构建**使用 RopeBuilder，避免重复创建临时对象
3. **缓存 String()**结果，避免重复转换
4. **优先使用 Slice**而非逐字符访问
5. **利用不可变性**简化并发代码和撤销/重做逻辑
6. **选择合适的迭代器**类型（Iterator/BytesIterator/ChunksIterator）
7. **小文本使用原生 string**，Rope 适合大型文本
8. **避免频繁的 Rope↔string 转换**

---

## 更多资源

- [GoDoc 文档](https://pkg.go.dev/github.com/yourusername/texere-rope/pkg/rope)
- [源代码](https://github.com/yourusername/texere-rope)
- [性能分析](PERFORMANCE_ANALYSIS.md)
- [基准测试](pkg/rope/bench_test.go)

---

**License**: MIT
**作者**: [Your Name]
**最后更新**: 2025-01-30
