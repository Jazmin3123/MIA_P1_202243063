package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MIA_P1_202243063/comandos"
)

const reportsDir = "/tmp/mia_p2_reportes"

type MkDiskRequest struct {
	Size int    `json:"size"`
	Unit string `json:"unit"`
	Fit  string `json:"fit"`
	Path string `json:"path"`
}

type ExecuteRequest struct {
	Command string `json:"command"`
}

type ExecuteResponse struct {
	OK     bool   `json:"ok"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

type MountsResponse struct {
	OK     bool        `json:"ok"`
	Mounts interface{} `json:"mounts"`
	Error  string      `json:"error,omitempty"`
}

type SessionResponse struct {
	OK      bool        `json:"ok"`
	Session interface{} `json:"session"`
	Error   string      `json:"error,omitempty"`
}

type FileSystemTreeResponse struct {
	OK    bool                     `json:"ok"`
	Path  string                   `json:"path,omitempty"`
	Items []comandos.DirectoryItem `json:"items,omitempty"`
	Error string                   `json:"error,omitempty"`
}

type FileSystemFileResponse struct {
	OK      bool   `json:"ok"`
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
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
	http.HandleFunc("/execute", handleExecute)
	http.HandleFunc("/mounts", handleMounts)
	http.HandleFunc("/session", handleSession)
	http.HandleFunc("/fs/tree", handleFileSystemTree)
	http.HandleFunc("/fs/file", handleFileSystemFile)
	http.HandleFunc("/reports/file", handleReportFile)
	return http.ListenAndServe(":8080", nil)
}

func handleMkDisk(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

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

func handleExecute(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, ExecuteResponse{
			OK:     false,
			Output: "",
			Error:  "metodo no permitido, use POST",
		})
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, ExecuteResponse{
			OK:     false,
			Output: "",
			Error:  "json invalido",
		})
		return
	}

	if err := comandos.ExecuteLine(req.Command); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, ExecuteResponse{
			OK:     false,
			Output: "",
			Error:  err.Error(),
		})
		return
	}

	writeJSON(w, ExecuteResponse{
		OK:     true,
		Output: "comando ejecutado",
		Error:  "",
	})
}

func handleMounts(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, MountsResponse{
			OK:     false,
			Mounts: []interface{}{},
			Error:  "metodo no permitido, use GET",
		})
		return
	}

	writeJSON(w, MountsResponse{
		OK:     true,
		Mounts: comandos.ListMountedPartitions(),
	})
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, SessionResponse{
			OK:      false,
			Session: map[string]bool{"logged": false},
			Error:   "metodo no permitido, use GET",
		})
		return
	}

	writeJSON(w, SessionResponse{
		OK:      true,
		Session: comandos.GetCurrentSession(),
	})
}

func handleFileSystemTree(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, FileSystemTreeResponse{
			OK:    false,
			Error: "metodo no permitido, use GET",
		})
		return
	}

	id := r.URL.Query().Get("id")
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	items, err := comandos.ListDirectory(id, path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, FileSystemTreeResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, FileSystemTreeResponse{
		OK:    true,
		Path:  path,
		Items: items,
	})
}

func handleFileSystemFile(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, FileSystemFileResponse{
			OK:    false,
			Error: "metodo no permitido, use GET",
		})
		return
	}

	id := r.URL.Query().Get("id")
	path := r.URL.Query().Get("path")
	content, err := comandos.ReadFileContent(id, path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, FileSystemFileResponse{
			OK:    false,
			Error: err.Error(),
		})
		return
	}

	writeJSON(w, FileSystemFileResponse{
		OK:      true,
		Path:    path,
		Content: content,
	})
}

func handleReportFile(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: "metodo no permitido, use GET",
		})
		return
	}

	name := filepath.Base(r.URL.Query().Get("name"))
	if name == "." || name == "/" || name == "" || strings.Contains(name, "..") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: "nombre de reporte invalido",
		})
		return
	}

	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: "no se pudo preparar la carpeta de reportes",
		})
		return
	}

	reportPath := filepath.Join(reportsDir, name)
	extension := strings.ToLower(filepath.Ext(name))
	switch extension {
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".txt":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, ApiResponse{
			OK:    false,
			Error: "tipo de reporte no permitido",
		})
		return
	}

	http.ServeFile(w, r, reportPath)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, response any) {
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
