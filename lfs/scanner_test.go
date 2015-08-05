package lfs

import (
	"strings"
	"testing"

	"github.com/github/git-lfs/vendor/_nuts/github.com/technoweenie/assert"
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
+oid sha256:ebff26d6b557b1416a6fded097fd9b9102e2d8195532c377ac365c736c87d4bc
+size 127142413
`
)

func TestParseLogOutputToPointersAdditions(t *testing.T) {

	// test + diff, no filtering
	r := strings.NewReader(pointerParseLogOutput)
	pchan := make(chan *WrappedPointer, chanBufSize)
	go parseLogOutputToPointers(r, LogDiffAdditions, nil, nil, pchan)
	pointers := make([]*WrappedPointer, 0, 5)
	for p := range pchan {
		pointers = append(pointers, p)
	}
	assert.Equal(t, 4, len(pointers))

	// modification, + side
	assert.Equal(t, "radial_1.png", pointers[0].Name)
	assert.Equal(t, "3301b3da173d231f0f6b1f9bf075e573758cd79b3cfeff7623a953d708d6688b", pointers[0].Oid)
	assert.Equal(t, int64(3152388), pointers[0].Size)
	// addition, + side
	assert.Equal(t, "1D_Noise.png", pointers[1].Name)
	assert.Equal(t, "f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6", pointers[1].Oid)
	assert.Equal(t, int64(1289), pointers[1].Size)
	// addition, + side
	assert.Equal(t, "waveNM.png", pointers[2].Name)
	assert.Equal(t, "fe2c2f236b97bba4585d9909a227a8fa64897d9bbe297fa272f714302d86c908", pointers[2].Oid)
	assert.Equal(t, int64(125873), pointers[2].Size)
	// addition, + side
	assert.Equal(t, "hobbit_5armies_2.mov", pointers[3].Name)
	assert.Equal(t, "ebff26d6b557b1416a6fded097fd9b9102e2d8195532c377ac365c736c87d4bc", pointers[3].Oid)
	assert.Equal(t, int64(127142413), pointers[3].Size)

	// test filtered, include
	r = strings.NewReader(pointerParseLogOutput)
	pointers = pointers[:0]
	pchan = make(chan *WrappedPointer, chanBufSize)
	go parseLogOutputToPointers(r, LogDiffAdditions, []string{"wave*"}, nil, pchan)
	for p := range pchan {
		pointers = append(pointers, p)
	}
	assert.Equal(t, 1, len(pointers))
	assert.Equal(t, "waveNM.png", pointers[0].Name)
	assert.Equal(t, "fe2c2f236b97bba4585d9909a227a8fa64897d9bbe297fa272f714302d86c908", pointers[0].Oid)
	assert.Equal(t, int64(125873), pointers[0].Size)

	// test filtered, exclude
	r = strings.NewReader(pointerParseLogOutput)
	pointers = pointers[:0]
	pchan = make(chan *WrappedPointer, chanBufSize)
	go parseLogOutputToPointers(r, LogDiffAdditions, nil, []string{"wave*"}, pchan)
	for p := range pchan {
		pointers = append(pointers, p)
	}
	assert.Equal(t, 3, len(pointers))
	assert.Equal(t, "radial_1.png", pointers[0].Name)
	assert.Equal(t, "3301b3da173d231f0f6b1f9bf075e573758cd79b3cfeff7623a953d708d6688b", pointers[0].Oid)
	assert.Equal(t, int64(3152388), pointers[0].Size)
	assert.Equal(t, "1D_Noise.png", pointers[1].Name)
	assert.Equal(t, "f5d84da40ab1f6aa28df2b2bf1ade2cdcd4397133f903c12b4106641b10e1ed6", pointers[1].Oid)
	assert.Equal(t, int64(1289), pointers[1].Size)
	assert.Equal(t, "hobbit_5armies_2.mov", pointers[2].Name)
	assert.Equal(t, "ebff26d6b557b1416a6fded097fd9b9102e2d8195532c377ac365c736c87d4bc", pointers[2].Oid)
	assert.Equal(t, int64(127142413), pointers[2].Size)

}

func TestParseLogOutputToPointersDeletion(t *testing.T) {

	// test - diff, no filtering
	r := strings.NewReader(pointerParseLogOutput)
	pchan := make(chan *WrappedPointer, chanBufSize)
	go parseLogOutputToPointers(r, LogDiffDeletions, nil, nil, pchan)
	pointers := make([]*WrappedPointer, 0, 5)
	for p := range pchan {
		pointers = append(pointers, p)
	}
	assert.Equal(t, 3, len(pointers))

	// deletion, - side
	assert.Equal(t, "smoke_1.png", pointers[0].Name)
	assert.Equal(t, "8eb65d66303acc60062f44b44ef1f7360d7189db8acf3d066e59e2528f39514e", pointers[0].Oid)
	assert.Equal(t, int64(35022), pointers[0].Size)
	// addition, + side
	assert.Equal(t, "flare_1.png", pointers[1].Name)
	assert.Equal(t, "ea61c67cc5e8b3504d46de77212364045f31d9a023ad4448a1ace2a2fb4eed28", pointers[1].Oid)
	assert.Equal(t, int64(72982), pointers[1].Size)
	// modification, - side
	assert.Equal(t, "radial_1.png", pointers[2].Name)
	assert.Equal(t, "334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd", pointers[2].Oid)
	assert.Equal(t, int64(16849), pointers[2].Size)

	// test filtered, include
	r = strings.NewReader(pointerParseLogOutput)
	pointers = pointers[:0]
	pchan = make(chan *WrappedPointer, chanBufSize)
	go parseLogOutputToPointers(r, LogDiffDeletions, []string{"flare*"}, nil, pchan)
	for p := range pchan {
		pointers = append(pointers, p)
	}
	assert.Equal(t, 1, len(pointers))
	assert.Equal(t, "flare_1.png", pointers[0].Name)
	assert.Equal(t, "ea61c67cc5e8b3504d46de77212364045f31d9a023ad4448a1ace2a2fb4eed28", pointers[0].Oid)
	assert.Equal(t, int64(72982), pointers[0].Size)

	// test filtered, exclude
	r = strings.NewReader(pointerParseLogOutput)
	pointers = pointers[:0]
	pchan = make(chan *WrappedPointer, chanBufSize)
	go parseLogOutputToPointers(r, LogDiffDeletions, nil, []string{"flare*"}, pchan)
	for p := range pchan {
		pointers = append(pointers, p)
	}
	assert.Equal(t, 2, len(pointers))
	assert.Equal(t, "smoke_1.png", pointers[0].Name)
	assert.Equal(t, "8eb65d66303acc60062f44b44ef1f7360d7189db8acf3d066e59e2528f39514e", pointers[0].Oid)
	assert.Equal(t, int64(35022), pointers[0].Size)
	assert.Equal(t, "radial_1.png", pointers[1].Name)
	assert.Equal(t, "334c8a0a520cf9f58189dba5a9a26c7bff2769b4a3cc199650c00618bde5b9dd", pointers[1].Oid)
	assert.Equal(t, int64(16849), pointers[1].Size)

}
