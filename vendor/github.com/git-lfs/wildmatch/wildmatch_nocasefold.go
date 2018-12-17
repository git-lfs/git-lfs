// +build !windows,!darwin

package wildmatch

func init() {
	SystemCase = func(w *Wildmatch) {}
}
