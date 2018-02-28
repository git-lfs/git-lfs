// +build !linux

package wildmatch

func init() {
	SystemCase = func(w *Wildmatch) {}
}
