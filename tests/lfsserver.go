package tests

import (
	"net/http"
)

func (run *runner) lfsHandler(w http.ResponseWriter, r *http.Request) {
	run.Fatalf("wat")
}
