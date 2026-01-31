package rope

import (
	"fmt"
	"runtime"
	"time"
)

// ========== Memory Profiling ==========

// MemStats tracks memory allocations.
type MemStats struct {
	Alloc      uint64 // Bytes allocated
	TotalAlloc uint64 // Total bytes allocated
	Mallocs    uint64 // Number of allocations
	Frees      uint64 // Number of frees
}

// GetMemStats captures current memory statistics.
func GetMemStats() MemStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return MemStats{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Mallocs:     m.Mallocs,
		Frees:      m.Frees,
	}
}

// Diff calculates the difference between two MemStats.
func (m MemStats) Diff(other MemStats) MemStats {
	return MemStats{
		Alloc:      m.Alloc - other.Alloc,
		TotalAlloc: m.TotalAlloc - other.TotalAlloc,
		Mallocs:     m.Mallocs - other.Mallocs,
		Frees:      m.Frees - other.Frees,
	}
}

// String returns a string representation of MemStats.
func (m MemStats) String() string {
	return fmt.Sprintf("Alloc: %d bytes, TotalAlloc: %d bytes, Mallocs: %d, Frees: %d",
		m.Alloc, m.TotalAlloc, m.Mallocs, m.Frees)
}

// ProfileMemory profiles memory usage of a function.
func ProfileMemory(fn func()) MemStats {
	// Force GC before starting
	runtime.GC()
	runtime.ReadMemStats(&runtime.MemStats{})

	before := GetMemStats()
	fn()
	after := GetMemStats()

	return after.Diff(before)
}

// ========== Performance Profiling ==========

// BenchmarkResult holds benchmark results.
type BenchmarkResult struct {
	Name      string
	Duration  time.Duration
	Ops       int
	OpsPerSec float64
	MemStats  MemStats
}

// ProfileOperation profiles an operation with memory tracking.
func ProfileOperation(name string, iterations int, fn func()) BenchmarkResult {
	memStats := ProfileMemory(func() {
		start := time.Now()
		for i := 0; i < iterations; i++ {
			fn()
		}
		duration := time.Since(start)
		_ = duration // Avoid unused variable warning
	})

	return BenchmarkResult{
		Name:     name,
		Duration: 0, // Will be set by caller
		Ops:      iterations,
		MemStats: memStats,
	}
}

// ========== Memory Analysis ==========

// AnalyzeMemory analyzes memory usage patterns of a rope.
func (r *Rope) AnalyzeMemory() MemoryReport {
	if r == nil {
		return MemoryReport{}
	}

	report := MemoryReport{
		TotalSize:    r.Size(),
		TotalLength:  r.Length(),
		ChunkCount:   r.ChunkCount(),
		NodeCount:    r.NodeCount(),
		Depth:        r.Depth(),
	}

	// Analyze chunks
	it := r.Chunks()
	chunks := make([]ChunkMemoryInfo, 0, it.Count())
	totalOverhead := 0

	for it.Next() {
		chunk := it.Current()
		info := ChunkMemoryInfo{
			Size:     len(chunk),
			Overhead: estimateOverhead(chunk),
		}
		chunks = append(chunks, info)
		totalOverhead += info.Overhead
	}

	report.Chunks = chunks
	report.OverheadBytes = totalOverhead
	report.Efficiency = float64(report.TotalSize) / float64(report.TotalSize+totalOverhead)

	return report
}

// ChunkMemoryInfo holds memory info for a chunk.
type ChunkMemoryInfo struct {
	Size     int // Content size in bytes
	Overhead int // Estimated overhead in bytes
}

// MemoryReport reports detailed memory usage.
type MemoryReport struct {
	TotalSize     int                // Total content size in bytes
	TotalLength   int                // Total characters
	ChunkCount    int                // Number of chunks
	NodeCount     int                // Number of tree nodes
	Depth         int                // Tree depth
	Chunks        []ChunkMemoryInfo  // Chunk details
	OverheadBytes int                // Total overhead in bytes
	Efficiency    float64            // Memory efficiency (0-1)
}

// String returns a formatted memory report.
func (mr MemoryReport) String() string {
	return fmt.Sprintf(
		"Memory Report:\n"+
			"  Total Size: %d bytes\n"+
			"  Total Length: %d chars\n"+
			"  Chunks: %d\n"+
			"  Nodes: %d\n"+
			"  Depth: %d\n"+
			"  Overhead: %d bytes (%.1f%%)\n"+
			"  Efficiency: %.1f%%",
		mr.TotalSize,
		mr.TotalLength,
		mr.ChunkCount,
		mr.NodeCount,
		mr.Depth,
		mr.OverheadBytes,
		float64(mr.OverheadBytes)/float64(mr.TotalSize)*100,
		mr.Efficiency*100,
	)
}

// estimateOverhead estimates memory overhead for a chunk.
func estimateOverhead(s string) int {
	// Go string overhead: 16 bytes (pointer + len + cap)
	// Plus allocation alignment
	const stringOverhead = 16
	const alignmentPadding = 8

	size := len(s)
	total := stringOverhead + size
	aligned := (total + alignmentPadding - 1) / alignmentPadding * alignmentPadding

	return aligned - size
}

// ========== Performance Analysis ==========

// PerformanceAnalysis holds comprehensive performance metrics.
type PerformanceAnalysis struct {
	MemoryReport MemoryReport
	Benchmarks   map[string]BenchmarkResult
	Recommendations []string
}

// AnalyzePerformance performs comprehensive performance analysis.
func (r *Rope) AnalyzePerformance() PerformanceAnalysis {
	analysis := PerformanceAnalysis{
		MemoryReport: r.AnalyzeMemory(),
		Benchmarks:   make(map[string]BenchmarkResult),
		Recommendations: []string{},
	}

	// Run benchmarks
	const iterations = 1000

	// Read operations
	analysis.Benchmarks["String"] = ProfileOperation("String", iterations, func() {
		_ = r.String()
	})

	analysis.Benchmarks["Length"] = ProfileOperation("Length", iterations, func() {
		_ = r.Length()
	})

	analysis.Benchmarks["CharAt"] = ProfileOperation("CharAt", iterations, func() {
		if r.Length() > 0 {
			_ = r.CharAt(r.Length() / 2)
		}
	})

	analysis.Benchmarks["Slice"] = ProfileOperation("Slice", iterations, func() {
		if r.Length() > 10 {
			_ = r.Slice(r.Length()/4, r.Length()*3/4)
		}
	})

	// Iteration
	analysis.Benchmarks["Iterator"] = ProfileOperation("Iterator", iterations, func() {
		it := r.NewIterator()
		for it.Next() {
			_ = it.Current()
		}
	})

	// Generate recommendations
	analysis.generateRecommendations()

	return analysis
}

// generateRecommendations analyzes performance and generates recommendations.
func (pa *PerformanceAnalysis) generateRecommendations() {
	mr := pa.MemoryReport

	// Check efficiency
	if mr.Efficiency < 0.7 {
		pa.Recommendations = append(pa.Recommendations,
			"Low memory efficiency detected (<70%)")
	}

	// Check chunk count vs size
	if mr.ChunkCount > mr.TotalSize/512 {
		pa.Recommendations = append(pa.Recommendations,
			"Too many small chunks - consider merging")
	}

	// Check depth
	if mr.Depth > 20 {
		pa.Recommendations = append(pa.Recommendations,
			"Tree too deep - consider rebalancing")
	}

	// Check overhead
	if mr.OverheadBytes > mr.TotalSize {
		pa.Recommendations = append(pa.Recommendations,
			"Memory overhead exceeds content size")
	}

	if len(pa.Recommendations) == 0 {
		pa.Recommendations = append(pa.Recommendations,
			"No issues detected - performance looks good!")
	}
}

// Print prints a formatted performance analysis.
func (pa PerformanceAnalysis) Print() {
	fmt.Println("=== Performance Analysis ===")
	fmt.Println()
	fmt.Println(pa.MemoryReport.String())
	fmt.Println()

	fmt.Println("=== Benchmark Results ===")
	for name, result := range pa.Benchmarks {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  Allocations: %d bytes total\n", result.MemStats.TotalAlloc)
		fmt.Printf("  Mallocs: %d\n", result.MemStats.Mallocs)
	}

	fmt.Println()
	fmt.Println("=== Recommendations ===")
	for i, rec := range pa.Recommendations {
		fmt.Printf("%d. %s\n", i+1, rec)
	}
}

// ========== Optimization Detection ==========

// DetectIssues detects common performance issues.
func (r *Rope) DetectIssues() []Issue {
	issues := []Issue{}

	if r == nil || r.Length() == 0 {
		return issues
	}

	// Check for excessive chunk fragmentation
	it := r.Chunks()
	smallChunks := 0
	totalSize := 0

	for it.Next() {
		chunk := it.Current()
		if len(chunk) < 64 { // Less than 64 bytes
			smallChunks++
		}
		totalSize += len(chunk)
	}

	if smallChunks > it.Count()/2 {
		issues = append(issues, Issue{
			Type:    "Fragmentation",
			Severity: "High",
			Message: fmt.Sprintf("Too many small chunks: %d/%d chunks are < 64 bytes",
				smallChunks, it.Count()),
		})
	}

	// Check for unbalanced tree
	depth := r.Depth()
	expectedDepth := expectedDepth(r.Length())
	if depth > expectedDepth*2 {
		issues = append(issues, Issue{
			Type:    "Balance",
			Severity: "Medium",
			Message: fmt.Sprintf("Tree is unbalanced: depth %d vs expected %d",
				depth, expectedDepth),
		})
	}

	// Check for inefficient string conversions
	// This is detected through profiling
	issues = append(issues, r.profileStringConversion()...)

	return issues
}

// Issue represents a performance issue.
type Issue struct {
	Type     string
	Severity string
	Message  string
}

// expectedDepth calculates expected tree depth for balanced tree.
func expectedDepth(n int) int {
	if n <= 1 {
		return 1
	}
	// Log2 of number of leaf nodes
	depth := 0
	for n > 1 {
		n /= 2
		depth++
	}
	return depth
}

// profileStringConversion detects string conversion issues.
func (r *Rope) profileStringConversion() []Issue {
	issues := []Issue{}

	// Measure String() call memory
	memStats := ProfileMemory(func() {
		_ = r.String()
	})

	// String() should allocate at most the rope size + overhead
	expectedAlloc := uint64(r.Size() + r.Size()/10) // Allow 10% overhead
	if memStats.TotalAlloc > expectedAlloc*2 {
		issues = append(issues, Issue{
			Type:     "Memory",
			Severity: "High",
			Message: fmt.Sprintf("String() allocates %d bytes, expected < %d",
				memStats.TotalAlloc, expectedAlloc),
		})
	}

	return issues
}
