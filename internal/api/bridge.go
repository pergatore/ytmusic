package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PythonBridge handles communication with the Python ytmusicapi bridge
type PythonBridge struct {
	pythonPath string
	scriptPath string
	logger     func(format string, v ...interface{})
	api        *YouTubeMusicAPI // Reference to the API for cookie access
}

// BridgeResponse represents the response from the Python bridge
type BridgeResponse struct {
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	Traceback string `json:"traceback,omitempty"`
}

// SearchResponse represents search results from the bridge
type SearchResponse struct {
	BridgeResponse
	Tracks []BridgeTrack `json:"tracks,omitempty"`
}

// PlaylistsResponse represents playlists from the bridge
type PlaylistsResponse struct {
	BridgeResponse
	Playlists []BridgePlaylist `json:"playlists,omitempty"`
}

// BridgeTrack represents a track from the Python bridge
type BridgeTrack struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Duration  int    `json:"duration"`
	Thumbnail string `json:"thumbnail"`
}

// BridgePlaylist represents a playlist from the Python bridge
type BridgePlaylist struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TrackCount  int    `json:"track_count"`
	Author      string `json:"author"`
}

// NewPythonBridge creates a new Python bridge instance
func NewPythonBridge(configPath string, logger func(format string, v ...interface{})) *PythonBridge {
	// Try to find Python executable
	pythonPath := "python3"
	if _, err := exec.LookPath("python3"); err != nil {
		pythonPath = "python"
		if _, err := exec.LookPath("python"); err != nil {
			if logger != nil {
				logger("Warning: Python not found in PATH")
			}
		}
	}
	
	// Determine script path - look for the script in the project directory
	scriptPath := ""
	
	// Try different possible locations
	possiblePaths := []string{
		"scripts/ytmusic_bridge.py",
		"../scripts/ytmusic_bridge.py",
		"../../scripts/ytmusic_bridge.py",
		filepath.Join(configPath, "ytmusic_bridge.py"),
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			scriptPath = path
			break
		}
	}
	
	if scriptPath == "" {
		if logger != nil {
			logger("Warning: ytmusic_bridge.py script not found")
		}
	}
	
	return &PythonBridge{
		pythonPath: pythonPath,
		scriptPath: scriptPath,
		logger:     logger,
	}
}

// SetAPI sets the API reference for cookie access
func (pb *PythonBridge) SetAPI(api *YouTubeMusicAPI) {
	pb.api = api
}

// IsAvailable checks if the Python bridge is available
func (pb *PythonBridge) IsAvailable() bool {
	if pb.scriptPath == "" {
		return false
	}
	
	if _, err := os.Stat(pb.scriptPath); os.IsNotExist(err) {
		return false
	}
	
	if _, err := exec.LookPath(pb.pythonPath); err != nil {
		return false
	}
	
	return true
}

// log helper function
func (pb *PythonBridge) log(format string, v ...interface{}) {
	if pb.logger != nil {
		pb.logger(format, v...)
	}
}

// getCookie extracts the __Secure-3PSID cookie value from the API
func (pb *PythonBridge) getCookie() string {
	if pb.api == nil || !pb.api.IsLoggedIn {
		return ""
	}
	
	// Get cookies from the HTTP client
	ytMusicURL, _ := url.Parse("https://music.youtube.com")
	cookies := pb.api.client.Jar.Cookies(ytMusicURL)
	
	for _, cookie := range cookies {
		if cookie.Name == "__Secure-3PSID" {
			return cookie.Value
		}
	}
	
	return ""
}

// runCommand executes a Python bridge command with cookie authentication
func (pb *PythonBridge) runCommand(args []string) ([]byte, error) {
	if !pb.IsAvailable() {
		return nil, fmt.Errorf("Python bridge not available")
	}
	
	cmdArgs := []string{pb.scriptPath}
	cmdArgs = append(cmdArgs, args...)
	
	// Add cookie if available
	if cookie := pb.getCookie(); cookie != "" {
		cmdArgs = append(cmdArgs, "--cookie", cookie)
	}
	
	pb.log("Running Python bridge command: %s %s", pb.pythonPath, strings.Join(cmdArgs, " "))
	
	cmd := exec.Command(pb.pythonPath, cmdArgs...)
	output, err := cmd.Output()
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			pb.log("Python bridge stderr: %s", string(exitError.Stderr))
		}
		return nil, fmt.Errorf("Python bridge command failed: %v", err)
	}
	
	pb.log("Python bridge output length: %d bytes", len(output))
	return output, nil
}

// Search searches for tracks using the Python bridge
func (pb *PythonBridge) Search(query string) ([]Track, error) {
	args := []string{"search", "--query", query, "--filter", "songs", "--limit", "20"}
	
	output, err := pb.runCommand(args)
	if err != nil {
		return nil, err
	}
	
	var response SearchResponse
	if err := json.Unmarshal(output, &response); err != nil {
		pb.log("Error unmarshaling search response: %v", err)
		return nil, fmt.Errorf("failed to parse search response: %v", err)
	}
	
	if !response.Success {
		pb.log("Search failed: %s", response.Error)
		return nil, fmt.Errorf("search failed: %s", response.Error)
	}
	
	// Convert bridge tracks to API tracks
	tracks := make([]Track, len(response.Tracks))
	for i, bridgeTrack := range response.Tracks {
		tracks[i] = Track{
			ID:         bridgeTrack.ID,
			TrackTitle: bridgeTrack.Title,
			Artist:     bridgeTrack.Artist,
			Duration:   bridgeTrack.Duration,
		}
	}
	
	pb.log("Search returned %d tracks", len(tracks))
	return tracks, nil
}

// GetPlaylists gets user playlists using the Python bridge
func (pb *PythonBridge) GetPlaylists() ([]Playlist, error) {
	args := []string{"playlists", "--limit", "25"}
	
	output, err := pb.runCommand(args)
	if err != nil {
		return nil, err
	}
	
	var response PlaylistsResponse
	if err := json.Unmarshal(output, &response); err != nil {
		pb.log("Error unmarshaling playlists response: %v", err)
		return nil, fmt.Errorf("failed to parse playlists response: %v", err)
	}
	
	if !response.Success {
		pb.log("Get playlists failed: %s", response.Error)
		return nil, fmt.Errorf("get playlists failed: %s", response.Error)
	}
	
	// Convert bridge playlists to API playlists
	playlists := make([]Playlist, len(response.Playlists))
	for i, bridgePlaylist := range response.Playlists {
		playlists[i] = Playlist{
			ID:            bridgePlaylist.ID,
			PlaylistTitle: bridgePlaylist.Title,
			PlaylistDesc:  bridgePlaylist.Description,
			TrackCount:    bridgePlaylist.TrackCount,
			Author:        bridgePlaylist.Author,
		}
	}
	
	pb.log("Get playlists returned %d playlists", len(playlists))
	return playlists, nil
}

// GetPlaylistTracks gets tracks from a playlist using the Python bridge
func (pb *PythonBridge) GetPlaylistTracks(playlistID string) ([]Track, error) {
	args := []string{"playlist_tracks", "--playlist-id", playlistID, "--limit", "100"}
	
	output, err := pb.runCommand(args)
	if err != nil {
		return nil, err
	}
	
	var response SearchResponse
	if err := json.Unmarshal(output, &response); err != nil {
		pb.log("Error unmarshaling playlist tracks response: %v", err)
		return nil, fmt.Errorf("failed to parse playlist tracks response: %v", err)
	}
	
	if !response.Success {
		pb.log("Get playlist tracks failed: %s", response.Error)
		return nil, fmt.Errorf("get playlist tracks failed: %s", response.Error)
	}
	
	// Convert bridge tracks to API tracks
	tracks := make([]Track, len(response.Tracks))
	for i, bridgeTrack := range response.Tracks {
		tracks[i] = Track{
			ID:         bridgeTrack.ID,
			TrackTitle: bridgeTrack.Title,
			Artist:     bridgeTrack.Artist,
			Duration:   bridgeTrack.Duration,
		}
	}
	
	pb.log("Get playlist tracks returned %d tracks", len(tracks))
	return tracks, nil
}

// GetLikedSongs gets user's liked songs using the Python bridge
func (pb *PythonBridge) GetLikedSongs() ([]Track, error) {
	args := []string{"liked_songs", "--limit", "100"}
	
	output, err := pb.runCommand(args)
	if err != nil {
		return nil, err
	}
	
	var response SearchResponse
	if err := json.Unmarshal(output, &response); err != nil {
		pb.log("Error unmarshaling liked songs response: %v", err)
		return nil, fmt.Errorf("failed to parse liked songs response: %v", err)
	}
	
	if !response.Success {
		pb.log("Get liked songs failed: %s", response.Error)
		return nil, fmt.Errorf("get liked songs failed: %s", response.Error)
	}
	
	// Convert bridge tracks to API tracks
	tracks := make([]Track, len(response.Tracks))
	for i, bridgeTrack := range response.Tracks {
		tracks[i] = Track{
			ID:         bridgeTrack.ID,
			TrackTitle: bridgeTrack.Title,
			Artist:     bridgeTrack.Artist,
			Duration:   bridgeTrack.Duration,
		}
	}
	
	pb.log("Get liked songs returned %d tracks", len(tracks))
	return tracks, nil
}
