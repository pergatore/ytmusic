package api

// Track represents a music track
type Track struct {
	ID       string
	Title    string
	Artist   string
	Duration int // in seconds
}

// For list.Item interface
func (t Track) FilterValue() string { return t.Title + " " + t.Artist }

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
