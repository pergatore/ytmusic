package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"ytmusic/internal/utils"
)

// Track represents a music track
type Track struct {
	ID       string
	Title    string
	Artist   string
	Duration int // in seconds
}

// For list.Item interface
func (t Track) FilterValue() string { return t.Title + " " + t.Artist }

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

// loadCookies loads cookies from the config file
func (api *YouTubeMusicAPI) loadCookies() {
	cookiePath := filepath.Join(api.configPath, "cookies.json")
	
	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		api.LogDebug("No cookies file found at %s", cookiePath)
		return
	}

	data, err := os.ReadFile(cookiePath)
	if err != nil {
		api.LogDebug("Error reading cookies file: %v", err)
		return
	}

	var cookies []*http.Cookie
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		api.LogDebug("Error unmarshalling cookies: %v", err)
		return
	}

	// Check if we have the required cookies
	for _, cookie := range cookies {
		if cookie.Name == "__Secure-3PSID" {
			api.IsLoggedIn = true
			api.LogDebug("Found valid __Secure-3PSID cookie, setting logged in")
			break
		}
	}

	// Set cookies on client
	if api.IsLoggedIn {
		ytMusicURL, _ := url.Parse("https://music.youtube.com")
		api.client.Jar.SetCookies(ytMusicURL, cookies)
		api.LogDebug("Loaded %d cookies into client", len(cookies))
	}
}

// saveCookies saves cookies to the config file
func (api *YouTubeMusicAPI) saveCookies() error {
	ytMusicURL, _ := url.Parse("https://music.youtube.com")
	cookies := api.client.Jar.Cookies(ytMusicURL)
	
	api.LogDebug("Saving %d cookies", len(cookies))
	
	data, err := json.Marshal(cookies)
	if err != nil {
		api.LogDebug("Error marshalling cookies: %v", err)
		return err
	}
	
	cookiePath := filepath.Join(api.configPath, "cookies.json")
	return os.WriteFile(cookiePath, data, 0644)
}

// ResetCookies removes saved cookies and resets login state
func (api *YouTubeMusicAPI) ResetCookies() error {
	api.LogDebug("Resetting cookies")
	
	// Clear cookies in the client
	api.client.Jar, _ = cookiejar.New(nil)
	api.IsLoggedIn = false
	
	// Remove the cookies file
	cookiePath := filepath.Join(api.configPath, "cookies.json")
	if _, err := os.Stat(cookiePath); !os.IsNotExist(err) {
		api.LogDebug("Removing cookies file at %s", cookiePath)
		err = os.Remove(cookiePath)
		if err != nil {
			api.LogDebug("Error removing cookies file: %v", err)
			return err
		}
	}
	
	return nil
}

// ManualLogin handles manual login with a provided cookie
func (api *YouTubeMusicAPI) ManualLogin(cookie string) error {
	if cookie == "" {
		return fmt.Errorf("no cookie provided")
	}
	
	api.LogDebug("Manual login attempt with cookie length: %d", len(cookie))
	
	// Set the cookie
	ytMusicURL, _ := url.Parse("https://music.youtube.com")
	api.client.Jar.SetCookies(ytMusicURL, []*http.Cookie{
		{
			Name:   "__Secure-3PSID",
			Value:  cookie,
			Domain: ".youtube.com",
			Path:   "/",
			Secure: true,
		},
	})
	
	api.IsLoggedIn = true
	return api.saveCookies()
}

// InitiateLogin starts the login process
func (api *YouTubeMusicAPI) InitiateLogin() error {
    api.LogDebug("Initiating login process")
    
    // Try to open browser to YouTube Music
    utils.ClearScreen()
    fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
    fmt.Println("│ Attempting to open YouTube Music in your browser...               │")
    fmt.Println("└───────────────────────────────────────────────────────────────────┘")
    
    browserOpened := utils.OpenBrowser("https://music.youtube.com")
    
    if !browserOpened {
        utils.ClearScreen()
        fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
        fmt.Println("│ Could not open browser automatically.                             │")
        fmt.Println("└───────────────────────────────────────────────────────────────────┘")
        fmt.Println("")
        fmt.Println("To get your cookie manually:")
        fmt.Println("  1. Open https://music.youtube.com in your browser")
        fmt.Println("  2. Log in if you're not already logged in")
        fmt.Println("  3. Open developer tools (F12 or right-click > Inspect)")
        fmt.Println("  4. Go to Application/Storage tab > Cookies > music.youtube.com")
        fmt.Println("  5. Find the '__Secure-3PSID' cookie and copy its value")
    } else {
        utils.ClearScreen()
        fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
        fmt.Println("│ Browser opened to YouTube Music.                                  │")
        fmt.Println("└───────────────────────────────────────────────────────────────────┘")
        fmt.Println("")
        fmt.Println("Please follow these steps:")
        fmt.Println("  1. Log in if you're not already logged in")
        fmt.Println("  2. Open developer tools (F12 or right-click > Inspect)")
        fmt.Println("  3. Go to Application/Storage tab > Cookies > music.youtube.com")
        fmt.Println("  4. Find the '__Secure-3PSID' cookie and copy its value")
    }
    
    fmt.Println("")
    fmt.Println("IMPORTANT: Make sure you're getting the cookie from music.youtube.com,")
    fmt.Println("not from google.com. The correct domain is .youtube.com")
    fmt.Println("")
    fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
    fmt.Print("│ Paste the cookie value here: ")
	
	var cookie string
	fmt.Scanln(&cookie)
	
	if cookie == "" {
		api.LogDebug("No cookie provided during login")
		return fmt.Errorf("no cookie provided")
	}
	
	api.LogDebug("Received cookie input with length: %d", len(cookie))
	
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	
	// Set the cookie
	ytMusicURL, _ := url.Parse("https://music.youtube.com")
	api.client.Jar.SetCookies(ytMusicURL, []*http.Cookie{
		{
			Name:   "__Secure-3PSID",
			Value:  cookie,
			Domain: ".youtube.com",
			Path:   "/",
			Secure: true,
		},
	})
	
	api.IsLoggedIn = true
	utils.ClearScreen()
	fmt.Println("┌─────────────────────────────────────────────────────────┐")
	fmt.Println("│ Login successful! Press any key to continue.            │")
	fmt.Println("└─────────────────────────────────────────────────────────┘")
	return api.saveCookies()
}

// Search searches for tracks on YouTube Music
func (api *YouTubeMusicAPI) Search(query string) ([]Track, error) {
	if !api.IsLoggedIn {
		return nil, fmt.Errorf("not logged in")
	}

	api.LogDebug("Searching for: %s", query)
	endpoint := "https://music.youtube.com/youtubei/v1/search"
	
	// Build the proper request payload that YouTube Music expects
	requestData := map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"clientName":    "WEB_REMIX",
				"clientVersion": "1.20230815.01.00",
				"hl":            "en",
				"gl":            "US",
			},
		},
		"query": query,
		"params": "EgWKAQIIAWoKEAMQBBAJEAoQBQ%3D%3D", // This parameter targets songs specifically
	}
	
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		api.LogDebug("Error marshalling search request: %v", err)
		return nil, err
	}
	
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		api.LogDebug("Error creating search request: %v", err)
		return nil, err
	}
	
	// Set headers
	for k, v := range api.headers {
		req.Header.Set(k, v)
	}
	
	// Add additional headers that may be needed
	req.Header.Set("X-YouTube-Client-Name", "67")
	req.Header.Set("X-YouTube-Client-Version", "1.20230815.01.00")
	
	// Make request
	api.LogDebug("Sending search request to %s", endpoint)
	resp, err := api.client.Do(req)
	if err != nil {
		api.LogDebug("Error making search request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		api.LogDebug("Search API returned non-OK status: %s", resp.Status)
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}
	
	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		api.LogDebug("Error reading search response body: %v", err)
		return nil, err
	}
	
	// Log response size in debug mode
	api.LogDebug("Received search response with size: %d bytes", len(body))
	
	// Parse response JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		api.LogDebug("Error unmarshalling search response: %v", err)
		return nil, err
	}
	
	// Save the response to a file for debugging
	if api.logger != nil {
		debugFile := filepath.Join(api.configPath, "logs", "last_response.json")
		if err := os.WriteFile(debugFile, body, 0644); err != nil {
			api.LogDebug("Error saving response JSON: %v", err)
		} else {
			api.LogDebug("Saved response JSON to %s", debugFile)
		}
	}
	
	// Extract tracks from the response
	var tracks []Track
	
	// Extract from the main structure
	contents, contentsOk := result["contents"].(map[string]interface{})
	if contentsOk {
		sectionList, sectionOk := contents["sectionListRenderer"].(map[string]interface{})
		if sectionOk {
			sections, sectionsOk := sectionList["contents"].([]interface{})
			if sectionsOk && len(sections) > 0 {
				api.LogDebug("Processing %d sections in search results", len(sections))
				
				// Process each section
				for _, section := range sections {
					sectionMap, isMap := section.(map[string]interface{})
					if !isMap {
						continue
					}
					
					musicShelf, hasMusicShelf := sectionMap["musicShelfRenderer"]
					if !hasMusicShelf {
						continue
					}
					
					musicShelfMap, isMusicShelfMap := musicShelf.(map[string]interface{})
					if !isMusicShelfMap {
						continue
					}
					
					contentItems, hasContents := musicShelfMap["contents"].([]interface{})
					if !hasContents {
						continue
					}
					
					api.LogDebug("Found music shelf with %d items", len(contentItems))
					
					// Process each track item
					for _, item := range contentItems {
						itemMap, isItemMap := item.(map[string]interface{})
						if !isItemMap {
							continue
						}
						
						renderer, hasRenderer := itemMap["musicResponsiveListItemRenderer"]
						if !hasRenderer {
							continue
						}
						
						rendererMap, isRendererMap := renderer.(map[string]interface{})
						if !isRendererMap {
							continue
						}
						
						// Extract track info
						flexColumns, hasFlexColumns := rendererMap["flexColumns"].([]interface{})
						if !hasFlexColumns || len(flexColumns) < 2 {
							continue
						}
						
						// Safely extract title
						var title, artist, trackID string
						
						// Title is usually in the first flex column
						firstColumn, isMap := flexColumns[0].(map[string]interface{})
						if isMap {
							columnRenderer, hasColumnRenderer := firstColumn["musicResponsiveListItemFlexColumnRenderer"]
							if hasColumnRenderer {
								columnRendererMap, isColumnRendererMap := columnRenderer.(map[string]interface{})
								if isColumnRendererMap {
									textObj, hasText := columnRendererMap["text"].(map[string]interface{})
									if hasText {
										runs, hasRuns := textObj["runs"].([]interface{})
										if hasRuns && len(runs) > 0 {
											firstRun, isRunMap := runs[0].(map[string]interface{})
											if isRunMap {
												titleText, hasText := firstRun["text"].(string)
												if hasText {
													title = titleText
												}
											}
										}
									}
								}
							}
						}
						
						// Artist is usually in the second flex column
						if len(flexColumns) > 1 {
							secondColumn, isMap := flexColumns[1].(map[string]interface{})
							if isMap {
								columnRenderer, hasColumnRenderer := secondColumn["musicResponsiveListItemFlexColumnRenderer"]
								if hasColumnRenderer {
									columnRendererMap, isColumnRendererMap := columnRenderer.(map[string]interface{})
									if isColumnRendererMap {
										textObj, hasText := columnRendererMap["text"].(map[string]interface{})
										if hasText {
											runs, hasRuns := textObj["runs"].([]interface{})
											if hasRuns && len(runs) > 0 {
												firstRun, isRunMap := runs[0].(map[string]interface{})
												if isRunMap {
													artistText, hasText := firstRun["text"].(string)
													if hasText {
														artist = artistText
													}
												}
											}
										}
									}
								}
							}
						}
						
						// Try to extract video ID from various possible locations
						// Method 1: playlistItemData
						playlistItemData, hasPlaylistData := rendererMap["playlistItemData"]
						if hasPlaylistData {
							playlistMap, isPlaylistMap := playlistItemData.(map[string]interface{})
							if isPlaylistMap {
								videoId, hasVideoId := playlistMap["videoId"].(string)
								if hasVideoId {
									trackID = videoId
								}
							}
						}
						
						// Method 2: overlay
						if trackID == "" {
							if trackID, err = api.extractTrackIDFromOverlay(rendererMap); err == nil && trackID != "" {
								// Track ID found
							}
						}
						
						// Method 3: navigationEndpoint
						if trackID == "" {
							if trackID, err = api.extractTrackIDFromMenu(rendererMap); err == nil && trackID != "" {
								// Track ID found
							}
						}
						
						// Only add tracks with ID and title
						if trackID != "" && title != "" {
							track := Track{
								ID:       trackID,
								Title:    title,
								Artist:   artist,
								Duration: 180, // Default duration since it's hard to extract
							}
							tracks = append(tracks, track)
							api.LogDebug("Added track: %s - %s (ID: %s)", title, artist, trackID)
						}
					}
					
					// If we found tracks in this section, we can stop looking
					if len(tracks) > 0 {
						break
					}
				}
			}
		}
	}
	
	// If we didn't find any tracks, return mock data
	if len(tracks) == 0 {
		api.LogDebug("No tracks found in search results, returning mock data")
		tracks = []Track{
			{ID: "dQw4w9WgXcQ", Title: "Sample: " + query, Artist: "Try another search term", Duration: 180},
			{ID: "xvFZjo5PgG0", Title: "Demo song", Artist: "Click to play a demo", Duration: 240},
		}
	}
	
	api.LogDebug("Returning %d tracks from search", len(tracks))
	return tracks, nil
}

// extractTrackIDFromOverlay extracts a track ID from the overlay renderer
func (api *YouTubeMusicAPI) extractTrackIDFromOverlay(rendererMap map[string]interface{}) (string, error) {
	overlay, hasOverlay := rendererMap["overlay"]
	if !hasOverlay {
		return "", fmt.Errorf("no overlay found")
	}
	
	overlayMap, isOverlayMap := overlay.(map[string]interface{})
	if !isOverlayMap {
		return "", fmt.Errorf("overlay is not a map")
	}
	
	thumbnailOverlay, hasThumbnailOverlay := overlayMap["musicItemThumbnailOverlayRenderer"]
	if !hasThumbnailOverlay {
		return "", fmt.Errorf("no thumbnail overlay found")
	}
	
	thumbnailMap, isThumbnailMap := thumbnailOverlay.(map[string]interface{})
	if !isThumbnailMap {
		return "", fmt.Errorf("thumbnail overlay is not a map")
	}
	
	content, hasContent := thumbnailMap["content"]
	if !hasContent {
		return "", fmt.Errorf("no content found in thumbnail overlay")
	}
	
	contentMap, isContentMap := content.(map[string]interface{})
	if !isContentMap {
		return "", fmt.Errorf("content is not a map")
	}
	
	playButton, hasPlayButton := contentMap["musicPlayButtonRenderer"]
	if !hasPlayButton {
		return "", fmt.Errorf("no play button found")
	}
	
	playButtonMap, isPlayButtonMap := playButton.(map[string]interface{})
	if !isPlayButtonMap {
		return "", fmt.Errorf("play button is not a map")
	}
	
	navEndpoint, hasNavEndpoint := playButtonMap["playNavigationEndpoint"]
	if !hasNavEndpoint {
		return "", fmt.Errorf("no navigation endpoint found")
	}
	
	navEndpointMap, isNavEndpointMap := navEndpoint.(map[string]interface{})
	if !isNavEndpointMap {
		return "", fmt.Errorf("navigation endpoint is not a map")
	}
	
	watchEndpoint, hasWatchEndpoint := navEndpointMap["watchEndpoint"]
	if !hasWatchEndpoint {
		return "", fmt.Errorf("no watch endpoint found")
	}
	
	watchEndpointMap, isWatchEndpointMap := watchEndpoint.(map[string]interface{})
	if !isWatchEndpointMap {
		return "", fmt.Errorf("watch endpoint is not a map")
	}
	
	videoId, hasVideoId := watchEndpointMap["videoId"].(string)
	if !hasVideoId {
		return "", fmt.Errorf("no video ID found")
	}
	
	return videoId, nil
}

// extractTrackIDFromMenu extracts a track ID from the menu renderer
func (api *YouTubeMusicAPI) extractTrackIDFromMenu(rendererMap map[string]interface{}) (string, error) {
	menu, hasMenu := rendererMap["menu"]
	if !hasMenu {
		return "", fmt.Errorf("no menu found")
	}
	
	menuMap, isMenuMap := menu.(map[string]interface{})
	if !isMenuMap {
		return "", fmt.Errorf("menu is not a map")
	}
	
	menuRenderer, hasMenuRenderer := menuMap["menuRenderer"]
	if !hasMenuRenderer {
		return "", fmt.Errorf("no menu renderer found")
	}
	
	menuRendererMap, isMenuRendererMap := menuRenderer.(map[string]interface{})
	if !isMenuRendererMap {
		return "", fmt.Errorf("menu renderer is not a map")
	}
	
	menuItems, hasMenuItems := menuRendererMap["items"].([]interface{})
	if !hasMenuItems || len(menuItems) == 0 {
		return "", fmt.Errorf("no menu items found")
	}
	
	for _, menuItem := range menuItems {
		menuItemMap, isMenuItemMap := menuItem.(map[string]interface{})
		if !isMenuItemMap {
			continue
		}
		
		menuServiceItem, hasMenuServiceItem := menuItemMap["menuServiceItemRenderer"]
		if !hasMenuServiceItem {
			continue
		}
		
		menuServiceItemMap, isMenuServiceItemMap := menuServiceItem.(map[string]interface{})
		if !isMenuServiceItemMap {
			continue
		}
		
		serviceEndpoint, hasServiceEndpoint := menuServiceItemMap["serviceEndpoint"]
		if !hasServiceEndpoint {
			continue
		}
		
		serviceEndpointMap, isServiceEndpointMap := serviceEndpoint.(map[string]interface{})
		if !isServiceEndpointMap {
			continue
		}
		
		watchEndpoint, hasWatchEndpoint := serviceEndpointMap["watchEndpoint"]
		if !hasWatchEndpoint {
			continue
		}
		
		watchEndpointMap, isWatchEndpointMap := watchEndpoint.(map[string]interface{})
		if !isWatchEndpointMap {
			continue
		}
		
		videoId, hasVideoId := watchEndpointMap["videoId"].(string)
		if hasVideoId {
			return videoId, nil
		}
	}
	
	return "", fmt.Errorf("no video ID found in menu")
}

// GetStreamURL gets the streaming URL for a track
func (api *YouTubeMusicAPI) GetStreamURL(trackID string) (string, error) {
	if !api.IsLoggedIn {
		return "", fmt.Errorf("not logged in")
	}

	api.LogDebug("Getting stream URL for track ID: %s", trackID)
	
	// YouTube Music doesn't provide direct stream URLs easily
	// For our TUI, we'll use the YouTube watch URL which works with mpv
	url := "https://www.youtube.com/watch?v=" + trackID
	
	// For a real implementation, you could use youtube-dl or yt-dlp to extract
	// the actual stream URL, but that would require additional dependencies.
	
	api.LogDebug("Returning stream URL: %s", url)
	return url, nil
}
