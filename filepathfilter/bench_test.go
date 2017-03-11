package filepathfilter_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/git-lfs/git-lfs/tools"
)

func BenchmarkFilterIncludeWildcardOnly(b *testing.B) {
	files := benchmarkTree(b)
	filter := filepathfilter.New([]string{"*.go"}, nil)
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			filter.Allows(f)
		}
	}
}

func BenchmarkFilterIncludeDoubleAsterisk(b *testing.B) {
	files := benchmarkTree(b)
	filter := filepathfilter.New([]string{"**/README.md"}, nil)
	for i := 0; i < b.N; i++ {
		for _, f := range files {
			filter.Allows(f)
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
