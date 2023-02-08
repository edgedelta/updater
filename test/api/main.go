package main

import (
	"encoding/json"
	"net/http"

	"github.com/edgedelta/updater/core"

	"github.com/gorilla/mux"
)

var (
	responseData = core.LatestTagResponse{
		Tag:   "v0.1.47",
		Image: "my-image",
		URL:   "gcr.io/my-org/image:v0.1.47",
	}
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(responseData)
		if err != nil {
			panic(err)
		}
		w.Write(b)
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
