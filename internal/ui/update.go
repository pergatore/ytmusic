package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update updates the model based on messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case loginStatusMsg:
		m.LoginMode = !msg.isLoggedIn
		if m.LoginMode {
			return m, nil
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
				// Enter reset mode to confirm cookie reset
				m.ResetMode = true
				return m, nil
			
			case "/":
				m.SearchMode = true
				m.SearchInput.Focus()
				return m, nil
			
			case " ":
				if m.Player.IsPlaying || (!m.Player.IsPlaying && m.CurrentTrack.ID != "") {
					m.Player.TogglePause()
					if m.Player.IsPlaying {
						return m, ProgressTickCmd()
					}
				}
				return m, nil
			
			case "enter":
				if len(m.List.Items()) > 0 {
					selectedItem, ok := m.List.SelectedItem().(api.Track)
					if !ok {
						return m, nil
					}
					
					m.CurrentTrack = selectedItem
					m.IsLoading = true
					m.ErrorMsg = "" // Clear previous errors
					
					return m, tea.Batch(
						m.Spinner.Tick,
						GetStreamURLCmd(m.Api, selectedItem.ID),
					)
				}
			}
		}
		
	case searchResultMsg:
		m.IsLoading = false
		
		if msg.err != nil {
			m.ErrorMsg = "Search error: " + msg.err.Error()
			return m, nil
		}
		
		if len(msg.tracks) == 0 {
			m.ErrorMsg = "No results found for: " + m.SearchInput.Value()
			return m, nil
		}
		
		items := make([]list.Item, len(msg.tracks))
		for i, track := range msg.tracks {
			items[i] = track
		}
		
		m.List.SetItems(items)
		m.SearchInput.SetValue("")
		return m, nil
		
	case streamURLMsg:
		m.IsLoading = false
		
		if msg.err != nil {
			m.ErrorMsg = "Error getting stream: " + msg.err.Error()
			return m, nil
		}
		
		err := m.Player.Play(msg.url, m.CurrentTrack.Duration)
		if err != nil {
			m.ErrorMsg = "Error playing track: " + err.Error()
			return m, nil
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
				m.Player.CurrentPos = 0
				m.Player.IsPlaying = false
			} else {
				return m, ProgressTickCmd()
			}
		}
		return m, nil
		
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		
		m.List.SetWidth(msg.Width - 4)
		m.List.SetHeight(msg.Height - 10)
		
		m.Progress.Width = msg.Width - 10
		
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
		m.List, cmd = m.List.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}
