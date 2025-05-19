package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"time"
)

// YouTubeMusicAPI handles API requests to YouTube Music
type YouTubeMusicAPI struct {
	client     *http.Client
	headers    map[string]string
	configPath string
	IsLoggedIn bool
	logger     *log.Logger
}

// NewYouTubeMusicAPI creates a new YouTubeMusicAPI instance
func NewYouTubeMusicAPI(debugMode bool) *YouTubeMusicAPI {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 10 * time.Second,
		Jar:     jar,
	}

	configDir, _ := os.UserHomeDir()
	configPath := filepath.Join(configDir, ".ytmusic")
	
	// Create config directory if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		os.MkdirAll(configPath, 0755)
	}
	
	// Create logs directory if it doesn't exist
	logPath := filepath.Join(configPath, "logs")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		os.MkdirAll(logPath, 0755)
	}
	
	// Set up logger
	var logger *log.Logger
	if debugMode {
		logFile := filepath.Join(logPath, fmt.Sprintf("ytmusic_%s.log", time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Error opening log file: %v\n", err)
		} else {
			logger = log.New(f, "YTMusic: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}

	api := &YouTubeMusicAPI{
		client:     client,
		configPath: configPath,
		headers: map[string]string{
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:88.0) Gecko/20100101 Firefox/88.0",
			"Accept":          "*/*",
			"Accept-Language": "en-US,en;q=0.5",
			"Content-Type":    "application/json",
			"X-Goog-AuthUser": "0",
			"Origin":          "https://music.youtube.com",
		},
		IsLoggedIn: false,
		logger:     logger,
	}

	// Try to load cookies
	api.loadCookies()
	
	if debugMode && logger != nil {
		logger.Println("YouTubeMusicAPI initialized")
		logger.Printf("Login status: %v", api.IsLoggedIn)
	}

	return api
}

// LogDebug logs messages if in debug mode
func (api *YouTubeMusicAPI) LogDebug(format string, v ...interface{}) {
	if api.logger != nil {
		api.logger.Printf(format, v...)
	}
}
