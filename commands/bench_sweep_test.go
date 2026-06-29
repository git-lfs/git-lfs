package commands

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/v3/config"
	"github.com/git-lfs/git-lfs/v3/git/gitattr"
	"github.com/git-lfs/git-lfs/v3/lfs"
)

func setupAutoTrackBench(b *testing.B, gitattributesContent string) func() {
	b.Helper()
	dir := b.TempDir()

	err := os.MkdirAll(filepath.Join(dir, ".git", "lfs", "tmp"), 0755)
	if err != nil {
		b.Fatal(err)
	}

	if gitattributesContent != "" {
		err = os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(gitattributesContent), 0644)
		if err != nil {
			b.Fatal(err)
		}
	}

	saved := cfg
	cfg = config.NewIn(dir, "")
	return func() { cfg = saved }
}

// --- GetAutoTrackSize: sweep pattern count ---

func BenchmarkSweepGetAutoTrackSize(b *testing.B) {
	patternCounts := []int{1, 2, 5, 10, 20, 50, 100, 200, 500}
	targetFile := "target.dat"

	for _, n := range patternCounts {
		// Build .gitattributes with n patterns; target pattern at position 0
		b.Run(fmt.Sprintf("match-first-%d", n), func(b *testing.B) {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=100000\n"))
			for i := 1; i < n; i++ {
				sb.WriteString(fmt.Sprintf("*.ext%d filter=lfs autotracksize=%d\n", i, i*1000))
			}
			cleanup := setupAutoTrackBench(b, sb.String())
			defer cleanup()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), targetFile)
				if !ok || size != 100000 {
					b.Fatalf("unexpected: ok=%v size=%d", ok, size)
				}
			}
		})

		// Target at the end of the pattern list (worst-case match)
		b.Run(fmt.Sprintf("match-last-%d", n), func(b *testing.B) {
			var sb strings.Builder
			for i := 0; i < n-1; i++ {
				sb.WriteString(fmt.Sprintf("*.ext%d filter=lfs autotracksize=%d\n", i, i*1000))
			}
			sb.WriteString(fmt.Sprintf("*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=999999\n"))
			cleanup := setupAutoTrackBench(b, sb.String())
			defer cleanup()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), targetFile)
				if !ok || size != 999999 {
					b.Fatalf("unexpected: ok=%v size=%d", ok, size)
				}
			}
		})

		// No match (scan all patterns)
		b.Run(fmt.Sprintf("no-match-%d", n), func(b *testing.B) {
			var sb strings.Builder
			for i := 0; i < n; i++ {
				sb.WriteString(fmt.Sprintf("*.ext%d filter=lfs autotracksize=%d\n", i, i*1000))
			}
			cleanup := setupAutoTrackBench(b, sb.String())
			defer cleanup()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), targetFile)
				if ok || size != 0 {
					b.Fatalf("expected no match: ok=%v size=%d", ok, size)
				}
			}
		})
	}
}

// --- Sweep: .gitattributes file sizes directly ---

func BenchmarkSweepGitattributesRead(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000, 10000, 50000} // bytes of .gitattributes

	for _, sz := range sizes {
		b.Run(fmt.Sprintf("read-%d-bytes", sz), func(b *testing.B) {
			// Build .gitattributes content of approximately sz bytes
			var sb strings.Builder
			for sb.Len() < sz {
				sb.WriteString("*.ext000 filter=lfs autotracksize=1000\n")
			}
			attrContent := sb.String()
			cleanup := setupAutoTrackBench(b, attrContent)
			defer cleanup()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), "target.dat")
				if ok {
					b.Fatal("expected no match")
				}
			}
		})
	}
}

// --- Clean pass-through: sweep file size ---

func BenchmarkSweepCleanPassThrough(b *testing.B) {
	threshold := int64(100 * 1 << 20)

	sizes := []int{1 << 10, 10 << 10, 100 << 10, 1 << 20, 10 << 20, 50 << 20}

	for _, sz := range sizes {
		szLabel := fmt.Sprintf("%.0fKB", float64(sz)/(1<<10))
		if sz >= 1<<20 {
			szLabel = fmt.Sprintf("%.0fMB", float64(sz)/(1<<20))
		}
		b.Run(szLabel, func(b *testing.B) {
			cleanup := setupAutoTrackBench(b, fmt.Sprintf("* autotracksize=%d", threshold))
			defer cleanup()

			gf := lfs.NewGitFilter(cfg)
			data := make([]byte, sz)
			var to bytes.Buffer

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				to.Reset()
				from := bytes.NewReader(data)
				ptr, err := clean(gf, &to, from, "test.dat", -1)
				if ptr != nil || err != nil {
					b.Fatalf("expected pass-through: ptr=%v err=%v", ptr, err)
				}
				if to.Len() != len(data) {
					b.Fatalf("output size mismatch")
				}
			}
		})
	}
}

// --- Clean pass-through at different threshold ratios (file size as % of threshold) ---

func BenchmarkSweepCleanThresholdRatio(b *testing.B) {
	threshold := int64(1 << 20)
	ratios := []float64{0.1, 0.5, 0.9, 0.99, 0.999}

	for _, r := range ratios {
		sz := int(float64(threshold) * r)
		b.Run(fmt.Sprintf("ratio-%.3f", r), func(b *testing.B) {
			cleanup := setupAutoTrackBench(b, fmt.Sprintf("* autotracksize=%d", threshold))
			defer cleanup()

			gf := lfs.NewGitFilter(cfg)
			data := make([]byte, sz)
			var to bytes.Buffer

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				to.Reset()
				from := bytes.NewReader(data)
				ptr, err := clean(gf, &to, from, "test.dat", -1)
				if ptr != nil || err != nil {
					b.Fatalf("expected pass-through: ptr=%v err=%v", ptr, err)
				}
				if to.Len() != sz {
					b.Fatalf("output size mismatch: got %d want %d", to.Len(), sz)
				}
			}
		})
	}
}

// --- Clean over-threshold: file exceeds threshold, goes to gf.Clean ---

func BenchmarkSweepCleanOverThreshold(b *testing.B) {
	sizes := []int{1 << 20, 5 << 20, 10 << 20, 25 << 20}

	for _, sz := range sizes {
		threshold := int64(sz / 2)
		szLabel := fmt.Sprintf("%.0fMB", float64(sz)/(1<<20))
		b.Run(szLabel, func(b *testing.B) {
			cleanup := setupAutoTrackBench(b, fmt.Sprintf("* autotracksize=%d", threshold))
			defer cleanup()

			gf := lfs.NewGitFilter(cfg)
			data := make([]byte, sz)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var to bytes.Buffer
				from := bytes.NewReader(data)
				_, _ = clean(gf, &to, from, "test.dat", -1)
			}
		})
	}
}

// --- findLargeFiles stat loop: sweep file count ---

func BenchmarkSweepFindLargeFilesStat(b *testing.B) {
	counts := []int{100, 500, 1000, 5000, 10000, 25000, 50000}

	for _, n := range counts {
		b.Run(fmt.Sprintf("%d-files", n), func(b *testing.B) {
			dir := b.TempDir()
			paths := make([]string, n)
			for i := 0; i < n; i++ {
				paths[i] = filepath.Join(dir, fmt.Sprintf("file%07d.dat", i))
				if err := os.WriteFile(paths[i], []byte("x"), 0644); err != nil {
					b.Fatal(err)
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				count := 0
				for _, path := range paths {
					info, err := os.Stat(path)
					if err == nil && info.Size() >= 0 {
						count++
					}
				}
				if count != n {
					b.Fatalf("expected %d files, got %d", n, count)
				}
			}
		})
	}
}

// --- Cumulative autotrack: vary number of sequential GetAutoTrackSize calls ---

func BenchmarkSweepSequentialGetAutoTrackSize(b *testing.B) {
	callCounts := []int{1, 10, 100, 500, 1000}

	for _, n := range callCounts {
		b.Run(fmt.Sprintf("%d-calls-match", n), func(b *testing.B) {
			cleanup := setupAutoTrackBench(b, "*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=100000")
			defer cleanup()

			// distinct file paths to avoid any OS caching ambiguity
			paths := make([]string, n)
			for i := range paths {
				paths[i] = fmt.Sprintf("file%07d.dat", i)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, p := range paths {
					size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), p)
					if !ok || size != 100000 {
						b.Fatalf("bad result for %s", p)
					}
				}
			}
		})

		b.Run(fmt.Sprintf("%d-calls-no-match", n), func(b *testing.B) {
			cleanup := setupAutoTrackBench(b, "*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=100000")
			defer cleanup()

			paths := make([]string, n)
			for i := range paths {
				paths[i] = fmt.Sprintf("file%07d.txt", i)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, p := range paths {
					size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), p)
					if ok || size != 0 {
						b.Fatalf("bad result for %s", p)
					}
				}
			}
		})
	}
}

// --- Variable .gitattributes with very long lines (large autotracksize values) ---

func BenchmarkSweepGitattributesExtreme(b *testing.B) {
	// Extremely large autotracksize values
	b.Run("large-values", func(b *testing.B) {
		attrContent := fmt.Sprintf("*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=%d\n", math.MaxInt64)
		cleanup := setupAutoTrackBench(b, attrContent)
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), "test.dat")
			if !ok || size != math.MaxInt64 {
				b.Fatalf("unexpected: ok=%v size=%d", ok, size)
			}
		}
	})

	// Very long pattern names
	b.Run("long-pattern", func(b *testing.B) {
		longPattern := strings.Repeat("x", 1000) + ".dat"
		attrContent := fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text autotracksize=1000\n", longPattern)
		cleanup := setupAutoTrackBench(b, attrContent)
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), "test.dat")
			if ok {
				b.Fatal("expected no match for non-matching long pattern")
			}
		}
	})
}

// --- Sweep clean pass-through with many .gitattributes patterns ---

func BenchmarkSweepCleanPassThroughManyPatterns(b *testing.B) {
	patternCounts := []int{1, 10, 50, 100}

	for _, n := range patternCounts {
		b.Run(fmt.Sprintf("%d-patterns", n), func(b *testing.B) {
			var sb strings.Builder
			for i := 0; i < n; i++ {
				sb.WriteString(fmt.Sprintf("*.ext%04d filter=lfs autotracksize=%d\n", i, 100000))
			}
			sb.WriteString("*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=10000000\n")
			cleanup := setupAutoTrackBench(b, sb.String())
			defer cleanup()

			gf := lfs.NewGitFilter(cfg)
			data := make([]byte, 1000)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var to bytes.Buffer
					from := bytes.NewReader(data)
				ptr, err := clean(gf, &to, from, "test.dat", -1)
				if ptr != nil || err != nil {
					b.Fatalf("expected pass-through: ptr=%v err=%v", ptr, err)
				}
			}
		})
	}
}

// --- Baseline: clean without autotracksize in .gitattributes ---

func BenchmarkSweepCleanNoAutoTrack(b *testing.B) {
	cleanup := setupAutoTrackBench(b, "*.dat filter=lfs diff=lfs merge=lfs -text")
	defer cleanup()
	gf := lfs.NewGitFilter(cfg)

	sizes := []int{1 << 10, 100 << 10, 1 << 20}
	for _, sz := range sizes {
		label := fmt.Sprintf("%.0fKB", float64(sz)/(1<<10))
		if sz >= 1<<20 {
			label = fmt.Sprintf("%.0fMB", float64(sz)/(1<<20))
		}
		b.Run(label, func(b *testing.B) {
			data := make([]byte, sz)
			var to bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				to.Reset()
				from := bytes.NewReader(data)
				_, err := clean(gf, &to, from, "test.dat", -1)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// --- Already-pointer: content is a valid LFS pointer ---

func BenchmarkSweepCleanAlreadyPointer(b *testing.B) {
	thresholds := []int64{100, 10000000}
	pointerContent := "version https://git-lfs.github.com/spec/v1\noid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5\nsize 1024\n"

	for _, thresh := range thresholds {
		b.Run(fmt.Sprintf("thresh-%d", thresh), func(b *testing.B) {
			cleanup := setupAutoTrackBench(b, fmt.Sprintf("* autotracksize=%d", thresh))
			defer cleanup()

			gf := lfs.NewGitFilter(cfg)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var to bytes.Buffer
				from := strings.NewReader(pointerContent)
				ptr, err := clean(gf, &to, from, "test.dat", -1)
				if ptr == nil || err != nil {
					b.Fatalf("expected pointer: ptr=%v err=%v", ptr, err)
				}
			}
		})
	}
}

// --- Git-critical and autotrack-excluded paths ---

func BenchmarkSweepCleanExcludedPaths(b *testing.B) {
	cleanup := setupAutoTrackBench(b, "* autotracksize=10")
	defer cleanup()
	gf := lfs.NewGitFilter(cfg)

	data := []byte("this is some file content that is long enough to exceed the threshold")

	b.Run("git-critical", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var to bytes.Buffer
			from := bytes.NewReader(data)
			ptr, err := clean(gf, &to, from, ".gitattributes", -1)
			if ptr != nil || err != nil {
				b.Fatalf("expected pass-through: ptr=%v err=%v", ptr, err)
			}
		}
	})

	b.Run("autotrack-excluded", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var to bytes.Buffer
			from := bytes.NewReader(data)
			ptr, err := clean(gf, &to, from, "readme.md", -1)
			if ptr != nil || err != nil {
				b.Fatalf("expected pass-through: ptr=%v err=%v", ptr, err)
			}
		}
	})
}

// --- Smudge pass-through vs autotrack ---

func BenchmarkSweepSmudgePassThrough(b *testing.B) {
	b.Run("without-autotrack", func(b *testing.B) {
		cleanup := setupAutoTrackBench(b, "*.dat filter=lfs diff=lfs merge=lfs -text")
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			smudgePassThrough("test.dat", 100)
		}
	})

	b.Run("with-autotrack-under", func(b *testing.B) {
		cleanup := setupAutoTrackBench(b, "* autotracksize=100000")
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			smudgePassThrough("test.dat", 100)
		}
	})

	b.Run("with-autotrack-over", func(b *testing.B) {
		cleanup := setupAutoTrackBench(b, "* autotracksize=10")
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			smudgePassThrough("test.dat", 100000)
		}
	})

	b.Run("git-critical", func(b *testing.B) {
		cleanup := setupAutoTrackBench(b, "* autotracksize=10")
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			smudgePassThrough(".gitattributes", 100000)
		}
	})

	b.Run("excluded-file", func(b *testing.B) {
		cleanup := setupAutoTrackBench(b, "* autotracksize=100000")
		defer cleanup()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			smudgePassThrough("readme.md", 100000)
		}
	})
}

// --- IsAutoTrackExcluded: defaults vs user config ---

func BenchmarkSweepIsAutoTrackExcluded(b *testing.B) {
	b.Run("defaults-excluded", func(b *testing.B) {
		saved := cfg
		cfg = config.NewFrom(config.Values{})
		defer func() { cfg = saved }()

		excluded := []string{"readme.md", "notes.txt", "config.cfg", "setup.ini", ".gitlab-ci.yml"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, f := range excluded {
				if !isAutoTrackExcluded(f) {
					b.Fatalf("expected %s to be excluded", f)
				}
			}
		}
	})

	b.Run("defaults-not-excluded", func(b *testing.B) {
		saved := cfg
		cfg = config.NewFrom(config.Values{})
		defer func() { cfg = saved }()

		notExcluded := []string{"normal.dat", "script.py", "archive.tar.gz", "image.png"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, f := range notExcluded {
				if isAutoTrackExcluded(f) {
					b.Fatalf("expected %s to not be excluded", f)
				}
			}
		}
	})

	b.Run("user-config-excluded", func(b *testing.B) {
		saved := cfg
		cfg = config.NewFrom(config.Values{
			Git: map[string][]string{
				"lfs.autotrackexclude": {"*.pdf *.log *.tmp *.cache"},
			},
		})
		defer func() { cfg = saved }()

		excluded := []string{"document.pdf", "trace.log", "temp.tmp", "data.cache", "path/to/doc.pdf"}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, f := range excluded {
				if !isAutoTrackExcluded(f) {
					b.Fatalf("expected %s to be excluded", f)
				}
			}
		}
	})

	b.Run("user-config-replaces-defaults", func(b *testing.B) {
		saved := cfg
		cfg = config.NewFrom(config.Values{
			Git: map[string][]string{
				"lfs.autotrackexclude": {"*.pdf"},
			},
		})
		defer func() { cfg = saved }()

		if isAutoTrackExcluded("readme.md") {
			b.Fatal("readme.md should not be excluded when user overrides")
		}
	})
}

// --- File name length impact on clean autotrack ---

func BenchmarkSweepCleanFileNameLength(b *testing.B) {
	cleanup := setupAutoTrackBench(b, "* autotracksize=10000000")
	defer cleanup()
	gf := lfs.NewGitFilter(cfg)
	data := make([]byte, 1000)

	lengths := []int{10, 50, 100, 200, 500}
	for _, n := range lengths {
		b.Run(fmt.Sprintf("name-%d-chars", n), func(b *testing.B) {
			name := strings.Repeat("x", n-4) + ".dat"
			var to bytes.Buffer
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				to.Reset()
				from := bytes.NewReader(data)
				ptr, err := clean(gf, &to, from, name, -1)
				if ptr != nil || err != nil {
					b.Fatalf("expected pass-through")
				}
			}
		})
	}
}

// --- Cold cache: unique .gitattributes per iteration ---

func BenchmarkSweepGetAutoTrackSizeCold(b *testing.B) {
	dir := b.TempDir()
	err := os.MkdirAll(filepath.Join(dir, ".git", "lfs", "tmp"), 0755)
	if err != nil {
		b.Fatal(err)
	}

	saved := cfg
	defer func() { cfg = saved }()

	cfg = config.NewIn(dir, "")

	b.Run("cold-no-match", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			content := fmt.Sprintf("*.ext%08d filter=lfs diff=lfs merge=lfs -text autotracksize=100000\n", i)
			if err := os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(content), 0644); err != nil {
				b.Fatal(err)
			}
			size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), "target.dat")
			if ok || size != 0 {
				b.Fatalf("expected no match: ok=%v", ok)
			}
		}
	})

	b.Run("cold-match-first", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			content := fmt.Sprintf("*.dat filter=lfs diff=lfs merge=lfs -text autotracksize=100000\n*.ext%08d filter=lfs autotracksize=%d\n", i, i)
			if err := os.WriteFile(filepath.Join(dir, ".gitattributes"), []byte(content), 0644); err != nil {
				b.Fatal(err)
			}
			size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), "target.dat")
			if !ok || size != 100000 {
				b.Fatalf("unexpected: ok=%v size=%d", ok, size)
			}
		}
	})
}

func BenchmarkSweepGetAutoTrackSizePatternShape(b *testing.B) {
	patternSets := []struct {
		name        string
		pat         string
		file        string
		expectMatch bool
		expectSize  int64
	}{
		{"simple-glob", "*.dat", "test.dat", true, 50000},
		{"dir-glob", "**/*.dat", "a/b/test.dat", true, 50000},
		{"exact", ".gitattributes", ".gitattributes", true, 50000},
		{"mixed-case", "*.Data", "test.Data", true, 50000},
		{"no-match", "*.dat", "test.txt", false, 0},
	}
	for _, ps := range patternSets {
		for _, n := range []int{5, 50, 200} {
			b.Run(fmt.Sprintf("%s-%d-pats", ps.name, n), func(b *testing.B) {
				var sb strings.Builder
				for j := 0; j < n; j++ {
					fmt.Fprintf(&sb, "*.ext%04d filter=lfs diff=lfs merge=lfs -text autotracksize=100000\n", j)
				}
				fmt.Fprintf(&sb, "%s filter=lfs diff=lfs merge=lfs -text autotracksize=50000\n", ps.pat)
				cleanup := setupAutoTrackBench(b, sb.String())
				defer cleanup()

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					size, ok := gitattr.GetAutoTrackSize(cfg.LocalWorkingDir(), ps.file)
					if ps.expectMatch {
						if !ok || size != ps.expectSize {
							b.Fatalf("unexpected: ok=%v size=%d", ok, size)
						}
					} else {
						if ok {
							b.Fatalf("expected no match but got ok=true size=%d", size)
						}
					}
				}
			})
		}
	}
}
