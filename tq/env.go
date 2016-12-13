package tq

// copy of config.Environment. Removes forced coupling with config pkg.
type Environment interface {
	Get(key string) (val string, ok bool)
	Bool(key string, def bool) (val bool)
	Int(key string, def int) (val int)
	All() map[string]string
}
