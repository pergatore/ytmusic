package ui

import (
	"fmt"
	"strings"
)

// View renders the UI and returns it as a string
func (m *Model) View() string {
	if m.ResetMode {
		return appStyle.Render(
			titleStyle.Render("Reset YouTube Music Cookie") + "\n\n" +
			warningStyle.Render("Are you sure you want to reset your login credentials?") + "\n" +
			"This will remove the current cookie and require you to log in again.\n\n" +
			"Press 'y' to confirm or 'n' to cancel.")
	}
	
	if m.LoginMode {
		return appStyle.Render(
			titleStyle.Render("YouTube Music TUI") + "\n\n" +
			"You need to log in to YouTube Music to use this application.\n\n" +
			"Press 'l' to log in or 'q' to quit.\n\n" +
			"When logging in, you'll need to provide the '__Secure-3PSID' cookie\n" +
			"from YouTube Music. Instructions will be provided during login.\n\n" +
			warningStyle.Render("IMPORTANT: Make sure to use the cookie from .youtube.com domain,") + "\n" +
			warningStyle.Render("NOT from google.com domain."))
	}
	
	if m.IsLoading {
		return appStyle.Render(
			titleStyle.Render("YouTube Music TUI") + "\n\n" +
			m.Spinner.View() + " Loading...")
	}
	
	var s strings.Builder
	
	// Error message
	if m.ErrorMsg != "" {
		s.WriteString(errorStyle.Render(m.ErrorMsg) + "\n\n")
	}
	
	// Search or music list
	listView := m.List.View()
	
	// Current playing info
	currentlyPlaying := ""
	if m.CurrentTrack.ID != "" {
		playStatus := "⏸️"
		if m.Player.IsPlaying {
			playStatus = "▶️"
		}
		
		// Format time as MM:SS
		currentMinutes := m.Player.CurrentPos / 60
		currentSeconds := m.Player.CurrentPos % 60
		totalMinutes := m.Player.Duration / 60
		totalSeconds := m.Player.Duration % 60
		
		timeInfo := fmt.Sprintf("%02d:%02d / %02d:%02d", 
			currentMinutes, currentSeconds,
			totalMinutes, totalSeconds)
		
		progressBar := m.Progress.ViewAs(float64(m.Player.CurrentPos) / float64(m.Player.Duration))
		
		currentlyPlaying = fmt.Sprintf(
			"%s %s - %s\n%s\n%s",
			playStatus,
			playingStyle.Render(m.CurrentTrack.Title),
			infoStyle.Render(m.CurrentTrack.Artist),
			progressBar,
			timeInfo,
		)
	} else {
		currentlyPlaying = "No song playing"
	}
	
	// Status bar with controls
	statusBar := statusBarStyle.Render(
		"[q] Quit  [↑/↓] Navigate  [Enter] Play  [Space] Pause/Play  [/] Search  [r] Reset Cookie")
	
	// If in search mode, show search input
	if m.SearchMode {
		searchView := m.SearchInput.View()
		s.WriteString(fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", 
			titleStyle.Render("YouTube Music - Search"), 
			searchView, 
			listView, 
			statusBar))
	} else {
		s.WriteString(fmt.Sprintf("%s\n\n%s\n\n%s", 
			listView, 
			currentlyPlaying, 
			statusBar))
	}
	
	return appStyle.Render(s.String())
}
