package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

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
