package api

import (
	"fmt"
)

// Playlist represents a YouTube Music playlist
type Playlist struct {
	ID           string
	PlaylistTitle string
	PlaylistDesc string
	TrackCount   int
	Author       string
	Tracks       []Track // Tracks included in the playlist
}

// FilterValue implements list.Item interface for filtering
func (p Playlist) FilterValue() string { 
	return p.PlaylistTitle + " " + p.Author 
}

// Title implements list.Item interface for displaying in the list
func (p Playlist) Title() string {
	return p.PlaylistTitle
}

// Description implements list.Item interface for displaying in the list
func (p Playlist) Description() string {
	return fmt.Sprintf("by %s (%d tracks)", p.Author, p.TrackCount)
}

