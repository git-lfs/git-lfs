package lfs

import (
	"strings"
	"testing"

	"github.com/git-lfs/git-lfs/filepathfilter"
	"github.com/stretchr/testify/assert"
)

var (
	pointerParseLogOutput = `lfs-commit-sha: 637908bf28b38ab238e1b5e6a5bfbfb2e513a0df 07d571b413957508679042e45508af5945b3f1e5

diff --git a/smoke_1.png b/smoke_1.png
deleted file mode 100644
index 2fe5451..0000000
--- a/smoke_1.png
+++ /dev/null
@@ -1,3 +0,0 @@
-version https://git-lfs.github.com/spec/v1
-oid sha256:8eb65d66303acc60062f44b44ef1f7360d7189db8acf3d066e59e2528f39514e
-size 35022
lfs-commit-sha: 07d571b413957508679042e45508af5945b3f1e5 8e5bd456b754f7d61c7157e82edc5ed124be4da6

diff --git a/flare_1.png b/flare_1.png
deleted file mode 100644
index 1cfc5a1..0000000
--- a/flare_1.png
+++ /dev/null
@@ -1,3 +0,0 @@
-version https://git-lfs.github.com/spec/v1
-ext-0-foo sha256:36485434f4f8a55150282ae7c78619a89de52721c00f48159f2562463df9c043
-ext-1-bar sha256:382a2a13e705bbd8de7e2e13857c26551db17121ac57edca5dec9b5bd753e9c8
-ext-2-ray sha256:423ee9e5988fb4670bf815990e9307c3b23296210c31581dec4d4ae89dabae46
-oid sha256:ea61c67cc5e8b3504d46de77212364045f31d9a023ad4448a1ace2a2fb4eed28
-size 72982
diff --git a/radial_1.png b/radial_1.png
index 9daa2e5..c648385 100644
--- a/radial_1.png
+++ b/radial_1.png
@@ -1,3 +1,3 @@
 version https://git-lfs.github.com/spec/v1
-oid sha256:334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd
-size 16849
+oid sha256:3301b3da173d231f0f6b1f9bf075e573758cd79b3cfeff7623a953d708d6688b
+size 3152388
diff --git a/radial_2.png b/radial_2.png
index 9daa2e5..c648385 100644
--- a/radial_2.png
+++ b/radial_2.png
@@ -1,3 +1,3 @@
 version https://git-lfs.github.com/spec/v1
-ext-0-foo sha256:36485434f4f8a55150282ae7c78619a89de52721c00f48159f2562463df9c043
-ext-1-bar sha256:382a2a13e705bbd8de7e2e13857c26551db17121ac57edca5dec9b5bd753e9c8
-ext-2-ray sha256:423ee9e5988fb4670bf815990e9307c3b23296210c31581dec4d4ae89dabae46
-oid sha256:334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd
-size 16849
+ext-0-foo sha256:95d8260e8365a9dfd39842bdeee9b20e0e3fe3daf9bb4a8c0a1acb31008ed7b4
+ext-1-bar sha256:674bf4995720a43e03e174bcc1132ca95de6a8e4155fe3b2c482dceb42cbc0a5
+ext-2-ray sha256:0d323c95ae4b0a9c195ddc437c470678bddd2ee0906fb2f7b8166cd2474e22d9
+oid sha256:4b666195c133d8d0541ad0bc0e77399b9dc81861577a98314ac1ff1e9877893a
+size 3152388
lfs-commit-sha: 60fde3d23553e10a55e2a32ed18c20f65edd91e7 e2eaf1c10b57da7b98eb5d722ec5912ddeb53ea1

diff --git a/1D_Noise.png b/1D_Noise.png
new file mode 100644
index 0000000..2622b4a
--- /dev/null
+++ b/1D_Noise.png
@@ -0,0 +1,3 @@
+version https://git-lfs.github.com/spec/v1
+oid sha256:f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6
+size 1289
diff --git a/waveNM.png b/waveNM.png
new file mode 100644
index 0000000..8519883
--- /dev/null
+++ b/waveNM.png
@@ -0,0 +1,3 @@
+version https://git-lfs.github.com/spec/v1
+oid sha256:fe2c2f236b97bba4585d9909a227a8fa64897d9bbe297fa272f714302d86c908
+size 125873
lfs-commit-sha: 64b3372e108daaa593412d5e1d9df8169a9547ea e99c9cac7ff3f3cf1b2e670a64a5a381c44ffceb

diff --git a/hobbit_5armies_2.mov b/hobbit_5armies_2.mov
new file mode 100644
index 0000000..92a88f8
--- /dev/null
+++ b/hobbit_5armies_2.mov
@@ -0,0 +1,3 @@
+version https://git-lfs.github.com/spec/v1
+ext-0-foo sha256:b37197ac149950d057521bcb7e00806f0528e19352bd72767165bc390d4f055e
+ext-1-bar sha256:c71772e5ea8e8c6f053f0f1dc89f8c01243975b1a040acbcf732fe2dbc0bcb61
+oid sha256:ebff26d6b557b1416a6fded097fd9b9102e2d8195532c377ac365c736c87d4bc
+size 127142413
`
)

func TestLogScannerAdditionsNoFiltering(t *testing.T) {
	r := strings.NewReader(pointerParseLogOutput)
	scanner := newLogScanner(LogDiffAdditions, r)

	// modification, + side
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_1.png", p.Name)
		assert.Equal(t, "3301b3da173d231f0f6b1f9bf075e573758cd79b3cfeff7623a953d708d6688b", p.Oid)
		assert.Equal(t, int64(3152388), p.Size)
	}

	// modification, + side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_2.png", p.Name)
		assert.Equal(t, "4b666195c133d8d0541ad0bc0e77399b9dc81861577a98314ac1ff1e9877893a", p.Oid)
		assert.Equal(t, int64(3152388), p.Size)
	}

	// addition, + side
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "1D_Noise.png", p.Name)
		assert.Equal(t, "f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6", p.Oid)
		assert.Equal(t, int64(1289), p.Size)
	}

	// addition, + side
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "waveNM.png", p.Name)
		assert.Equal(t, "fe2c2f236b97bba4585d9909a227a8fa64897d9bbe297fa272f714302d86c908", p.Oid)
		assert.Equal(t, int64(125873), p.Size)
	}

	// addition, + side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "hobbit_5armies_2.mov", p.Name)
		assert.Equal(t, "ebff26d6b557b1416a6fded097fd9b9102e2d8195532c377ac365c736c87d4bc", p.Oid)
		assert.Equal(t, int64(127142413), p.Size)
	}

	assertScannerDone(t, scanner)
	assert.Nil(t, scanner.Pointer())
}

func TestLogScannerAdditionsFilterInclude(t *testing.T) {
	r := strings.NewReader(pointerParseLogOutput)
	scanner := newLogScanner(LogDiffAdditions, r)
	scanner.Filter = filepathfilter.New([]string{"wave*"}, nil)

	// addition, + side
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "waveNM.png", p.Name)
		assert.Equal(t, "fe2c2f236b97bba4585d9909a227a8fa64897d9bbe297fa272f714302d86c908", p.Oid)
		assert.Equal(t, int64(125873), p.Size)
	}
	assertScannerDone(t, scanner)
	assert.Nil(t, scanner.Pointer())
}

func TestLogScannerAdditionsFilterExclude(t *testing.T) {
	r := strings.NewReader(pointerParseLogOutput)
	scanner := newLogScanner(LogDiffAdditions, r)
	scanner.Filter = filepathfilter.New(nil, []string{"wave*"})

	// modification, + side
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_1.png", p.Name)
		assert.Equal(t, "3301b3da173d231f0f6b1f9bf075e573758cd79b3cfeff7623a953d708d6688b", p.Oid)
		assert.Equal(t, int64(3152388), p.Size)
	}

	// modification, + side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_2.png", p.Name)
		assert.Equal(t, "4b666195c133d8d0541ad0bc0e77399b9dc81861577a98314ac1ff1e9877893a", p.Oid)
		assert.Equal(t, int64(3152388), p.Size)
	}

	// addition, + side
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "1D_Noise.png", p.Name)
		assert.Equal(t, "f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6", p.Oid)
		assert.Equal(t, int64(1289), p.Size)
	}

	// addition, + side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "hobbit_5armies_2.mov", p.Name)
		assert.Equal(t, "ebff26d6b557b1416a6fded097fd9b9102e2d8195532c377ac365c736c87d4bc", p.Oid)
		assert.Equal(t, int64(127142413), p.Size)
	}

	assertScannerDone(t, scanner)
	assert.Nil(t, scanner.Pointer())
}

func TestLogScannerDeletionsNoFiltering(t *testing.T) {
	r := strings.NewReader(pointerParseLogOutput)
	scanner := newLogScanner(LogDiffDeletions, r)

	// deletion, - side
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "smoke_1.png", p.Name)
		assert.Equal(t, "8eb65d66303acc60062f44b44ef1f7360d7189db8acf3d066e59e2528f39514e", p.Oid)
		assert.Equal(t, int64(35022), p.Size)
	}

	// deletion, - side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "flare_1.png", p.Name)
		assert.Equal(t, "ea61c67cc5e8b3504d46de77212364045f31d9a023ad4448a1ace2a2fb4eed28", p.Oid)
		assert.Equal(t, int64(72982), p.Size)
	}

	// modification, - side
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_1.png", p.Name)
		assert.Equal(t, "334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd", p.Oid)
		assert.Equal(t, int64(16849), p.Size)
	}

	// modification, - side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_2.png", p.Name)
		assert.Equal(t, "334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd", p.Oid)
		assert.Equal(t, int64(16849), p.Size)
	}

	assertScannerDone(t, scanner)
}

func TestLogScannerDeletionsFilterInclude(t *testing.T) {
	r := strings.NewReader(pointerParseLogOutput)
	scanner := newLogScanner(LogDiffDeletions, r)
	scanner.Filter = filepathfilter.New([]string{"flare*"}, nil)

	// deletion, - side with extensions
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "flare_1.png", p.Name)
		assert.Equal(t, "ea61c67cc5e8b3504d46de77212364045f31d9a023ad4448a1ace2a2fb4eed28", p.Oid)
		assert.Equal(t, int64(72982), p.Size)
	}

	assertScannerDone(t, scanner)
}

func TestLogScannerDeletionsFilterExclude(t *testing.T) {
	r := strings.NewReader(pointerParseLogOutput)
	scanner := newLogScanner(LogDiffDeletions, r)
	scanner.Filter = filepathfilter.New(nil, []string{"flare*"})

	// deletion, - side
	assertNextScan(t, scanner)
	p := scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "smoke_1.png", p.Name)
		assert.Equal(t, "8eb65d66303acc60062f44b44ef1f7360d7189db8acf3d066e59e2528f39514e", p.Oid)
		assert.Equal(t, int64(35022), p.Size)
	}

	// modification, - side
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_1.png", p.Name)
		assert.Equal(t, "334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd", p.Oid)
		assert.Equal(t, int64(16849), p.Size)
	}

	// modification, - side with extensions
	assertNextScan(t, scanner)
	p = scanner.Pointer()
	if assert.NotNil(t, p) {
		assert.Equal(t, "radial_2.png", p.Name)
		assert.Equal(t, "334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd", p.Oid)
		assert.Equal(t, int64(16849), p.Size)
	}

	assertScannerDone(t, scanner)
}

func TestLsTreeParser(t *testing.T) {
	stdout := "100644 blob d899f6551a51cf19763c5955c7a06a2726f018e9      42	.gitattributes\000100644 blob 4d343e022e11a8618db494dc3c501e80c7e18197     126	PB SCN 16 Odhrán.wav"
	scanner := newLsTreeScanner(strings.NewReader(stdout))

	assertNextTreeBlob(t, scanner, "d899f6551a51cf19763c5955c7a06a2726f018e9", ".gitattributes")
	assertNextTreeBlob(t, scanner, "4d343e022e11a8618db494dc3c501e80c7e18197", "PB SCN 16 Odhrán.wav")
	assertScannerDone(t, scanner)
}

func assertNextTreeBlob(t *testing.T, scanner *lsTreeScanner, oid, filename string) {
	assertNextScan(t, scanner)
	b := scanner.TreeBlob()
	assert.NotNil(t, b)
	assert.Equal(t, oid, b.Sha1)
	assert.Equal(t, filename, b.Filename)
}

func BenchmarkLsTreeParser(b *testing.B) {
	stdout := "100644 blob d899f6551a51cf19763c5955c7a06a2726f018e9      42	.gitattributes\000100644 blob 4d343e022e11a8618db494dc3c501e80c7e18197     126	PB SCN 16 Odhrán.wav"

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		scanner := newLsTreeScanner(strings.NewReader(stdout))
		for scanner.Scan() {
		}
	}
}
