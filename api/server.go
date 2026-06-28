package api

import (
	"encoding/json"
	"net/http"

	"MIA_P1_202243063/comandos"
)

type MkDiskRequest struct {
	Size int    `json:"size"`
	Unit string `json:"unit"`
	Fit  string `json:"fit"`
	Path string `json:"path"`
}

type ApiResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
	Path      string `json:"path,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

func StartAPI() error {
	http.HandleFunc("/mkdisk", handleMkDisk)
	return http.ListenAndServe(":8080", nil)
}

func handleMkDisk(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: "metodo no permitido, use POST",
		})
		return
	}

	var req MkDiskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: "json invalido",
		})
		return
	}

	sizeBytes, err := comandos.CrearDisco(req.Size, req.Unit, req.Fit, req.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, ApiResponse{
		OK:        true,
		Message:   "disco creado correctamente",
		Path:      req.Path,
		SizeBytes: sizeBytes,
	})
}

func writeJSON(w http.ResponseWriter, response ApiResponse) {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
