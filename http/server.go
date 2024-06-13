package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pixels-broadcaster/data"
)

func StartServer(broadcaster *data.Broadcaster, port string) error {
	http.HandleFunc("/data", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		broadcaster.Broadcast(payload)
		fmt.Fprintf(w, "Data received and broadcasted")
	})

	log.Printf("HTTP server running on %s", port)
	return http.ListenAndServe(port, nil)
}
