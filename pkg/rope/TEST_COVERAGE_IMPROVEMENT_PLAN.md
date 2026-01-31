# 测试覆盖率提升方案

> **日期**: 2026-01-31
> **当前覆盖率**: 41.7% of statements
> **目标覆盖率**: 70%+ of statements
> **策略**: 迭代式提升，分阶段执行

---

## 📊 当前覆盖率分析

### 覆盖率最低的文件 (0% 测试)

```
balance.go          - 平衡和优化相关功能
byte_cache.go       - 字节缓存和节点池
bytes_iter.go       - 字节迭代器
byte_char_conv.go   - 字节/字符转换
char_ops.go         - 字符操作
pools.go            - 对象池管理
pool.go             - 节点池
```

### 中等覆盖率文件 (10-40%)

```
builder.go          - Rope 构建器
savepoint_enhanced.go - 增强的保存点功能
transaction.go      - 事务处理
balance.go          - 部分功能有测试
line_ops.go         - 行操作
```

### 高覆盖率文件 (60%+)

```
rope.go             - 核心 Rope 实现
document.go         - 文档接口
rope_concat.go      - 连接操作
rope_io.go          - I/O 操作
```

---

## 🎯 改进策略

### 阶段 1: 快速胜利 (Quick Wins) - 目标: 41.7% → 50%

**优先级**: 高
**时间估计**: 2-3 小时
**文件**: 简单工具函数

#### 1.1 字符/字节转换测试
**文件**: `byte_char_conv.go`, `char_ops.go`

需要添加的测试:
```go
func TestByteToRune(t *testing.T)
func TestRuneToByte(t *testing.T)
func TestIsASCII(t *testing.T)
func TestCharOperations(t *testing.T)
```

**预期提升**: +3-5%

#### 1.2 字符串工具测试
**文件**: `str_utils.go`

需要添加的测试:
```go
func TestCommonPrefix(t *testing.T)
func TestCommonSuffix(t *testing.T)
func TestSplitLines(t *testing.T)
func TestTrimSpace(t *testing.T)
```

**预期提升**: +2-3%

#### 1.3 哈希功能测试
**文件**: `hash.go`

需要添加的测试:
```go
func TestHashRope(t *testing.T)
func TestHashString(t *testing.T)
func TestHashCollision(t *testing.T)
```

**预期提升**: +1-2%

---

### 阶段 2: 核心功能完善 - 目标: 50% → 60%

**优先级**: 高
**时间估计**: 4-6 小时
**文件**: 迭代器和优化功能

#### 2.1 字节迭代器测试
**文件**: `bytes_iter.go`

需要添加的测试:
```go
func TestBytesIterator_Basic(t *testing.T)
func TestBytesIterator_Seek(t *testing.T)
func TestBytesIterator_Peek(t *testing.T)
func TestBytesIterator_Collect(t *testing.T)
func TestBytesIterator_EdgeCases(t *testing.T)
```

**预期提升**: +3-4%

#### 2.2 反向迭代器测试
**文件**: `reverse_iter.go`

需要添加的测试:
```go
func TestReverseIterator_Basic(t *testing.T)
func TestReverseIterator_Seek(t *testing.T)
func TestReverseIterator_Skip(t *testing.T)
func TestReverseIterator_Collect(t *testing.T)
```

**预期提升**: +2-3%

#### 2.3 优化操作测试
**文件**: `micro_optimizations.go`, `zero_alloc_ops.go`

需要添加的测试:
```go
func TestInsertOptimized(t *testing.T)
func TestDeleteOptimized(t *testing.T)
func TestInsertZeroAlloc(t *testing.T)
func TestDeleteZeroAlloc(t *testing.T)
func TestBatchInsert(t *testing.T)
func TestBatchDelete(t *testing.T)
```

**预期提升**: +4-5%

---

### 阶段 3: 高级功能覆盖 - 目标: 60% → 70%

**优先级**: 中
**时间估计**: 6-8 小时
**文件**: 复杂功能

#### 3.1 Builder 测试
**文件**: `builder.go`

需要添加的测试:
```go
func TestBuilder_Replace(t *testing.T)
func TestBuilder_ResetFromRope(t *testing.T)
func TestBuilder_InsertString(t *testing.T)
func TestBuilder_InsertRune(t *testing.T)
func TestBuilder_InsertByte(t *testing.T)
func TestBuilder_AppendRune(t *testing.T)
func TestBuilder_AppendByte(t *testing.T)
func TestBuilder_Pool(t *testing.T)
```

**预期提升**: +3-4%

#### 3.2 Balance 测试
**文件**: `balance.go`

需要添加的测试:
```go
func TestOptimize(t *testing.T)
func TestCompact(t *testing.T)
func TestValidate(t *testing.T)
func TestSuggestedConfig(t *testing.T)
func TestRebalanceTree(t *testing.T)
func TestMergeLeaves(t *testing.T)
```

**预期提升**: +4-5%

#### 3.3 SavePoint 测试
**文件**: `savepoint_enhanced.go`

需要添加的测试:
```go
func TestQueryPreallocated(t *testing.T)
func TestQueryByTime(t *testing.T)
func TestQueryByTag(t *testing.T)
func TestQueryConcurrent(t *testing.T)
func TestSavePointStats(t *testing.T)
```

**预期提升**: +3-4%

#### 3.4 Transaction 测试
**文件**: `transaction.go`, `transaction_advanced.go`

需要添加的测试:
```go
func TestTransaction_Invert(t *testing.T)
func TestTransaction_Compose(t *testing.T)
func TestChangeBySelection(t *testing.T)
func TestPositionMapping(t *testing.T)
func TestMultipleSelections(t *testing.T)
```

**预期提升**: +3-4%

---

### 阶段 4: 分支覆盖率提升 - 目标: 70% → 80%

**优先级**: 中
**时间估计**: 8-10 小时
**策略**: 边界条件和错误路径

#### 4.1 边界条件测试
为所有现有测试添加边界情况:
- 空字符串
- 单字符
- 超长字符串
- 特殊字符 (Unicode, emoji, 组合字符)
- nil 检查
- 越界访问

**预期提升**: +5-7%

#### 4.2 错误路径测试
```go
func TestErrorCases_InvalidPosition(t *testing.T)
func TestErrorCases_InvalidRange(t *testing.T)
func TestErrorCases_NilRope(t *testing.T)
func TestErrorCases_EmptyRope(t *testing.T)
func TestErrorCases_UTF8Errors(t *testing.T)
```

**预期提升**: +3-5%

#### 4.3 并发测试
```go
func TestConcurrent_Reads(t *testing.T)
func TestConcurrent_Writes(t *testing.T)
func TestConcurrent_Mixed(t *testing.T)
func TestRaceConditions(t *testing.T)
```

**预期提升**: +2-3%

---

## 🔧 实施策略

### 迭代式执行流程

```
┌─────────────────────────────────────────────┐
│  1. 生成当前覆盖率报告                        │
│  2. 选择最低覆盖率的文件                      │
│  3. 编写测试用例                              │
│  4. 运行测试并验证覆盖率提升                    │
│  5. 提交代码                                  │
│  6. 重复 1-5 直到达到目标                      │
└─────────────────────────────────────────────┘
```

### 自动化工具

使用 `go test` 的覆盖率功能:
```bash
# 生成覆盖率报告
go test ./pkg/rope -coverprofile=coverage.out -covermode=atomic

# 查看函数级别覆盖率
go tool cover -func=coverage.out | sort -k3 -n

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 按文件汇总
go tool cover -func=coverage.out | grep "\.go:" | awk '...' | sort
```

---

## 📋 测试模板

### 基础测试模板

```go
package rope

import "testing"

func TestFeatureName_Basic(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"empty string", "", ""},
        {"single char", "a", "a"},
        {"simple text", "hello", "hello"},
        {"with unicode", "你好世界", "你好世界"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            r := New(tt.input)
            result := r.Feature()
            if result != tt.expected {
                t.Errorf("got %q, want %q", result, tt.expected)
            }
        })
    }
}
```

### 边界条件测试模板

```go
func TestFeatureName_EdgeCases(t *testing.T) {
    t.Run("nil rope", func(t *testing.T) {
        var r *Rope
        // 测试 nil 情况
    })

    t.Run("empty rope", func(t *testing.T) {
        r := Empty()
        // 测试空字符串
    })

    t.Run("single character", func(t *testing.T) {
        r := New("a")
        // 测试单字符
    })

    t.Run("out of bounds", func(t *testing.T) {
        r := New("hello")
        // 测试越界
        assert.Panics(t, func() {
            r.CharAt(100)
        })
    })
}
```

### 并发测试模板

```go
func TestFeatureName_Concurrent(t *testing.T) {
    const goroutines = 100
    const operations = 1000

    r := New("initial text")

    var wg sync.WaitGroup
    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < operations; j++ {
                _ = r.Length() // 只读操作
            }
        }(i)
    }

    wg.Wait()
    // 验证结果
}
```

---

## 📊 进度跟踪

### 里程碑

| 阶段 | 目标覆盖率 | 预计时间 | 状态 |
|------|-----------|----------|------|
| 阶段 1 | 50% | 2-3h | 待开始 |
| 阶段 2 | 60% | 4-6h | 待开始 |
| 阶段 3 | 70% | 6-8h | 待开始 |
| 阶段 4 | 80% | 8-10h | 待开始 |

### 每日检查清单

- [ ] 运行完整测试套件
- [ ] 生成覆盖率报告
- [ ] 识别最低覆盖率文件
- [ ] 编写至少 3 个新测试
- [ ] 验证覆盖率提升
- [ ] 提交代码

---

## 🎯 成功标准

### 数值目标

- **语句覆盖率**: 从 41.7% → 70%+
- **分支覆盖率**: 从当前 → 60%+
- **文件覆盖率**: 所有文件 > 50%

### 质量目标

- 所有新测试必须通过
- 没有测试代码重复
- 测试代码清晰易读
- 包含边界条件和错误路径

---

## 📚 参考资源

- [Go Testing Blog](https://blog.golang.org/cover)
- [Effective Go: Testing](https://golang.org/doc/effective_go#testing)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

**文档版本**: 1.0
**创建日期**: 2026-01-31
**状态**: 准备执行
