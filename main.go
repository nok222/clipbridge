package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/getlantern/systray"
	"golang.design/x/clipboard"
)

//go:embed icon.ico
var iconData []byte

type ClipboardRequest struct {
	Text string `json:"text"`
}

type ClipboardResponse struct {
	Success bool   `json:"success"`
	Text    string `json:"text,omitempty"`
	Error   string `json:"error,omitempty"`
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleClipboard(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		var text string

		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/json") {
			var req ClipboardRequest
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: false, Error: "Failed to read request body"})
				return
			}
			if err := json.Unmarshal(bodyBytes, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: false, Error: "Invalid JSON format"})
				return
			}
			text = req.Text
		} else {
			// Support fallback raw body if sent as text/plain
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: false, Error: "Failed to read request body"})
				return
			}
			text = string(bodyBytes)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		clipboard.Write(ctx, clipboard.FmtText, []byte(text))

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: true, Text: text})

	case http.MethodGet:
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, err := clipboard.Read(ctx, clipboard.FmtText)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: false, Error: fmt.Sprintf("Failed to read clipboard: %v", err)})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: true, Text: string(data)})

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(ClipboardResponse{Success: false, Error: "Method not allowed"})
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"server": "cp-server",
		"port":   8765,
	})
}

func startServer(addr string) {
	http.HandleFunc("/clipboard", handleClipboard)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handleHealth(w, r)
			return
		}
		http.NotFound(w, r)
	})

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}

func getStartupShortcutPath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", "Startup", "cp-server.lnk"), nil
}

func isRunAtStartupEnabled() bool {
	path, err := getStartupShortcutPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func createShortcut(shortcutPath, targetExe, workDir string) error {
	psScript := fmt.Sprintf(`$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut('%s'); $s.TargetPath = '%s'; $s.WorkingDirectory = '%s'; $s.Save()`,
		strings.ReplaceAll(shortcutPath, `'`, `''`),
		strings.ReplaceAll(targetExe, `'`, `''`),
		strings.ReplaceAll(workDir, `'`, `''`),
	)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func toggleRunAtStartup(mItem *systray.MenuItem) {
	shortcutPath, err := getStartupShortcutPath()
	if err != nil {
		return
	}

	if isRunAtStartupEnabled() {
		_ = os.Remove(shortcutPath)
		mItem.Uncheck()
	} else {
		exePath, err := os.Executable()
		if err != nil {
			return
		}
		workDir := filepath.Dir(exePath)
		_ = createShortcut(shortcutPath, exePath, workDir)
		if isRunAtStartupEnabled() {
			mItem.Check()
		}
	}
}

func onReady() {
	systray.SetIcon(iconData)
	systray.SetTitle("Clipboard Server")
	systray.SetTooltip("Clipboard Server (Port 8765) - Running")

	mStatus := systray.AddMenuItem("Status: Running on :8765", "Server status")
	mStatus.Disable()

	systray.AddSeparator()

	mStartup := systray.AddMenuItemCheckbox("Start with Windows", "Toggle running this server at Windows startup", isRunAtStartupEnabled())

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit Clipboard Server", "Exit the server")

	go func() {
		for {
			select {
			case <-mStartup.ClickedCh:
				toggleRunAtStartup(mStartup)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	os.Exit(0)
}

func main() {
	port := flag.Int("port", 8765, "Port to listen on")
	bind := flag.String("bind", "0.0.0.0", "IP interface to bind to")
	flag.Parse()

	// Initialize clipboard
	err := clipboard.Init()
	if err != nil {
		log.Fatalf("Failed to initialize system clipboard: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", *bind, *port)

	// Run HTTP server in background goroutine
	go startServer(addr)

	// Run system tray (must be on main thread for Windows message loop)
	systray.Run(onReady, onExit)
}
