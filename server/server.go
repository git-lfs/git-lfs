package gitmediaserver

import (
	".."
	"../filters"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

func PutObject(w http.ResponseWriter, r *http.Request) {
	cleaned, err := gitmediafilters.Clean(r.Body)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
	}
	defer cleaned.File.Close()

	vars := mux.Vars(r)
	sha := vars["oid"]
	if cleaned.Sha != sha {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprint(w, "CONFLICT")
		return
	}

	localpath := gitmedia.LocalMediaPath(sha)
	if stat, _ := os.Stat(localpath); stat != nil {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}

	if err = os.Rename(cleaned.File.Name(), localpath); err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, "OK")
}

func GetObject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sha := vars["oid"]
	localpath := gitmedia.LocalMediaPath(sha)
	file, err := os.Open(localpath)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprint(w, "nope nope nope")
	}

	io.Copy(w, file)
}
