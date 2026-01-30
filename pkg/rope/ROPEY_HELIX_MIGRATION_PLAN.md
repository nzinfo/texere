# Ropey & Helix 功能差异分析与迁移计划

> **日期**: 2026-01-31
> **目的**: 全面分析 ropey 和 helix 的功能，识别 texere-rope 缺失的部分，制定迁移计划

---

## 📊 执行摘要

### 当前状态
- ✅ **Helix 对齐**: 100% 完成（HELIX_ALIGNMENT.md）
- ⚠️ **Ropey 对齐**: 约 23.5% 测试覆盖率（ROPEY_MISSING_FEATURES.md）
- 📈 **已实现功能**: Grapheme、Chunk_at、Position Mapping、Time-based Undo

### 关键发现
1. **Helix 的所有核心功能已完全实现** ✅
2. **Ropey 基础 API 已实现**，但缺少一些高级功能
3. **需要迁移的主要功能**：
   - UTF-16 支持（JS/Windows 互操作）
   - 单字符便捷方法
   - Rope 拼接优化
   - Hash 支持
   - CRLF 智能处理

---

## 第一部分：Ropey API 详细分析

### 1.1 核心构造函数

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `new()` | `New("")` | ✅ | 等价 |
| `from_str(text: &str)` | `New(text)` | ✅ | 等价 |
| `from_reader(reader)` | ❌ | ⚠️ P1 | 流式读取 |
| `write_to(writer)` | ❌ | ⚠️ P2 | 流式写入 |

### 1.2 查询方法

#### 1.2.1 长度查询

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `len_bytes()` | `Size()` / `LenBytes()` | ✅ | 等价 |
| `len_chars()` | `Length()` | ✅ | 等价 |
| `len_lines()` | `LenLines()` | ✅ | 等价 |
| `len_utf16_cu()` | ❌ | ⚠️ **P0** | UTF-16 code units |
| `capacity()` | ❌ | ⏭️ | Go 中不需要 |
| `shrink_to_fit()` | ❌ | ⏭️ | Go GC 自动处理 |

**优先级说明**：
- **P0**: 立即实现（核心功能）
- **P1**: 尽快实现（性能优化）
- **P2**: 可选实现（增强功能）
- **⏭️**: 不需要实现（语言差异）

#### 1.2.2 索引转换

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `byte_to_char(byte_idx)` | ❌ | ⚠️ P1 | 字节→字符 |
| `byte_to_line(byte_idx)` | ❌ | ⚠️ P2 | 字节→行 |
| `char_to_byte(char_idx)` | `ByteIndex()` / `IndexToByte()` | ✅ | 等价 |
| `char_to_line(char_idx)` | `LineAtChar()` | ✅ | 已实现 |
| `char_to_utf16_cu(char_idx)` | ❌ | ⚠️ P0 | 字符→UTF16 |
| `utf16_cu_to_char(utf16_idx)` | ❌ | ⚠️ P0 | UTF16→字符 |
| `line_to_byte(line_idx)` | ❌ | ⚠️ P2 | 行→字节 |
| `line_to_char(line_idx)` | `LineToChar()` | ✅ | 等价 |

### 1.3 编辑操作

#### 1.3.1 插入操作

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `insert(char_idx, text: &str)` | `Insert(pos, text)` | ✅ | 等价 |
| `insert_char(char_idx, ch: char)` | ❌ | ⚠️ **P0** | 单字符插入 |
| `try_insert(...)` | ❌ | ⏭️ | Go 使用 panic/recover |
| `try_insert_char(...)` | ❌ | ⏭️ | Go 使用 panic/recover |

**实现建议**：
```go
// InsertChar 在指定位置插入单个字符
func (r *Rope) InsertChar(pos int, ch rune) *Rope {
    return r.Insert(pos, string(ch))
}
```

#### 1.3.2 删除操作

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `remove(char_range)` | `Delete(start, end)` | ✅ | 等价 |
| `try_remove(...)` | ❌ | ⏭️ | Go 使用 panic/recover |

#### 1.3.3 Rope 拼接

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `append(other: Rope)` | `Append(text)` | ⚠️ | 仅支持字符串 |
| `split_off(char_idx)` | ❌ | ⚠️ P2 | 分割 Rope |
| **`append_rope(other: Rope)`** | ❌ | ⚠️ **P1** | **需要实现** |
| **`prepend(text: &str)`** | `Insert(0, text)` | ✅ | 可用 Insert 替代 |
| **`prepend_rope(other: Rope)`** | ❌ | ⚠️ **P1** | **需要实现** |

**实现建议**：
```go
// AppendRope 高效拼接两个 Rope（避免字符串转换）
func (r *Rope) AppendRope(other *Rope) *Rope {
    if other.Length() == 0 {
        return r
    }
    if r.Length() == 0 {
        return other.Clone()
    }
    // 直接合并节点树，避免转换为字符串
    return r.mergeTree(other)
}

// Prepend 在开头添加内容
func (r *Rope) Prepend(text string) *Rope {
    return r.Insert(0, text)
}

// PrependRope 在开头添加另一个 Rope
func (r *Rope) PrependRope(other *Rope) *Rope {
    if other.Length() == 0 {
        return r
    }
    if r.Length() == 0 {
        return other.Clone()
    }
    return other.mergeTree(r)
}
```

### 1.4 内容访问

#### 1.4.1 单个元素访问

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `byte(byte_idx)` | ❌ | ⏭️ | 使用 CharAt() |
| `char(char_idx)` | `CharAt(pos)` | ✅ | 等价 |
| `get_byte(byte_idx)` | ❌ | ⏭️ | 返回 Option |
| `get_char(char_idx)` | ❌ | ⏭️ | 返回 Option |

#### 1.4.2 行访问

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `line(line_idx)` | `LineWithEnding(lineNum)` | ✅ | 等价 |
| `get_line(line_idx)` | ❌ | ⏭️ | 返回 Option |

#### 1.4.3 Chunk 访问（高级优化）

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `chunk_at_byte(byte_idx)` | `ChunkAtByte(pos)` | ✅ | 已实现 |
| `chunk_at_char(char_idx)` | `ChunkAtChar(pos)` | ✅ | 已实现 |
| `chunk_at_line_break(idx)` | ❌ | ⭐ P3 | 高级优化 |
| `get_chunk_at_byte(byte_idx)` | ❌ | ⏭️ | 返回 Option |

**说明**：Chunk 访问用于高级性能优化，普通用户不需要。当前已实现基础版本。

### 1.5 切片操作

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `slice(char_range)` | `Slice(start, end)` | ✅ | 等价 |
| `byte_slice(byte_range)` | ❌ | ⚠️ P2 | 字节级切片 |
| `get_slice(char_range)` | ❌ | ⏭️ | 返回 Option |
| `get_byte_slice(byte_range)` | ❌ | ⏭️ | 返回 Option |

### 1.6 迭代器

#### 1.6.1 字节迭代器

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `bytes()` | ❌ | ⚠️ **P1** | 字节迭代器 |
| `bytes_at(byte_idx)` | ❌ | ⚠️ **P1** | 从位置开始 |
| `get_bytes_at(byte_idx)` | ❌ | ⏭️ | 返回 Option |

**实现建议**：
```go
// BytesIterator 字节级迭代器
type BytesIterator struct {
    rope  *Rope
    pos   int  // 字节位置
}

func (r *Rope) Bytes() *BytesIterator {
    return &BytesIterator{rope: r, pos: 0}
}

func (r *Rope) BytesAt(byteIdx int) *BytesIterator {
    return &BytesIterator{rope: r, pos: byteIdx}
}

func (it *BytesIterator) Next() bool {
    it.pos++
    return it.pos < it.rope.Size()
}

func (it *BytesIterator) Current() byte {
    return it.rope.byteAt(it.pos)
}
```

#### 1.6.2 字符迭代器

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `chars()` | `IterChars()` / `Graphemes()` | ✅ | 已实现 |
| `chars_at(char_idx)` | `IteratorAt(pos)` | ✅ | 已实现 |
| `get_chars_at(char_idx)` | ❌ | ⏭️ | 返回 Option |

#### 1.6.3 行迭代器

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `lines()` | `IterLines()` | ✅ | 已实现 |
| `lines_at(line_idx)` | ❌ | ⚠️ P2 | 从指定行开始 |

#### 1.6.4 Chunk 迭代器

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `chunks()` | ❌ | ⭐ P3 | 高级优化 |
| `chunks_at_byte(byte_idx)` | ❌ | ⭐ P3 | 高级优化 |
| `chunks_at_char(char_idx)` | ❌ | ⭐ P3 | 高级优化 |

#### 1.6.5 反向迭代器

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `chars_at_rev(char_idx)` | ❌ | ⚠️ P2 | 反向字符迭代 |

**实现建议**：
```go
// ReverseIterator 反向字符迭代器
type ReverseIterator struct {
    rope *Rope
    pos  int  // 当前字符位置
}

func (r *Rope) IterCharsReverse(pos int) *ReverseIterator {
    return &ReverseIterator{rope: r, pos: pos}
}

func (it *ReverseIterator) Next() bool {
    it.pos--
    return it.pos >= 0
}

func (it *ReverseIterator) Current() rune {
    return it.rope.CharAt(it.pos)
}
```

### 1.7 完整性检查

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| `is_instance(other: &Rope)` | ❌ | ⏭️ | Rust 特有 |
| `assert_integrity()` | ❌ | ⭐ | 测试工具 |
| `assert_invariants()` | ❌ | ⭐ | 测试工具 |

### 1.8 Try 方法（错误处理）

| Ropey API | texere-rope | 状态 | 备注 |
|-----------|-------------|------|------|
| 所有 `try_*` 方法 | ❌ | ⏭️ | Go 使用 panic/recover |

**说明**：Ropey 使用 Result<T, Error>，Go 使用 panic/recover 机制，这是语言差异。

### 1.9 其他实用功能

| 功能 | Ropey | texere-rope | 状态 | 优先级 |
|------|-------|-------------|------|--------|
| **Hash 支持** | `impl Hash` | ❌ | ⚠️ **P0** | HashMap 键 |
| **Common Prefix** | `common_prefix()` | ❌ | ⚠️ P1 | 文本比较 |
| **Common Suffix** | `common_suffix()` | ❌ | ⚠️ P1 | 文本比较 |
| **CRLF 处理** | `find_good_split()` | ⚠️ 部分 | ⚠️ **P0** | Windows 兼容 |
| **Copy-on-Write** | Cow<RopeNode> | ❌ | ⏭️ | Rust 特有 |

**Hash 实现建议**：
```go
func (r *Rope) HashCode() uint32 {
    h := fnv.New32a()
    it := r.IterChars()
    for it.Next() {
        ch := it.Current()
        h.WriteRune(ch)
    }
    return h.Sum32()
}
```

**Common Prefix/Suffix 实现**：
```go
// CommonPrefix 返回两个 Rope 的最长公共前缀长度
func (r *Rope) CommonPrefix(other *Rope) int {
    it1 := r.IterChars()
    it2 := other.IterChars()
    count := 0

    for it1.Next() && it2.Next() {
        if it1.Current() != it2.Current() {
            break
        }
        count++
    }
    return count
}

// CommonSuffix 返回两个 Rope 的最长公共后缀长度
func (r *Rope) CommonSuffix(other *Rope) int {
    // 使用反向迭代器
    it1 := r.IterCharsReverse(r.Length() - 1)
    it2 := other.IterCharsReverse(other.Length() - 1)
    count := 0

    for it1.Next() && it2.Next() {
        if it1.Current() != it2.Current() {
            break
        }
        count++
    }
    return count
}
```

---

## 第二部分：Helix Transaction 系统分析

### 2.1 Helix 核心概念

根据 HELIX_ALIGNMENT.md，Helix 的 transaction 系统包含：

#### 2.1.1 Operation 类型

```rust
pub enum Operation {
    Retain(usize),        // 保留 n 个字符
    Delete(usize),        // 删除 n 个字符
    Insert(String),       // 插入文本
}
```

**texere-rope 等价实现**：✅ 完全对齐
```go
const (
    OpRetain OperationType = iota
    OpDelete
    OpInsert
)
```

#### 2.1.2 Assoc 类型（光标关联）

```rust
pub enum Assoc {
    Before,           // 光标在编辑之前
    After,            // 光标在编辑之后
    BeforeWord,       // 光标在之前单词开头
    AfterWord,        // 光标在之后单词开头
    BeforeSticky,     // 粘性定位
    AfterSticky,      // 粘性定位
}
```

**texere-rope 等价实现**：✅ 完全对齐
```go
type Assoc int

const (
    AssocBefore Assoc = iota
    AssocAfter
    AssocBeforeWord
    AssocAfterWord
    AssocBeforeSticky
    AssocAfterSticky
)
```

#### 2.1.3 ChangeSet 组合

**Helix 功能**：
```rust
pub fn compose(&self, other: &ChangeSet) -> ChangeSet
pub fn map_position(&self, pos: usize, assoc: Assoc) -> usize
pub fn split(&self, pos: usize) -> (ChangeSet, ChangeSet)
pub fn merge(&self, other: &ChangeSet) -> ChangeSet
pub fn transform(&self, other: &ChangeSet) -> ChangeSet
```

**texere-rope 状态**：✅ 完全实现
- `Compose(cs1, cs2)` - ✅
- `MapPosition(pos, assoc)` - ✅
- `MapPositions(positions, associations)` - ✅ 批量映射
- `Split(pos)` - ✅
- `Merge(other)` - ✅
- `Transform(other)` - ✅

#### 2.1.4 时间导航

**Helix 功能**：
```rust
pub fn earlier_by_time(&self, duration: Duration) -> Transaction
pub fn later_by_time(&self, duration: Duration) -> Transaction
```

**texere-rope 状态**：✅ 完全实现（使用二分查找 O(log N)）
- `EarlierByTime(duration)` - ✅
- `LaterByTime(duration)` - ✅
- `EarlierByDuration(duration)` - ✅ 返回 History
- `LaterByDuration(duration)` - ✅ 返回 History

#### 2.1.5 词边界检测

**Helix 功能**：
```rust
pub fn prev_word_start(&self, pos: usize) -> usize
pub fn next_word_start(&self, pos: usize) -> usize
pub fn word_at(&self, pos: usize) -> (String, usize, usize)
```

**texere-rope 状态**：✅ 完全实现，额外提供
- `PrevWordStart(pos)` - ✅
- `NextWordStart(pos)` - ✅
- `WordAt(pos)` - ✅
- `SelectWord(pos)` - ✅
- `BigWordStart/End` - ✅ 额外
- `ParagraphStart/End` - ✅ 额外
- `LineStart/End` - ✅ 额外

### 2.2 Helix vs texere-rope 对比总结

| 功能类别 | Helix | texere-rope | 状态 |
|---------|-------|-------------|------|
| **ChangeSet 基础** | ✅ | ✅ | 完全对齐 |
| **Operation 类型** | 3种 | 3种 | 完全对齐 |
| **Assoc 模式** | 6种 | 6种 | 完全对齐 |
| **Compose** | ✅ | ✅ | 完全对齐 |
| **Position Mapping** | ✅ | ✅ + 批量优化 | **超越** |
| **Split/Merge** | ✅ | ✅ | 完全对齐 |
| **Transform** | ✅ | ✅ | 完全对齐 |
| **Time Navigation** | ✅ | ✅ O(log N) | 完全对齐 |
| **Word Boundaries** | ✅ | ✅ + 额外 | **超越** |
| **Savepoint** | ✅ | ✅ | 完全对齐 |
| **Memory Pooling** | ❌ | ✅ | **超越** |
| **Lazy Evaluation** | ❌ | ✅ | **超越** |

**结论**：texere-rope 的 Helix 对齐度为 **100%**，并在某些方面超越原实现。

---

## 第三部分：缺失功能详细分析与实现

### 3.1 P0 优先级（立即实现）

#### 3.1.1 UTF-16 支持

**需求来源**：
- JavaScript 互操作（String 使用 UTF-16）
- Windows 平台（内部使用 UTF-16）

**缺失的 API**：
```rust
// Ropey API
pub fn len_utf16_cu(&self) -> usize
pub fn char_to_utf16_cu(&self, char_idx: usize) -> usize
pub fn utf16_cu_to_char(&self, utf16_cu_idx: usize) -> usize
```

**实现计划**：
```go
// UTF-16 支持实现
package rope

// LenUTF16 返回 UTF-16 code units 数量
func (r *Rope) LenUTF16() int {
    count := 0
    it := r.IterGraphemes()
    for it.Next() {
        cluster := it.Current()
        for _, r := range cluster {
            if r <= 0xFFFF {
                count += 1  // BMP 字符
            } else {
                count += 2  // 代理对
            }
        }
    }
    return count
}

// CharToUTF16 将字符索引转换为 UTF-16 索引
func (r *Rope) CharToUTF16(charIdx int) int {
    utf16Idx := 0
    charCount := 0
    it := r.IterGraphemes()

    for it.Next() && charCount < charIdx {
        cluster := it.Current()
        for _, r := range cluster {
            if r <= 0xFFFF {
                utf16Idx += 1
            } else {
                utf16Idx += 2
            }
        }
        charCount++
    }

    return utf16Idx
}

// UTF16ToChar 将 UTF-16 索引转换为字符索引
func (r *Rope) UTF16ToChar(utf16Idx int) int {
    currentUtf16 := 0
    charCount := 0
    it := r.IterGraphemes()

    for it.Next() {
        cluster := it.Current()
        clusterUtf16 := 0
        for _, r := range cluster {
            if r <= 0xFFFF {
                clusterUtf16 += 1
            } else {
                clusterUtf16 += 2
            }
        }

        if currentUtf16 + clusterUtf16 > utf16Idx {
            return charCount
        }

        currentUtf16 += clusterUtf16
        charCount++
    }

    return charCount
}
```

**测试计划**：
```go
func TestUTF16Support(t *testing.T) {
    // BMP 字符
    r1 := New("Hello")
    assert.Equal(t, 5, r1.LenUTF16())

    // 包含代理对的字符
    r2 := New("Hello 世界𠮷")  // 𠮷 需要代理对
    assert.Equal(t, 5 + 2*3 + 2, r2.LenUTF16())

    // 索引转换
    r3 := New("AB𠮷CD")  // 𠮷 在位置 2
    assert.Equal(t, 0, r3.CharToUTF16(0))    // A
    assert.Equal(t, 1, r3.CharToUTF16(1))    // B
    assert.Equal(t, 2, r3.CharToUTF16(2))    // 𠮷 开始
    assert.Equal(t, 4, r3.CharToUTF16(3))    // C
    assert.Equal(t, 5, r3.CharToUTF16(4))    // D
}
```

#### 3.1.2 单字符操作

**实现**：
```go
// InsertChar 在指定位置插入单个字符
func (r *Rope) InsertChar(pos int, ch rune) *Rope {
    return r.Insert(pos, string(ch))
}

// RemoveChar 删除指定位置的单个字符
func (r *Rope) RemoveChar(pos int) *Rope {
    return r.Delete(pos, pos+1)
}
```

**测试**：
```go
func TestSingleCharOperations(t *testing.T) {
    r := New("Hello World")

    // InsertChar
    r = r.InsertChar(5, ',')
    assert.Equal(t, "Hello, World", r.String())

    // RemoveChar
    r = r.RemoveChar(5)
    assert.Equal(t, "Hello World", r.String())
}
```

#### 3.1.3 Hash 支持

**实现**：
```go
import "hash/fnv"

// HashCode 返回 Rope 的哈希值
func (r *Rope) HashCode() uint32 {
    h := fnv.New32a()
    it := r.IterGraphemes()
    for it.Next() {
        h.Write([]byte(it.Current()))
    }
    return h.Sum32()
}
```

**使用场景**：
```go
// 作为 map 键
type CachedDocument struct {
    content *Rope
    hash    uint32
}

func (d *CachedDocument) UpdateContent(r *Rope) {
    d.content = r
    d.hash = r.HashCode()
}

// 文档去重
func DeduplicateDocs(docs []*Rope) []*Rope {
    seen := make(map[uint32]bool)
    result := make([]*Rope, 0)

    for _, doc := range docs {
        hash := doc.HashCode()
        if !seen[hash] {
            seen[hash] = true
            result = append(result, doc)
        }
    }
    return result
}
```

#### 3.1.4 CRLF 智能处理

**需求**：避免在 CRLF 中间分割（Windows 换行符）

**实现**：
```go
// findGoodSplit 查找合适的分割点，避免拆分 CRLF
func findGoodSplit(pos int, text []byte, minSplit bool) int {
    // 检查是否在 CRLF 中间
    if pos > 0 && pos < len(text) {
        if text[pos-1] == '\r' && text[pos] == '\n' {
            // 调整位置避免分割 CRLF
            if minSplit {
                return pos - 1  // 向前调整
            }
            return pos + 1  // 向后调整
        }
    }
    return pos
}

// 在 Rope 创建时应用
func (b *RopeBuilder) AppendWithCRLF(text string) *RopeBuilder {
    bytes := []byte(text)
    splitPoints := calculateSplitPoints(bytes)

    for _, pt := range splitPoints {
        adjusted := findGoodSplit(pt, bytes, false)
        // 使用 adjusted 作为分割点
    }
    return b
}
```

**测试**：
```go
func TestCRLFHandling(t *testing.T) {
    text := "Line1\r\nLine2\r\nLine3"

    // 测试不会在 CRLF 中间分割
    r := New(text)
    chunks := r.Chunks()

    for _, chunk := range chunks {
        // 验证没有 "\r" 单独在末尾
        if strings.HasSuffix(chunk, "\r") {
            t.Fatal("Chunk ends with bare \\r")
        }
        // 验证没有 "\n" 单独在开头
        if strings.HasPrefix(chunk, "\n") {
            t.Fatal("Chunk starts with bare \\n")
        }
    }
}
```

### 3.2 P1 优先级（尽快实现）

#### 3.2.1 Rope 拼接优化

**实现**：
```go
// AppendRope 高效拼接两个 Rope
func (r *Rope) AppendRope(other *Rope) *Rope {
    if other.Length() == 0 {
        return r
    }
    if r.Length() == 0 {
        return other.Clone()
    }

    // 创建新的内部节点，直接合并树
    newNode := &InternalNode{
        left:  r.root,
        right: other.root,
        length: r.length,
        size:   r.size,
    }

    return &Rope{
        root:   newNode,
        length: r.length + other.length,
        size:   r.size + other.size,
    }
}

// PrependRope 在开头添加另一个 Rope
func (r *Rope) PrependRope(other *Rope) *Rope {
    return other.AppendRope(r)
}
```

**性能对比**：
```go
// 当前方式（字符串转换）
r1 := New("Hello")
r2 := New(" World")
r3 := r1.Append(r2.String())  // 需要转换 r2 为字符串

// 优化方式（直接拼接）
r1 := New("Hello")
r2 := New(" World")
r3 := r1.AppendRope(r2)  // 直接合并节点树
```

#### 3.2.2 字节级迭代器

**实现**：
```go
// BytesIterator 字节级迭代器
type BytesIterator struct {
    rope    *Rope
    bytePos int
    chunk   []byte
    chunkIdx int
}

func (r *Rope) Bytes() *BytesIterator {
    return &BytesIterator{
        rope:    r,
        bytePos: 0,
    }
}

func (r *Rope) BytesAt(byteIdx int) *BytesIterator {
    return &BytesIterator{
        rope:    r,
        bytePos: byteIdx,
    }
}

func (it *BytesIterator) Next() bool {
    it.bytePos++
    // 加载下一个 chunk
    return it.bytePos < it.rope.Size()
}

func (it *BytesIterator) Current() byte {
    // 从当前 chunk 返回字节
    return it.rope.byteAt(it.bytePos)
}

func (it *BytesIterator) Seek(byteIdx int) {
    it.bytePos = byteIdx
}
```

**测试**：
```go
func TestBytesIterator(t *testing.T) {
    r := New("Hello 世界")

    it := r.Bytes()
    bytes := make([]byte, 0)

    for it.Next() {
        bytes = append(bytes, it.Current())
    }

    expected := []byte("Hello 世界")
    assert.Equal(t, expected, bytes)
}
```

#### 3.2.3 索引转换方法

**实现**：
```go
// ByteToChar 将字节索引转换为字符索引
func (r *Rope) ByteToChar(byteIdx int) int {
    if byteIdx < 0 || byteIdx > r.Size() {
        panic("byte index out of bounds")
    }

    charCount := 0
    byteCount := 0
    it := r.IterGraphemes()

    for it.Next() {
        cluster := it.Current()
        clusterBytes := len([]byte(cluster))

        if byteCount + clusterBytes > byteIdx {
            return charCount
        }

        byteCount += clusterBytes
        charCount++
    }

    return charCount
}

// ByteToLine 将字节索引转换为行号
func (r *Rope) ByteToLine(byteIdx int) int {
    charIdx := r.ByteToChar(byteIdx)
    return r.LineAtChar(charIdx)
}

// LineToByte 将行号转换为字节索引
func (r *Rope) LineToByte(lineIdx int) int {
    charIdx := r.LineToChar(lineIdx)
    return r.IndexToByte(charIdx)
}
```

#### 3.2.4 CommonPrefix/CommonSuffix

**实现**（已在 1.9 节展示）

### 3.3 P2 优先级（可选实现）

#### 3.2.5 反向迭代器

**实现**（已在 1.6.5 节展示）

#### 3.2.6 SplitOff

**实现**：
```go
// SplitOff 将 Rope 从指定位置分割成两个
func (r *Rope) SplitOff(pos int) (*Rope, *Rope) {
    if pos <= 0 {
        return Empty(), r.Clone()
    }
    if pos >= r.Length() {
        return r.Clone(), Empty()
    }

    left := r.Slice(0, pos)
    right := r.Slice(pos, r.Length())

    return left.AsRope(), right.AsRope()
}
```

#### 3.2.7 流式 I/O

**实现**：
```go
import (
    "io"
    "bufio"
)

// FromReader 从 io.Reader 读取内容创建 Rope
func FromReader(reader io.Reader) (*Rope, error) {
    b := NewBuilder()
    bufReader := bufio.NewReader(reader)
    buf := make([]byte, 4096)

    for {
        n, err := bufReader.Read(buf)
        if n > 0 {
            b.Append(string(buf[:n]))
        }
        if err != nil {
            if err == io.EOF {
                return b.Build(), nil
            }
            return nil, err
        }
    }
}

// WriteTo 将 Rope 内容写入 io.Writer
func (r *Rope) WriteTo(writer io.Writer) (int, error) {
    it := r.IterChunks()
    total := 0

    for it.Next() {
        chunk := it.Current()
        n, err := writer.Write([]byte(chunk))
        total += n
        if err != nil {
            return total, err
        }
    }

    return total, nil
}
```

---

## 第四部分：迁移实施计划

### 4.1 阶段划分

#### 阶段 1：P0 核心功能（1-2 周）

**目标**：实现最常用和最重要的功能

- [ ] **UTF-16 支持**（2-3 天）
  - [ ] `LenUTF16()`
  - [ ] `CharToUTF16()`
  - [ ] `UTF16ToChar()`
  - [ ] 测试覆盖

- [ ] **单字符操作**（1 天）
  - [ ] `InsertChar()`
  - [ ] `RemoveChar()`
  - [ ] 测试覆盖

- [ ] **Hash 支持**（1 天）
  - [ ] `HashCode()` 方法
  - [ ] 文档和使用示例

- [ ] **CRLF 智能处理**（2-3 天）
  - [ ] `findGoodSplit()` 函数
  - [ ] Builder 集成
  - [ ] 测试覆盖

**预期成果**：
- 新增 4 个核心功能
- 测试覆盖率提升至 ~30%
- 完全支持 JavaScript/Windows 互操作

#### 阶段 2：P1 性能优化（2-3 周）

**目标**：实现性能优化相关功能

- [ ] **Rope 拼接优化**（3-4 天）
  - [ ] `AppendRope()`
  - [ ] `PrependRope()`
  - [ ] `Prepend()`
  - [ ] 性能基准测试

- [ ] **字节级迭代器**（2-3 天）
  - [ ] `Bytes()` 迭代器
  - [ ] `BytesAt()` 迭代器
  - [ ] 测试覆盖

- [ ] **索引转换**（2-3 天）
  - [ ] `ByteToChar()`
  - [ ] `ByteToLine()`
  - [ ] `LineToByte()`
  - [ ] 测试覆盖

- [ ] **CommonPrefix/Suffix**（1-2 天）
  - [ ] `CommonPrefix()`
  - [ ] `CommonSuffix()`
  - [ ] 测试覆盖

**预期成果**：
- 新增 10+ 个 API 方法
- 性能优化（避免字符串转换）
- 测试覆盖率提升至 ~40%

#### 阶段 3：P2 增强功能（按需实现）

- [ ] **反向迭代器**（2-3 天）
- [ ] **SplitOff**（1 天）
- [ ] **流式 I/O**（3-4 天）
- [ ] **行迭代器增强**（1-2 天）

**预期成果**：
- 功能完整性达到 ~90%
- 测试覆盖率提升至 ~45%

#### 阶段 4：P3 高级优化（长期）

- [ ] **Chunk 迭代器**（高级优化）
- [ ] **Copy-on-Write**（需要重构）
- [ ] **完整性检查工具**（测试辅助）

### 4.2 实施顺序建议

**推荐路径**：
```
阶段 1 (P0)
  ↓
阶段 2 (P1)
  ↓
根据实际需求选择阶段 3 功能
  ↓
长期优化 (P3)
```

**并行工作**：
- UTF-16 支持和 Hash 支持可以并行（独立模块）
- Rope 拼接和字节迭代器可以并行
- 测试编写与实现并行

### 4.3 测试策略

#### 4.3.1 单元测试

每个新功能都需要完整的单元测试：
```go
func TestFeatureName(t *testing.T) {
    // 正常情况
    // 边界情况
    // 错误情况
    // 性能基准
}
```

#### 4.3.2 对比测试

使用 ropey 作为参考实现：
```go
func TestComparisonWithRopey(t *testing.T) {
    // 创建相同的测试场景
    // 对比结果
    // 验证一致性
}
```

#### 4.3.3 性能基准

为关键功能添加基准测试：
```go
func BenchmarkAppendRope(b *testing.B) {
    r := New("Hello")
    other := New(" World")
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        r = r.AppendRope(other)
    }
}
```

### 4.4 文档计划

#### 4.4.1 API 文档

为每个新功能添加完整的文档：
```go
// LenUTF16 returns the number of UTF-16 code units in the rope.
//
// This is useful for JavaScript interoperability and Windows platform
// development, where text is measured in UTF-16 code units.
//
// Note that characters outside the Basic Multilingual Plane (BMP)
// (e.g., emoji, some CJK characters) count as 2 code units.
//
// Example:
//   r := rope.New("Hello 𠮷")
//   fmt.Println(r.LenUTF16())  // Output: 9 (5 + 1 + 1*2)
func (r *Rope) LenUTF16() int
```

#### 4.4.2 使用指南

创建专门的使用指南：
- `UTF16_SUPPORT.md` - UTF-16 功能使用指南
- `ADVANCED_IO.md` - 流式 I/O 使用指南
- `PERFORMANCE_OPTIMIZATIONS.md` - 性能优化最佳实践

#### 4.4.3 迁移指南

为从 ropey 迁移的用户提供：
```markdown
# From Ropey to texere-rope

## API Mapping

| Ropey (Rust) | texere-rope (Go) | Notes |
|--------------|-----------------|-------|
| `len_utf16_cu()` | `LenUTF16()` | Identical |
| `insert_char()` | `InsertChar()` | Identical |
| `append(other)` | `AppendRope(other)` | More efficient |
```

---

## 第五部分：优先级矩阵

### 5.1 功能优先级评估

| 功能 | 重要性 | 紧急性 | 实现难度 | ROI | 优先级 |
|------|-------|--------|---------|-----|--------|
| **UTF-16 支持** | 高 | 高 | 中 | 高 | **P0** |
| **单字符操作** | 高 | 高 | 低 | 高 | **P0** |
| **Hash 支持** | 中 | 高 | 低 | 高 | **P0** |
| **CRLF 处理** | 高 | 中 | 中 | 中 | **P0** |
| **Rope 拼接** | 中 | 中 | 中 | 中 | **P1** |
| **字节迭代器** | 中 | 中 | 中 | 中 | **P1** |
| **索引转换** | 中 | 中 | 低 | 中 | **P1** |
| **CommonPrefix** | 低 | 低 | 低 | 低 | **P1** |
| **反向迭代器** | 低 | 低 | 中 | 低 | **P2** |
| **SplitOff** | 低 | 低 | 低 | 低 | **P2** |
| **流式 I/O** | 中 | 低 | 中 | 中 | **P2** |
| **Chunk 迭代器** | 低 | 低 | 高 | 低 | **P3** |

**Legend**:
- **重要性**: 对用户的影响程度
- **紧迫性**: 需求的紧急程度
- **实现难度**: 技术复杂度
- **ROI**: 投资回报率（价值/成本）

### 5.2 实施决策矩阵

**立即实施（P0）**：
1. ✅ UTF-16 支持 - JavaScript/Windows 必须
2. ✅ 单字符操作 - 编辑器常用操作
3. ✅ Hash 支持 - Go 生态需要
4. ✅ CRLF 处理 - Windows 兼容性

**尽快实施（P1）**：
1. Rope 拼接 - 性能优化
2. 字节迭代器 - 二进制场景
3. 索引转换 - 完善功能
4. CommonPrefix/Suffix - 实用工具

**按需实施（P2）**：
1. 反向迭代器 - 某些编辑操作
2. SplitOff - 特殊场景
3. 流式 I/O - 大文件处理

**长期优化（P3）**：
1. Chunk 迭代器 - 内部优化
2. Copy-on-Write - 架构级重构
3. 完整性检查 - 测试工具

---

## 第六部分：风险评估与缓解

### 6.1 技术风险

#### 风险 1：UTF-16 性能问题

**描述**：UTF-16 转换可能影响性能

**影响**：中等

**缓解措施**：
- 使用缓存（Memoization）
- 仅在需要时计算
- 提供快速路径（纯 ASCII）

```go
func (r *Rope) LenUTF16() int {
    // 快速路径：纯 ASCII
    if r.isPureASCII() {
        return r.Length()
    }

    // 缓存结果
    if r.utf16Len > 0 {
        return r.utf16Len
    }

    // 计算并缓存
    r.utf16Len = r.calculateUTF16Len()
    return r.utf16Len
}
```

#### 风险 2：Rope 拼接破坏平衡

**描述**：直接拼接可能导致树不平衡

**影响**：高

**缓解措施**：
- 实现再平衡逻辑
- 限制树深度
- 参考现有 `Insert()` 实现

```go
func (r *Rope) AppendRope(other *Rope) *Rope {
    // 检查深度
    if r.depth() + other.depth() > MAX_DEPTH {
        // 使用 Insert() 方法（包含再平衡）
        return r.Insert(r.Length(), other.String())
    }

    // 直接拼接
    return r.mergeTree(other)
}
```

### 6.2 兼容性风险

#### 风险 3：Grapheme vs Rune

**描述**：ropey 使用 rune，texere-rope 使用 grapheme

**影响**：低

**缓解措施**：
- 文档明确说明
- 提供两者转换
- 保持 API 一致性

### 6.3 维护风险

#### 风险 4：功能膨胀

**描述**：添加过多功能增加维护负担

**影响**：中等

**缓解措施**：
- 严格的优先级控制
- 定期审查低使用率功能
- 保持核心功能简洁

---

## 第七部分：成功标准

### 7.1 功能完整性

- [x] Helix 对齐度: 100% ✅
- [ ] Ropey 核心功能: 90%
- [ ] Ropey 高级功能: 60%
- [ ] 测试覆盖率: >40%

### 7.2 性能指标

- [ ] AppendRope vs Append: >2x 速度提升
- [ ] UTF-16 转换: <O(N) 时间复杂度
- [ ] Hash 计算: <1ms for 1MB 文本

### 7.3 质量指标

- [ ] 所有新功能有完整文档
- [ ] 所有新功能有单元测试
- [ ] 所有新功能有使用示例
- [ ] 零已知 bug

---

## 第八部分：总结与建议

### 8.1 当前状态总结

**已完成**：
1. ✅ Helix transaction 系统 - 100% 对齐
2. ✅ Grapheme 支持 - Unicode UAX #29
3. ✅ Chunk_at 方法 - 基础实现
4. ✅ Position mapping 优化 - O(M log M + N)
5. ✅ Time-based undo - 完整实现

**待完成**（按优先级）：
1. ⚠️ UTF-16 支持 - P0
2. ⚠️ 单字符操作 - P0
3. ⚠️ Hash 支持 - P0
4. ⚠️ CRLF 处理 - P0
5. ⭐ Rope 拼接 - P1
6. ⭐ 字节迭代器 - P1
7. ⭐ 索引转换 - P1

### 8.2 实施建议

**短期（1-2 周）**：
1. 实现 P0 核心功能
2. 完善测试覆盖
3. 更新文档

**中期（1-2 月）**：
1. 实现 P1 性能优化
2. 性能基准测试
3. 用户体验优化

**长期（按需）**：
1. 实现 P2 增强功能
2. 架构级优化（P3）
3. 生态系统集成

### 8.3 最终目标

通过实施本迁移计划，texere-rope 将：

1. **功能对齐**：
   - 100% Helix 对齐 ✅
   - ~90% Ropey 核心功能
   - ~60% Ropey 高级功能

2. **性能优势**：
   - 保持优于 ropey 的性能
   - 额外优化（对象池、惰性求值）

3. **易用性**：
   - 完整的 Go idiomatic API
   - 详尽的文档和示例
   - 良好的错误处理

4. **生产就绪**：
   - 完整的测试覆盖
   - 稳定的 API
   - 长期维护承诺

---

**文档版本**: 1.0
**最后更新**: 2026-01-31
**维护者**: texere-rope team
**参考**: [ropey](S:/src.editor/ropey) | [helix](S:/src.editor/helix)
