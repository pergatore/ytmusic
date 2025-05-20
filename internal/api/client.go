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

// YouTubeMusicAPI handles API requests to YouTube Music via Python bridge
type YouTubeMusicAPI struct {
	client     *http.Client
	configPath string
	IsLoggedIn bool
	logger     *log.Logger
	bridge     *PythonBridge // Use the Python bridge instead of direct HTTP calls
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
		IsLoggedIn: false,
		logger:     logger,
	}

	// Initialize Python bridge
	api.bridge = NewPythonBridge(configPath, api.LogDebug)
	api.bridge.SetAPI(api)

	// Try to load cookies
	api.loadCookies()
	
	if debugMode && logger != nil {
		logger.Println("YouTubeMusicAPI initialized")
		logger.Printf("Login status: %v", api.IsLoggedIn)
		logger.Printf("Python bridge available: %v", api.bridge.IsAvailable())
	}

	return api
}

// LogDebug logs messages if in debug mode
func (api *YouTubeMusicAPI) LogDebug(format string, v ...interface{}) {
	if api.logger != nil {
		api.logger.Printf(format, v...)
	}
}

// Search searches for tracks using the Python bridge
func (api *YouTubeMusicAPI) Search(query string) ([]Track, error) {
	if !api.IsLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	api.LogDebug("Searching for: %s", query)

	// Check if Python bridge is available
	if !api.bridge.IsAvailable() {
		api.LogDebug("Python bridge not available, falling back to placeholder results")
		// Return some placeholder results
		return []Track{
			{ID: "dQw4w9WgXcQ", TrackTitle: "Sample: " + query, Artist: "Python bridge not available", Duration: 180},
			{ID: "xvFZjo5PgG0", TrackTitle: "Install ytmusicapi", Artist: "pip install ytmusicapi", Duration: 240},
		}, nil
	}

	// Use Python bridge
	tracks, err := api.bridge.Search(query)
	if err != nil {
		api.LogDebug("Python bridge search failed: %v", err)
		return nil, err
	}

	api.LogDebug("Found %d tracks via Python bridge", len(tracks))
	return tracks, nil
}

// GetUserPlaylists fetches playlists using the Python bridge
func (api *YouTubeMusicAPI) GetUserPlaylists() ([]Playlist, error) {
	if !api.IsLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	api.LogDebug("Fetching user playlists via Python bridge")

	// Check if Python bridge is available
	if !api.bridge.IsAvailable() {
		api.LogDebug("Python bridge not available, returning placeholder playlists")
		return []Playlist{
			{ID: "PLACEHOLDER_1", PlaylistTitle: "Python Bridge Not Available", PlaylistDesc: "Install ytmusicapi", TrackCount: 0, Author: "System"},
			{ID: "PLACEHOLDER_2", PlaylistTitle: "Install Dependencies", PlaylistDesc: "pip install ytmusicapi", TrackCount: 0, Author: "System"},
		}, nil
	}

	// Use Python bridge
	playlists, err := api.bridge.GetPlaylists()
	if err != nil {
		api.LogDebug("Python bridge get playlists failed: %v", err)
		return nil, err
	}

	api.LogDebug("Found %d playlists via Python bridge", len(playlists))
	return playlists, nil
}

// GetPlaylistTracks fetches playlist tracks using the Python bridge
func (api *YouTubeMusicAPI) GetPlaylistTracks(playlistID string) ([]Track, error) {
	if !api.IsLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	api.LogDebug("Fetching playlist tracks for ID: %s via Python bridge", playlistID)

	// Check if Python bridge is available
	if !api.bridge.IsAvailable() {
		api.LogDebug("Python bridge not available, returning placeholder tracks")
		return []Track{
			{ID: "dQw4w9WgXcQ", TrackTitle: "Python Bridge Required", Artist: "Install ytmusicapi", Duration: 180},
		}, nil
	}

	// Use Python bridge
	tracks, err := api.bridge.GetPlaylistTracks(playlistID)
	if err != nil {
		api.LogDebug("Python bridge get playlist tracks failed: %v", err)
		return nil, err
	}

	api.LogDebug("Found %d tracks in playlist via Python bridge", len(tracks))
	return tracks, nil
}
