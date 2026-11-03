package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"

	"github.com/octarudin/usb-thermal-printer/internal/printer"
	"github.com/octarudin/usb-thermal-printer/internal/receipt"
)

//go:embed static/*
var staticFiles embed.FS

type apiResponse struct {
	OK           bool   `json:"ok"`
	Message      string `json:"message"`
	Bytes        int    `json:"bytes,omitempty"`
	Endpoint     int    `json:"endpoint,omitempty"`
	TransferType string `json:"transferType,omitempty"`
}

func main() {
	address := flag.String("addr", "127.0.0.1:8787", "alamat GUI lokal")
	noBrowser := flag.Bool("no-browser", false, "jangan membuka browser secara otomatis")
	flag.Parse()

	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(staticRoot)))
	mux.HandleFunc("GET /api/info", infoHandler)
	mux.HandleFunc("POST /api/render", renderHandler)
	mux.HandleFunc("POST /api/print", printHandler)

	listener, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("gagal membuka %s: %v", *address, err)
	}

	url := "http://" + listener.Addr().String()
	server := &http.Server{
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Receipt Designer berjalan di %s", url)
	if !*noBrowser {
		go func() {
			time.Sleep(300 * time.Millisecond)
			if err := openBrowser(url); err != nil {
				log.Printf("browser tidak dapat dibuka otomatis: %v", err)
			}
		}()
	}

	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func infoHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":           "Taffware Receipt Designer",
		"paperWidthMM":   58,
		"printWidthDots": receipt.PaperWidthDots,
		"vendorID":       fmt.Sprintf("%04x", uint16(printer.DefaultVendorID)),
		"productID":      fmt.Sprintf("%04x", uint16(printer.DefaultProductID)),
	})
}

func renderHandler(w http.ResponseWriter, r *http.Request) {
	template, ok := decodeTemplate(w, r)
	if !ok {
		return
	}

	data, err := receipt.Build(template)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		OK:      true,
		Message: "Template valid dan siap dicetak.",
		Bytes:   len(data),
	})
}

func printHandler(w http.ResponseWriter, r *http.Request) {
	template, ok := decodeTemplate(w, r)
	if !ok {
		return
	}

	data, err := receipt.Build(template)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
		return
	}

	result, err := printer.Send(printer.DefaultVendorID, printer.DefaultProductID, data)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, apiResponse{OK: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		OK:           true,
		Message:      "Struk berhasil dikirim ke printer.",
		Bytes:        result.BytesWritten,
		Endpoint:     result.Endpoint,
		TransferType: result.TransferType,
	})
}

func decodeTemplate(w http.ResponseWriter, r *http.Request) (receipt.Template, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var template receipt.Template
	if err := decoder.Decode(&template); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{
			OK:      false,
			Message: "Template JSON tidak valid: " + err.Error(),
		})
		return receipt.Template{}, false
	}
	return template, true
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("gagal menulis respons JSON: %v", err)
	}
}

func openBrowser(url string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		command = exec.Command("open", url)
	default:
		command = exec.Command("xdg-open", url)
	}
	return command.Start()
}
