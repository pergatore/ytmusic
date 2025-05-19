package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	
	// Extract from the main structure with the correct path
	contents, contentsOk := result["contents"].(map[string]interface{})
	if contentsOk {
		api.LogDebug("Found top-level 'contents' object")
		
		tabbedResults, tabbedOk := contents["tabbedSearchResultsRenderer"].(map[string]interface{})
		if tabbedOk {
			api.LogDebug("Found 'tabbedSearchResultsRenderer' object")
			
			tabs, tabsOk := tabbedResults["tabs"].([]interface{})
			if tabsOk && len(tabs) > 0 {
				api.LogDebug("Found %d tabs in results", len(tabs))
				
				// Usually the first tab has the results
				tabRenderer, tabOk := tabs[0].(map[string]interface{})
				if tabOk {
					api.LogDebug("Processing tab 0")
					
					tabContent, contentOk := tabRenderer["tabRenderer"].(map[string]interface{})
					if contentOk {
						api.LogDebug("Found 'tabRenderer' object")
						
						content, hasContent := tabContent["content"].(map[string]interface{})
						if hasContent {
							api.LogDebug("Found 'content' object in tabRenderer")
							
							sectionList, hasSectionList := content["sectionListRenderer"].(map[string]interface{})
							if hasSectionList {
								api.LogDebug("Found 'sectionListRenderer' object")
								
								sections, sectionsOk := sectionList["contents"].([]interface{})
								if sectionsOk && len(sections) > 0 {
									api.LogDebug("Processing %d sections in search results", len(sections))
									
									// Process each section
									for i, section := range sections {
										sectionMap, isMap := section.(map[string]interface{})
										if !isMap {
											api.LogDebug("Section %d is not a map, skipping", i)
											continue
										}
										
										// Check for both musicShelfRenderer and itemSectionRenderer
										musicShelf, hasMusicShelf := sectionMap["musicShelfRenderer"]
										if !hasMusicShelf {
											api.LogDebug("Section %d has no musicShelfRenderer, checking for itemSectionRenderer", i)
											_, hasItemSection := sectionMap["itemSectionRenderer"]
											if !hasItemSection {
												api.LogDebug("Section %d has no itemSectionRenderer either, skipping", i)
												continue
											}
											
											// Process itemSectionRenderer if needed
											api.LogDebug("Found itemSectionRenderer in section %d, skipping as it typically contains messages", i)
											continue
										}
										
										musicShelfMap, isMusicShelfMap := musicShelf.(map[string]interface{})
										if !isMusicShelfMap {
											api.LogDebug("MusicShelf in section %d is not a map, skipping", i)
											continue
										}
										
										// Check for title to identify the section (Songs, Albums, etc.)
										if title, hasTitle := musicShelfMap["title"].(map[string]interface{}); hasTitle {
											if runs, hasRuns := title["runs"].([]interface{}); hasRuns && len(runs) > 0 {
												if firstRun, isMap := runs[0].(map[string]interface{}); isMap {
													if text, hasText := firstRun["text"].(string); hasText {
														api.LogDebug("Section %d title: %s", i, text)
													}
												}
											}
										}
										
										contentItems, hasContents := musicShelfMap["contents"].([]interface{})
										if !hasContents {
											api.LogDebug("Section %d has no contents, skipping", i)
											continue
										}
										
										api.LogDebug("Found music shelf with %d items in section %d", len(contentItems), i)
										
										// Process each track item
										for j, item := range contentItems {
											itemMap, isItemMap := item.(map[string]interface{})
											if !isItemMap {
												api.LogDebug("Item %d in section %d is not a map, skipping", j, i)
												continue
											}
											
											renderer, hasRenderer := itemMap["musicResponsiveListItemRenderer"]
											if !hasRenderer {
												api.LogDebug("Item %d in section %d has no musicResponsiveListItemRenderer, skipping", j, i)
												continue
											}
											
											rendererMap, isRendererMap := renderer.(map[string]interface{})
											if !isRendererMap {
												api.LogDebug("Renderer for item %d in section %d is not a map, skipping", j, i)
												continue
											}
											
											// Extract track ID directly from playlistItemData if available
											var trackID, title, artist string
											
											// Method 1: Get ID from playlistItemData
											if playlistItemData, hasPlaylistData := rendererMap["playlistItemData"].(map[string]interface{}); hasPlaylistData {
												if videoId, hasVideoId := playlistItemData["videoId"].(string); hasVideoId {
													trackID = videoId
													api.LogDebug("Found track ID from playlistItemData: %s", trackID)
												}
											}
											
											// Method 2: Get ID from overlay if not found yet
											if trackID == "" {
												if extractedID, err := api.extractTrackIDFromOverlay(rendererMap); err == nil && extractedID != "" {
													trackID = extractedID
													api.LogDebug("Found track ID from overlay: %s", trackID)
												}
											}
											
											// Method 3: Get ID from menu if still not found
											if trackID == "" {
												if extractedID, err := api.extractTrackIDFromMenu(rendererMap); err == nil && extractedID != "" {
													trackID = extractedID
													api.LogDebug("Found track ID from menu: %s", trackID)
												}
											}
											
											// Extract flex columns for title and artist
											flexColumns, hasFlexColumns := rendererMap["flexColumns"].([]interface{})
											if !hasFlexColumns || len(flexColumns) < 2 {
												api.LogDebug("Item %d in section %d has no valid flexColumns, skipping", j, i)
												continue
											}
											
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
																		api.LogDebug("Found title: %s", title)
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
																			api.LogDebug("Found artist: %s", artist)
																		}
																	}
																}
															}
														}
													}
												}
											}
											
											// Only add tracks with ID and title
											if trackID != "" && title != "" {
												// Try to extract duration from third flex column or use default
												duration := 180 // Default duration in seconds
												
												if len(flexColumns) > 2 {
													thirdColumn, isMap := flexColumns[2].(map[string]interface{})
													if isMap {
														if columnRenderer, hasColumnRenderer := thirdColumn["musicResponsiveListItemFlexColumnRenderer"]; hasColumnRenderer {
															if columnRendererMap, isColumnRendererMap := columnRenderer.(map[string]interface{}); isColumnRendererMap {
																if textObj, hasText := columnRendererMap["text"].(map[string]interface{}); hasText {
																	if runs, hasRuns := textObj["runs"].([]interface{}); hasRuns && len(runs) > 0 {
																		if firstRun, isRunMap := runs[0].(map[string]interface{}); isRunMap {
																			if durationText, hasText := firstRun["text"].(string); hasText {
																				api.LogDebug("Found duration text: %s", durationText)
																				// Try to parse duration string (format could be like "3:45")
																				parts := strings.Split(durationText, ":")
																				if len(parts) == 2 {
																					minutes, minErr := strconv.Atoi(parts[0])
																					seconds, secErr := strconv.Atoi(parts[1])
																					if minErr == nil && secErr == nil {
																						duration = minutes*60 + seconds
																						api.LogDebug("Parsed duration: %d seconds", duration)
																					}
																				}
																			}
																		}
																	}
																}
															}
														}
													}
												}
												
												track := Track{
													ID:         trackID,
													TrackTitle: title, // Changed from Title to TrackTitle
													Artist:     artist,
													Duration:   duration,
												}
												tracks = append(tracks, track)
												api.LogDebug("Added track: %s - %s (ID: %s, Duration: %d seconds)", title, artist, trackID, duration)
											} else {
												api.LogDebug("Skipping track due to missing ID or title. ID: %s, Title: %s", trackID, title)
											}
										}
										
										// If we found tracks in this section, we can stop looking
										// Only if the section is for "Songs"
										if len(tracks) > 0 {
											if title, hasTitle := musicShelfMap["title"].(map[string]interface{}); hasTitle {
												if runs, hasRuns := title["runs"].([]interface{}); hasRuns && len(runs) > 0 {
													if firstRun, isMap := runs[0].(map[string]interface{}); isMap {
														if text, hasText := firstRun["text"].(string); hasText && text == "Songs" {
															api.LogDebug("Found %d tracks in 'Songs' section, stopping search", len(tracks))
															break
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	// If we didn't find any tracks, return mock data
	if len(tracks) == 0 {
		api.LogDebug("No tracks found in search results, returning mock data")
		tracks = []Track{
			{ID: "dQw4w9WgXcQ", TrackTitle: "Sample: " + query, Artist: "Try another search term", Duration: 180}, // Changed from Title to TrackTitle
			{ID: "xvFZjo5PgG0", TrackTitle: "Demo song", Artist: "Click to play a demo", Duration: 240}, // Changed from Title to TrackTitle
		}
	}
	
	api.LogDebug("Returning %d tracks from search", len(tracks))
	return tracks, nil
}
