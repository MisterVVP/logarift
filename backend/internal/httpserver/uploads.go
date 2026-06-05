package httpserver

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxUploadBytes int64 = 8 << 20 // 8 MiB

var allowedUploadTypes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type uploadResponse struct {
	ID          string    `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	URLPath     string    `json:"url_path"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *Server) uploadAttachment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"code": "invalid_upload", "message": "Expected multipart form upload with one image file under field 'file'."}})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"code": "missing_file", "message": "Missing multipart form file field 'file'."}})
		return
	}
	defer file.Close()

	peek := make([]byte, 512)
	n, _ := io.ReadFull(file, peek)
	peek = peek[:n]
	contentType := http.DetectContentType(peek)
	ext, ok := allowedUploadTypes[contentType]
	if !ok {
		if declared := header.Header.Get("Content-Type"); declared != "" {
			mediaType, _, _ := mime.ParseMediaType(declared)
			if allowedExt, exists := allowedUploadTypes[mediaType]; exists {
				contentType = mediaType
				ext = allowedExt
				ok = true
			}
		}
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": map[string]any{"code": "unsupported_file_type", "message": "Only PNG, JPEG, WebP, and GIF images are accepted."}})
			return
		}
	}

	id, err := randomID()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "internal_error", "message": "Could not create upload identifier."}})
		return
	}
	filename := sanitizeFilename(header.Filename)
	storedName := id + ext
	if filename != "" {
		storedName = id + "-" + strings.TrimSuffix(filename, filepath.Ext(filename)) + ext
	}
	if err := os.MkdirAll(s.cfg.UploadDir, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "upload_dir_error", "message": "Could not create local upload directory."}})
		return
	}
	path := filepath.Join(s.cfg.UploadDir, storedName)
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "upload_write_error", "message": "Could not store uploaded file locally."}})
		return
	}
	defer out.Close()
	if len(peek) > 0 {
		if _, err := out.Write(peek); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "upload_write_error", "message": "Could not store uploaded file locally."}})
			return
		}
	}
	if _, err := io.Copy(out, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]any{"code": "upload_write_error", "message": "Could not store uploaded file locally."}})
		return
	}

	createdAt := time.Now().UTC()
	writeJSON(w, http.StatusCreated, uploadResponse{
		ID:          id,
		Filename:    storedName,
		ContentType: contentType,
		URLPath:     "/uploads/" + storedName,
		CreatedAt:   createdAt,
	})
}

func (s *Server) serveUploadedAttachment(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.PathValue("filename"))
	if filename == "." || filename == string(filepath.Separator) || strings.TrimSpace(filename) == "" {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(s.cfg.UploadDir, filename)
	http.ServeFile(w, r, path)
}

func randomID() (string, error) {
	var buf [12]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ToLower(name)
	var b strings.Builder
	for _, r := range name {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if r == ' ' || r == '.' {
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 80 {
		out = out[:80]
	}
	return fmt.Sprintf("%s", out)
}
