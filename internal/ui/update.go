package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	
	"ytmusic/internal/api"
	"ytmusic/internal/player"
)

// Update updates the model based on messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case loginStatusMsg:
		m.LoginMode = !msg.isLoggedIn
		if m.LoginMode {
			return m, nil
		}
		
		// If we've just logged in, fetch playlists
		if msg.isLoggedIn {
			m.IsLoading = true
			return m, tea.Batch(
				m.Spinner.Tick,
				GetPlaylistsCmd(m.Api),
			)
		}
		
		return m, nil
		
	case tea.KeyMsg:
		if m.ResetMode {
			// Handle reset mode input
			switch msg.String() {
			case "y", "Y":
				m.IsLoading = true
				return m, ResetCookiesCmd(m.Api)
				
			case "n", "N", "esc", "q", "ctrl+c":
				m.ResetMode = false
				return m, nil
			}
			return m, nil
		} else if m.LoginMode {
			// Handle login mode input
			switch msg.String() {
			case "l":
				// Use a background routine to handle the login process
				go func() {
					err := m.Api.InitiateLogin()
					if err != nil {
						// Handle login error
					} else {
						// Force a refresh of the UI
						p := tea.NewProgram(m)
						p.Send(loginStatusMsg{isLoggedIn: true})
					}
				}()
				return m, nil
				
			case "q", "ctrl+c":
				return m, tea.Quit
			}
			return m, nil
		} else if m.IsLoading {
			// When loading, only handle quit
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			return m, nil
		} else if m.SearchMode {
			// When in search mode, handle Esc, Enter, and pass other keys to input
			switch msg.String() {
			case "esc":
				m.SearchMode = false
				m.SearchInput.Blur()
				return m, nil
				
			case "enter":
				m.SearchMode = false
				m.IsLoading = true
				m.ErrorMsg = "" // Clear previous errors
				query := m.SearchInput.Value()
				if query == "" {
					m.IsLoading = false
					m.ErrorMsg = "Please enter a search term"
					return m, nil
				}
				
				// Switch to tracks view when searching
				m.ViewMode = ViewTracks
				m.ActiveList = &m.TrackList
				
				return m, tea.Batch(
					m.Spinner.Tick,
					SearchCmd(m.Api, query),
				)
				
			default:
				// Pass other keys to text input
				m.SearchInput, cmd = m.SearchInput.Update(msg)
				return m, cmd
			}
		} else {
			// Not in special mode - handle normal commands
			switch msg.String() {
			case "ctrl+c", "q":
				m.Player.Stop()
				return m, tea.Quit
			
			case "r":
				// Toggle repeat mode
				mode := m.Player.CycleRepeatMode()
				modeNames := map[player.PlaybackMode]string{
					player.RepeatNone: "Repeat: Off",
					player.RepeatOne:  "Repeat: One",
					player.RepeatAll:  "Repeat: All",
				}
				m.ErrorMsg = modeNames[mode] // Use error message area to show mode change
				return m, nil
				
			case "s":
				// Toggle shuffle mode
				m.Player.ToggleShuffle()
				if m.Player.Queue.ShuffleMode {
					m.ErrorMsg = "Shuffle: On"
				} else {
					m.ErrorMsg = "Shuffle: Off"
				}
				return m, nil
				
			case "n":
				// Play next track
				m.ErrorMsg = "" // Clear previous errors
				if err := m.Player.PlayNext(); err != nil {
					m.ErrorMsg = "Error playing next track: " + err.Error()
				}
				return m, ProgressTickCmd()
				
			case "b":
				// Play previous track
				m.ErrorMsg = "" // Clear previous errors
				if err := m.Player.PlayPrevious(); err != nil {
					m.ErrorMsg = "Error playing previous track: " + err.Error()
				}
				return m, ProgressTickCmd()
				
			case "p":
				// Toggle between tracks and playlists views
				if m.ViewMode == ViewTracks {
					m.ViewMode = ViewPlaylists
					m.ActiveList = &m.PlaylistList
					
					// If we haven't loaded playlists yet, load them now
					if len(m.Playlists) == 0 {
						m.IsLoading = true
						return m, tea.Batch(
							m.Spinner.Tick,
							GetPlaylistsCmd(m.Api),
						)
					}
				} else {
					m.ViewMode = ViewTracks
					m.ActiveList = &m.TrackList
				}
				return m, nil
				
			case "R":
				// Enter reset mode to confirm cookie reset
				m.ResetMode = true
				return m, nil
			
			case "/":
				m.SearchMode = true
				m.SearchInput.Focus()
				return m, nil
			
			case " ":
				if m.Player.IsPlaying || (!m.Player.IsPlaying && m.Player.Queue.GetCurrentTrack() != nil) {
					m.Player.TogglePause()
					if m.Player.IsPlaying {
						return m, ProgressTickCmd()
					}
				}
				return m, nil
			
			case "enter":
				if m.ActiveList.Items() == nil || len(m.ActiveList.Items()) == 0 {
					return m, nil
				}
				
				m.ErrorMsg = "" // Clear previous errors
				
				if m.ViewMode == ViewTracks {
					// Handle track selection
					selectedItem, ok := m.ActiveList.SelectedItem().(api.Track)
					if !ok {
						return m, nil
					}
					
					// Update the queue with the selected track and all following tracks
					// First, get all tracks from the current list
					allTracks := make([]api.Track, len(m.TrackList.Items()))
					for i, item := range m.TrackList.Items() {
						if track, ok := item.(api.Track); ok {
							allTracks[i] = track
						}
					}
					
					// Set the queue to all tracks, starting from the selected one
					selectedIndex := m.TrackList.Index()
					m.Player.Queue.Clear()
					m.Player.Queue.AddTracks(allTracks[selectedIndex:])
					
					// Add tracks before the selected one to the end if repeat all is enabled
					if m.Player.Queue.RepeatMode == player.RepeatAll && selectedIndex > 0 {
						m.Player.Queue.AddTracks(allTracks[:selectedIndex])
					}
					
					// Play the first track in the queue (which is the selected one)
					m.IsLoading = true
					
					return m, tea.Batch(
						m.Spinner.Tick,
						GetStreamURLCmd(m.Api, selectedItem.ID),
					)
				} else if m.ViewMode == ViewPlaylists {
					// Handle playlist selection
					selectedItem, ok := m.ActiveList.SelectedItem().(api.Playlist)
					if !ok {
						return m, nil
					}
					
					// Load tracks from the selected playlist
					m.IsLoading = true
					return m, tea.Batch(
						m.Spinner.Tick,
						GetPlaylistTracksCmd(m.Api, selectedItem.ID),
					)
				}
			}
		}
		
	case searchResultMsg:
		m.IsLoading = false
		
		if msg.err != nil {
			m.ErrorMsg = "Search error: " + msg.err.Error()
			m.SearchResults = 0
			return m, nil
		}
		
		if len(msg.tracks) == 0 {
			m.ErrorMsg = "No results found for: " + m.SearchInput.Value()
			m.SearchResults = 0
			return m, nil
		}
		
		// Convert tracks to list items
		items := make([]list.Item, len(msg.tracks))
		for i, track := range msg.tracks {
			items[i] = track
		}
		
		// Switch to tracks view
		m.ViewMode = ViewTracks
		m.ActiveList = &m.TrackList
		m.TrackList.SetItems(items)
		m.SearchInput.SetValue("")
		m.SearchResults = len(msg.tracks)
		return m, nil
		
	case playlistsResultMsg:
		m.IsLoading = false
		
		if msg.err != nil {
			m.ErrorMsg = "Error fetching playlists: " + msg.err.Error()
			return m, nil
		}
		
		if len(msg.playlists) == 0 {
			m.ErrorMsg = "No playlists found"
			return m, nil
		}
		
		// Store playlists
		m.Playlists = msg.playlists
		
		// Convert playlists to list items
		items := make([]list.Item, len(msg.playlists))
		for i, playlist := range msg.playlists {
			items[i] = playlist
		}
		
		// Update the playlist list
		m.PlaylistList.SetItems(items)
		return m, nil
		
	case playlistTracksResultMsg:
		m.IsLoading = false
		
		if msg.err != nil {
			m.ErrorMsg = "Error fetching playlist tracks: " + msg.err.Error()
			return m, nil
		}
		
		if len(msg.tracks) == 0 {
			m.ErrorMsg = "No tracks found in playlist"
			return m, nil
		}
		
		// Convert tracks to list items
		items := make([]list.Item, len(msg.tracks))
		for i, track := range msg.tracks {
			items[i] = track
		}
		
		// Switch to tracks view
		m.ViewMode = ViewTracks
		m.ActiveList = &m.TrackList
		m.TrackList.SetItems(items)
		m.SearchResults = len(msg.tracks)
		
		// Update error message to show success
		selectedPlaylist, ok := m.PlaylistList.SelectedItem().(api.Playlist)
		if ok {
			m.ErrorMsg = "Loaded " + selectedPlaylist.PlaylistTitle + " with " + 
				fmt.Sprintf("%d", m.SearchResults) + " tracks"
		} else {
			m.ErrorMsg = "Loaded playlist with " + fmt.Sprintf("%d", m.SearchResults) + " tracks"
		}
		
		return m, nil
		
	case streamURLMsg:
		m.IsLoading = false
		
		if msg.err != nil {
			m.ErrorMsg = "Error getting stream: " + msg.err.Error()
			return m, nil
		}
		
		// Get the current track from the queue
		currentTrack := m.Player.Queue.GetCurrentTrack()
		if currentTrack == nil {
			m.ErrorMsg = "Error: No track in queue"
			return m, nil
		}
		
		// Play the track
		err := m.Player.Play(msg.url, currentTrack.Duration)
		if err != nil {
			m.ErrorMsg = "Error playing track: " + err.Error()
			return m, nil
		}
		
		// Update current track info
		m.CurrentTrack = *currentTrack
		
		// Important! Update duration with the real duration from the player
		if m.Player.Duration > 0 && m.Player.Duration != m.CurrentTrack.Duration {
			updatedTrack := m.CurrentTrack
			updatedTrack.Duration = m.Player.Duration
			m.CurrentTrack = updatedTrack
			
			// Also update the track in the queue
			for i, track := range m.Player.Queue.Tracks {
				if track.ID == m.CurrentTrack.ID {
					m.Player.Queue.Tracks[i].Duration = m.Player.Duration
					break
				}
			}
		}
		
		return m, ProgressTickCmd()
		
	case cookieResetMsg:
		m.IsLoading = false
		m.ResetMode = false
		
		if msg.err != nil {
			m.ErrorMsg = "Error resetting cookies: " + msg.err.Error()
			return m, nil
		}
		
		m.LoginMode = true
		return m, nil
		
	case progressMsg:
		if m.Player.IsPlaying {
			m.Player.CurrentPos++
			
			if m.Player.CurrentPos >= m.Player.Duration {
				// The track has ended
				m.Player.CurrentPos = 0
				
				// Try to play the next track automatically
				if nextTrack, ok := m.Player.Queue.NextTrack(); ok && nextTrack != nil {
					// Get stream URL and play
					go func() {
						url, err := m.Api.GetStreamURL(nextTrack.ID)
						if err == nil {
							m.Player.Play(url, nextTrack.Duration)
							
							// Update current track info
							m.CurrentTrack = *nextTrack
						}
					}()
				} else {
					m.Player.IsPlaying = false
				}
			}
			
			if m.Player.IsPlaying {
				return m, ProgressTickCmd()
			}
		}
		return m, nil
		
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		
		// Update list sizes more conservatively
		listWidth := msg.Width - 6  // Account for borders and padding
		listHeight := msg.Height - 12  // Reserve space for other UI elements
		
		// Ensure minimum sizes
		if listWidth < 20 {
			listWidth = 20
		}
		if listHeight < 5 {
			listHeight = 5
		}
		
		// Update both lists using SetSize instead of separate Width/Height calls
		m.TrackList.SetSize(listWidth, listHeight)
		m.PlaylistList.SetSize(listWidth, listHeight)
		
		// Update progress bar width
		progressWidth := msg.Width - 10
		if progressWidth < 10 {
			progressWidth = 10
		}
		m.Progress.Width = progressWidth
		
		return m, nil
		
	case spinner.TickMsg:
		var spinnerCmd tea.Cmd
		m.Spinner, spinnerCmd = m.Spinner.Update(msg)
		if m.IsLoading {
			cmds = append(cmds, spinnerCmd)
		}
	}
	
	// Handle list and input updates
	if m.SearchMode {
		m.SearchInput, cmd = m.SearchInput.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		// Update the active list
		if m.ActiveList != nil {
			*m.ActiveList, cmd = m.ActiveList.Update(msg)
			cmds = append(cmds, cmd)
		}
	}
	
	return m, tea.Batch(cmds...)
}
