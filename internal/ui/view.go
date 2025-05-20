package ui

import (
	"fmt"
	"strings"
	
	"ytmusic/internal/player"
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
			"You need to authenticate with YouTube Music to use this application.\n\n" +
			"Please run this command in a separate terminal:\n" +
			warningStyle.Render("ytmusicapi browser --file ~/.ytmusic/headers_auth.json") + "\n\n" +
			"Then restart this application.\n\n" +
			"Press 'q' to quit.")
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
	
	// Currently active list
	var listView string
	if m.ViewMode == ViewTracks {
		// Show track list with search results info if we have some
		if m.SearchResults > 0 && !m.SearchMode {
			s.WriteString(resultInfoStyle.Render(fmt.Sprintf("Found %d tracks. Use ↑/↓ to navigate and Enter to play.\n\n", m.SearchResults)))
		}
		listView = m.TrackList.View()
	} else {
		// Show playlist list
		listView = m.PlaylistList.View()
	}
	
	// Search input
	if m.SearchMode {
		searchView := m.SearchInput.View()
		s.WriteString(fmt.Sprintf("%s\n\n%s\n\n%s",
			titleStyle.Render("YouTube Music - Search"),
			searchView,
			listView))
	} else {
		// Current playing info
		currentlyPlaying := renderPlayingInfo(m)
		
		// Status bar with controls
		statusBar := renderStatusBar(m)
		
		s.WriteString(fmt.Sprintf("%s\n\n%s\n\n%s",
			listView,
			currentlyPlaying,
			statusBar))
	}
	
	return appStyle.Render(s.String())
}

// renderPlayingInfo renders the currently playing track info with progress bar
func renderPlayingInfo(m *Model) string {
	currentTrack := m.Player.Queue.GetCurrentTrack()
	
	if currentTrack != nil {
		// Get status icons
		playStatus := "⏸️"
		if m.Player.IsPlaying {
			playStatus = "▶️"
		}
		
		// Get repeat mode icon
		repeatIcon := ""
		switch m.Player.Queue.RepeatMode {
		case player.RepeatNone:
			repeatIcon = "🔁 Off"
		case player.RepeatOne:
			repeatIcon = "🔂 One"
		case player.RepeatAll:
			repeatIcon = "🔁 All"
		}
		
		// Get shuffle mode icon
		shuffleIcon := "🔀 Off"
		if m.Player.Queue.ShuffleMode {
			shuffleIcon = "🔀 On"
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
		
		playbackControls := fmt.Sprintf("  %s  %s", repeatIcon, shuffleIcon)
		
		// Add queue position info
		queueInfo := ""
		if len(m.Player.Queue.Tracks) > 0 {
			currentIndex := 0
			totalTracks := len(m.Player.Queue.Tracks)
			
			for i, track := range m.Player.Queue.Tracks {
				if track.ID == currentTrack.ID {
					currentIndex = i + 1
					break
				}
			}
			
			queueInfo = fmt.Sprintf(" (%d/%d in queue)", currentIndex, totalTracks)
		}
		
		return fmt.Sprintf(
			"%s %s - %s%s\n%s\n%s%s",
			playStatus,
			playingStyle.Render(currentTrack.TrackTitle),
			infoStyle.Render(currentTrack.Artist),
			queueInfo,
			progressBar,
			timeInfo,
			playbackControls,
		)
	} else {
		return "No song playing"
	}
}

// renderStatusBar renders the status bar with controls
func renderStatusBar(m *Model) string {
	// Basic controls
	controls := []string{
		"[q] Quit",
		"[↑/↓] Navigate",
		"[Enter] Play/Select",
		"[Space] Pause/Play",
		"[/] Search",
	}
	
	// Add playback controls
	controls = append(controls, 
		"[n] Next",
		"[b] Previous",
		"[r] Repeat Mode",
		"[s] Shuffle",
	)
	
	// Add view toggle
	viewToggle := "[p] Show Playlists"
	if m.ViewMode == ViewPlaylists {
		viewToggle = "[p] Show Tracks"
	}
	controls = append(controls, viewToggle)
	
	// Add reset cookie
	controls = append(controls, "[R] Reset Cookie")
	
	return statusBarStyle.Render(strings.Join(controls, "  "))
}
