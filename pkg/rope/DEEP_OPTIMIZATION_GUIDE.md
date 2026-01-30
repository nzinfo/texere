# Deep Performance Optimization Guide

> **Date**: 2026-01-31
> **Purpose**: Provide actionable optimization strategies for identified performance bottlenecks
> **Status**: Ready for implementation

---

## 📊 Performance Analysis Summary

Based on the baseline benchmarks established earlier, we've identified the following optimization opportunities:

| Component | Current Performance | Issue | Expected Improvement |
|-----------|-------------------|-------|---------------------|
| **FromReader** | 9-45 allocs | String conversion per chunk | 30-40% reduction |
| **WriteTo (Large)** | 368 KB alloc | Full string conversion | 90%+ reduction |
| **RopeReader** | 3-37 allocs | Iterator recreation | 50-70% reduction |
| **Manager.Query** | 111 allocs | Slice allocation + Mutex | 60-80% reduction |

---

## 🔧 Optimization 1: FromReader with unsafe.String

### Problem

Current implementation (`rope_io.go:30`):
```go
b.Append(string(buf[:n]))  // Allocates new string per chunk
```

**Performance Impact**: 9 allocations for small files, 45 for large files

### Solution

Add `AppendBytes` method to `RopeBuilder`:

```go
// AppendBytes appends a byte slice without string allocation.
func (b *RopeBuilder) AppendBytes(data []byte) *RopeBuilder {
    if len(data) == 0 {
        return b
    }

    // Unsafe but efficient conversion
    // Safe because rope copies data internally
    str := unsafe.String(unsafe.SliceData(data), len(data))

    b.pending = append(b.pending, pendingInsert{
        position: -1,
        text:     str,
    })
    return b
}
```

### Expected Results

- **Before**: 9,984 B, 9 allocs (Medium)
- **After**: ~6,000 B, ~6 allocs (30-35% improvement)

### Implementation Steps

1. Add `AppendBytes` to `builder.go`
2. Create `FromReaderOptimized` using `AppendBytes`
3. Add benchmarks to validate improvement
4. Update documentation with usage notes

---

## 🔧 Optimization 2: WriteTo with Chunked Writing

### Problem

Current implementation (`rope_io.go:53-54`):
```go
str := r.String()           // Allocates entire string (368 KB for large files)
return writer.Write([]byte(str))
```

**Performance Impact**: 368,801 B for large files (10,000 lines)

### Solution

Implement chunked writing to avoid large allocations:

```go
func (r *Rope) WriteToChunked(writer io.Writer, chunkSize int) (int, error) {
    if chunkSize <= 0 {
        chunkSize = 4096 // 4KB default
    }

    total := 0
    iter := r.IterBytes()
    buf := make([]byte, 0, chunkSize)

    for iter.Next() {
        b := iter.Current()
        buf = append(buf, b)

        if len(buf) >= chunkSize {
            n, err := writer.Write(buf)
            total += n
            if err != nil {
                return total, err
            }
            buf = buf[:0]
        }
    }

    // Write remaining
    if len(buf) > 0 {
        n, err := writer.Write(buf)
        total += n
        if err != nil {
            return total, err
        }
    }

    return total, nil
}
```

### Expected Results

- **Before**: 368,801 B, 6 allocs (Large)
- **After**: ~4,096 B (buffer size), 2-3 allocs (95%+ reduction)

### Implementation Steps

1. Add `WriteToChunked` to `rope_io.go`
2. Keep `WriteTo` for backward compatibility
3. Add benchmarks comparing both approaches
4. Document when to use each method

---

## 🔧 Optimization 3: RopeReader with Cached Iterator

### Problem

Current implementation (`rope_io.go:100-101`):
```go
bytes := rr.rope.IterBytes()  // Creates new iterator
bytes.Seek(rr.pos)           // Seeks to position
```

**Performance Impact**: 3-7 allocs per read operation

### Solution

Cache iterator in the reader struct:

```go
type optimizedRopeReader struct {
    rope     *Rope
    iter     *BytesIterator
    pos      int
    mu       sync.Mutex
    once     sync.Once
}

func (orr *optimizedRopeReader) Read(p []byte) (int, error) {
    orr.mu.Lock()
    defer orr.mu.Unlock()

    if orr.pos >= orr.rope.Size() {
        return 0, io.EOF
    }

    // Initialize iterator once
    orr.once.Do(func() {
        orr.iter = orr.rope.IterBytes()
    })

    // Seek to position (only if needed)
    if orr.pos > 0 && orr.iter != nil {
        orr.iter.Seek(orr.pos)
    }

    // Read bytes...
}
```

### Expected Results

- **Before**: 515,473 B, 37 allocs (Large)
- **After**: ~50,000 B, 1-2 allocs (90%+ reduction)

### Implementation Steps

1. Add `optimizedRopeReader` struct
2. Add `ReaderOptimized()` method to Rope
3. Add benchmarks comparing both
4. Keep original `Reader()` for backward compatibility

---

## 🔧 Optimization 4: Query with RWMutex + Pre-allocated Slices

### Problem

Current implementation uses:
- `sync.Mutex` - blocks all reads during writes
- Allocates new slice per query

**Performance Impact**: 111 allocs per query

### Solution Part 1: Use RWMutex

```go
type EnhancedSavepointManagerV2 struct {
    savepoints map[int]*EnhancedSavePoint
    mu         sync.RWMutex  // Changed from Mutex
    // ...
}

func (sm *EnhancedSavePointManagerV2) Query(query SavePointQuery) []SavePointResult {
    sm.mu.RLock()  // Allow concurrent reads
    defer sm.mu.RUnlock()
    // ... query logic
}
```

**Expected**: 2-5x better throughput for concurrent queries

### Solution Part 2: Pre-allocated Slices

```go
// Method 1: Use sync.Pool
var resultsPool = sync.Pool{
    New: func() interface{} {
        s := make([]SavePointResult, 0, 16)
        return &s
    },
}

func queryWithPool() []SavePointResult {
    results := resultsPool.Get().(*[]SavePointResult)
    results = (*results)[:0]
    // ... fill results
    return results
}

// Method 2: Pre-allocated slice parameter
func QueryPreallocated(query SavePointQuery, results []SavePointResult) []SavePointResult {
    if results == nil {
        results = make([]SavePointResult, 0, 16)
    }
    results = results[:0]
    // ... fill results
    return results
}
```

**Expected**: 60-80% reduction in allocations

### Implementation Steps

1. Create `EnhancedSavePointManagerV2` with RWMutex
2. Add `QueryOptimized` with sync.Pool
3. Add `QueryPreallocated` for caller-provided slices
4. Add benchmarks for comparison

---

## 📋 Implementation Priority

### Phase 1: High Impact, Low Risk ✅

1. **WriteTo Chunked** - Easiest to implement
   - Files: Modify `rope_io.go`
   - Risk: Low
   - Impact: High (90%+ memory reduction for large files)

2. **RopeReader Optimized** - Moderate effort
   - Files: Add to `rope_io.go`
   - Risk: Low
   - Impact: High (90%+ allocation reduction)

### Phase 2: Medium Impact, Moderate Risk

3. **FromReader with AppendBytes** - Requires careful testing
   - Files: Modify `builder.go`, `rope_io.go`
   - Risk: Medium (unsafe requires validation)
   - Impact: Medium (30-40% improvement)

4. **Query Optimization** - More complex
   - Files: New file or modify `savepoint_enhanced.go`
   - Risk: Medium
   - Impact: High for concurrent workloads

---

## 🧪 Testing Strategy

### Before Implementation

```bash
# Establish baseline
go test ./pkg/rope -bench=. -benchmem -run=^$ > baseline.txt
```

### After Implementation

```bash
# Compare with baseline
go test ./pkg/rope -bench=. -benchmem -run=^$ > optimized.txt
benchstat baseline.txt optimized.txt
```

### Regression Prevention

Add to CI/CD:
```yaml
performance_test:
  script:
    - go test ./pkg/rope -bench=BenchmarkWriteTo_Large -benchmem
  # Fail if >10% degradation
```

---

## 📊 Expected Overall Impact

### Memory Allocation Reduction

| Component | Before | After | Improvement |
|-----------|--------|-------|-------------|
| **FromReader** | 9,984 B | ~6,500 B | -35% |
| **WriteTo (Large)** | 368,801 B | ~4,100 B | **-99%** |
| **RopeReader (Large)** | 515,473 B | ~50,000 B | **-90%** |
| **Manager.Query** | 35,992 B | ~10,000 B | **-72%** |

### Performance Improvements

| Scenario | Before | After | Improvement |
|----------|--------|-------|-------------|
| **Large File Write** | 199,293 ns | ~80,000 ns | 2.5x faster |
| **Sequential Reads** | 485,347 ns | ~200,000 ns | 2.4x faster |
| **Concurrent Queries** | 18,000 ns | ~5,000 ns | 3.6x faster |

---

## 🎯 Recommendations

### Immediate (Recommended)

1. ✅ **Implement WriteTo Chunked**
    - **Why**: Easiest, biggest impact (90% memory reduction)
    - **Effort**: 1-2 hours
    - **Risk**: Very low

2. ✅ **Implement RopeReader Optimized**
    - **Why**: Large impact, low risk
    - **Effort**: 2-3 hours
    - **Risk**: Low

### Short Term (1-2 Weeks)

3. ⚠️ **Implement Query Optimization**
    - **Why**: Important for high-concurrency scenarios
    - **Effort**: 4-6 hours
    - **Risk**: Medium
    - **Note**: Keep both versions, use optimized only when needed

### Future (As Needed)

4. ⏭️ **Implement AppendBytes**
    - **Why**: Moderate improvement, but uses unsafe
    - **Effort**: 3-4 hours
    - **Risk**: Medium (requires extensive testing)
    - **Note**: Document safety guarantees clearly

---

## 🔐 Safety Considerations

### unsafe.String Usage

When using `unsafe.String` in `AppendBytes`:

**Safety Guarantees**:
1. Rope copies data internally into tree nodes
2. Byte slice is not modified after conversion
3. Each chunk is processed independently

**Validation**:
```go
// Test that AppendBytes produces same result as Append
func TestAppendBytes_Equivalent(t *testing.T) {
    r1 := New("")
    r2 := New("")

    data := []byte("Hello World")

    b1 := NewBuilder()
    b1.Append(string(data))
    r1 = b1.Build()

    b2 := NewBuilder()
    b2.AppendBytes(data)
    r2 = b2.Build()

    assert.Equal(t, r1.String(), r2.String())
}
```

---

## 📝 API Design

### Backward Compatibility

Keep existing APIs, add new `*Optimized` variants:

```go
// Existing API (unchanged)
func FromReader(reader io.Reader) (*Rope, error)
func (r *Rope) WriteTo(writer io.Writer) (int, error)
func (r *Rope) Reader() io.Reader

// New optimized variants
func FromReaderOptimized(reader io.Reader) (*Rope, error)
func (r *Rope) WriteToChunked(writer io.Writer, chunkSize int) (int, error)
func (r *Rope) ReaderOptimized() io.Reader
```

### Usage Recommendations

```go
// Default: Use standard API for normal use
rope, _ := FromReader(file)

// Optimized: Use for large files or performance-critical paths
rope, _ := FromReaderOptimized(largeFile)

// Large file writing
rope.WriteToChunked(writer, 64*1024)  // 64KB chunks

// Sequential reading
reader := rope.ReaderOptimized()
io.ReadAll(reader)
```

---

## 🚀 Next Steps

1. **Review this document** - Understand optimization strategies
2. **Choose priority** - Decide which optimizations to implement
3. **Create implementation plan** - Break down into tasks
4. **Implement incrementally** - One optimization at a time
5. **Benchmark each change** - Validate improvements
6. **Update documentation** - Add usage guidelines

---

## 📞 Contact

For questions or clarifications about these optimizations:
- Review PERFORMANCE_OPTIMIZATION_REPORT.md for baseline data
- Check perf_baseline_test.go for current benchmarks
- Refer to PERFORMANCE_OPTIMIZATION_FINAL.md for overall summary

---

**Document Version**: 1.0
**Last Updated**: 2026-01-31
**Status**: Ready for Implementation
