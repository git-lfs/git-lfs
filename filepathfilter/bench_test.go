package filepathfilter_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/tools"
)

func BenchmarkFilterSimplePath(b *testing.B) {
	files := benchmarkTree(b)
	filter := filepathfilter.New([]string{"lfs"}, nil)
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			filter.Allows(f)
		}
	}
}

func BenchmarkPatternSimplePath(b *testing.B) {
	files := benchmarkTree(b)
	pattern := filepathfilter.NewPattern("lfs")
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			pattern.Match(f)
		}
	}
}

func BenchmarkFilterSimpleExtension(b *testing.B) {
	files := benchmarkTree(b)
	filter := filepathfilter.New([]string{"*.go"}, nil)
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			filter.Allows(f)
		}
	}
}

func BenchmarkPatternSimpleExtension(b *testing.B) {
	files := benchmarkTree(b)
	pattern := filepathfilter.NewPattern("*.go")
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			pattern.Match(f)
		}
	}
}

func BenchmarkFilterComplexExtension(b *testing.B) {
	files := benchmarkTree(b)
	filter := filepathfilter.New([]string{"*.travis.yml"}, nil)
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			filter.Allows(f)
		}
	}
}

func BenchmarkPatternComplexExtension(b *testing.B) {
	files := benchmarkTree(b)
	pattern := filepathfilter.NewPattern("*.travis.yml")
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			pattern.Match(f)
		}
	}
}

func BenchmarkFilterDoubleAsterisk(b *testing.B) {
	files := benchmarkTree(b)
	filter := filepathfilter.New([]string{"**/README.md"}, nil)
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			filter.Allows(f)
		}
	}
}

func BenchmarkPatternDoubleAsterisk(b *testing.B) {
	files := benchmarkTree(b)
	pattern := filepathfilter.NewPattern("**/README.md")
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			pattern.Match(f)
		}
	}
}

var (
	benchmarkFiles []string
	benchmarkMu    sync.Mutex
)

func benchmarkTree(b *testing.B) []string {
	benchmarkMu.Lock()
	defer benchmarkMu.Unlock()

	if benchmarkFiles != nil {
		return benchmarkFiles
	}

	wd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}

	hasErrors := false
	tools.FastWalkGitRepo(filepath.Dir(wd), func(parent string, info os.FileInfo, err error) {
		if err != nil {
			hasErrors = true
			b.Error(err)
			return
		}
		benchmarkFiles = append(benchmarkFiles, filepath.Join(parent, info.Name()))
	})

	if hasErrors {
		b.Fatal("has errors :(")
	}

	return benchmarkFiles
}
