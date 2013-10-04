package main

import (
	".."
	"../server"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/objects/{oid}", gitmediaserver.GetObject).Methods("GET")
	r.HandleFunc("/objects/{oid}", gitmediaserver.PutObject).Methods("PUT")
	http.Handle("/", r)
	fmt.Println("Serving out of", gitmedia.LocalMediaDir)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
