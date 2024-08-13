package main

import (
	"encoding/json"
	"flag"
	"io/fs"
	"log/slog"
	"net/http"
	"proxy-manager/frontend"
	"proxy-manager/pkg/caddy"
	"time"
)

var Address = flag.String("address", ":8080", "Address to listen on")
var Key = flag.String("key", "secret", "API key")

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	flag.Parse()

	client := caddy.NewClient("srv0", "http://localhost:2019")
	err := client.Init()
	if err != nil {
		slog.Error("Failed to initialize Caddy client", "error", err)
		return
	}

	refreshTicker := time.NewTicker(5 * time.Second)
	go func() {
		for range refreshTicker.C {
			err := client.Refresh()
			if err != nil {
				slog.Error("Failed to refresh Caddy configuration", "error", err)
			}
		}
	}()
	defer refreshTicker.Stop()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /proxies", func(w http.ResponseWriter, r *http.Request) {
		proxies := client.ListProxies()
		RespondJSON(w, http.StatusOK, proxies)
	})
	mux.HandleFunc("POST /proxies", func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Key")
		if key != *Key {
			RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid API key"})
			return
		}

		var proxy caddy.Proxy
		err := json.NewDecoder(r.Body).Decode(&proxy)
		if err != nil {
			RespondJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
			return
		}

		err = client.AddRoute(proxy.ToRoute())
		if err != nil {
			RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		RespondJSON(w, http.StatusCreated, nil)
	})
	mux.HandleFunc("DELETE /proxies/{id}", func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Key")
		if key != *Key {
			RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid API key"})
			return
		}

		id := r.PathValue("id")
		slog.Info("Deleting proxy", "id", id)
		err := client.DeleteObject("id/" + id)
		if err != nil {
			RespondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		client.Refresh()

		RespondJSON(w, http.StatusOK, nil)
	})

	// add frontend fs
	subFS, err := fs.Sub(frontend.Assets, "dist")
	if err != nil {
		slog.Error("Failed to create sub FS", "error", err)
		return
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	slog.Info("Listening", "address", *Address)
	err = http.ListenAndServe(*Address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Key")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	}))
	if err != nil {
		slog.Error("Failed to start server", "error", err)
	}
}
