package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	
	"ytmusic/internal/api"
	"ytmusic/internal/player"
)

// ViewMode defines the different view modes for the application
type ViewMode int

const (
	ViewSearch ViewMode = iota
	ViewTracks
	ViewPlaylists
)

// Styling
var (
	appStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff0000")).
		Padding(1, 2).
		AlignHorizontal(lipgloss.Left).
		AlignVertical(lipgloss.Top)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#ff0000")).
		Bold(true).
		Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#EEEEEE")).
		Padding(0, 1)

	playingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFAA00")).
		Bold(true)
		
	resultInfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Italic(true)
		
	modeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00AAFF")).
		Bold(true)
)

// Model is the main application model
type Model struct {
	Api           *api.YouTubeMusicAPI
	Player        *player.Player
	TrackList     list.Model
	PlaylistList  list.Model
	SearchInput   textinput.Model
	Progress      progress.Model
	Spinner       spinner.Model
	CurrentTrack  api.Track
	Width         int
	Height        int
	SearchMode    bool
	LoginMode     bool
	ResetMode     bool
	IsLoading     bool
	ErrorMsg      string
	DebugMode     bool
	SearchResults int           // Number of search results
	Playlists     []api.Playlist // User playlists
	ViewMode      ViewMode       // Current view mode
	ActiveList    *list.Model    // Pointer to the currently active list
}

// InitialModel creates the initial application model
func InitialModel(debugMode bool) *Model {
	// Initialize API
	ytApi := api.NewYouTubeMusicAPI(debugMode)
	
	// Initialize list with custom delegate for better track display
	trackDelegate := list.NewDefaultDelegate()
	
	// Customize the delegate styles for better visual appearance
	trackDelegate.Styles.NormalTitle = trackDelegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)
		
	trackDelegate.Styles.NormalDesc = trackDelegate.Styles.NormalDesc.
		Foreground(lipgloss.Color("#AAAAAA"))
	
	trackDelegate.Styles.SelectedTitle = trackDelegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#ff0000")).
		Bold(true)
	
	trackDelegate.Styles.SelectedDesc = trackDelegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#ff0000"))
	
	// Initialize track list with default dimensions (will be updated on window size)
	trackList := list.New([]list.Item{}, trackDelegate, 80, 20)
	trackList.Title = "YouTube Music - Tracks"
	trackList.SetShowTitle(true)
	trackList.SetShowHelp(false)
	trackList.SetShowStatusBar(false) // Disable built-in status bar to save space
	trackList.SetFilteringEnabled(false)
	trackList.Styles.Title = titleStyle
	
	// Initialize playlist list with another delegate
	playlistDelegate := list.NewDefaultDelegate()
	playlistDelegate.Styles = trackDelegate.Styles // Reuse the same styling
	
	playlistList := list.New([]list.Item{}, playlistDelegate, 80, 20)
	playlistList.Title = "YouTube Music - Playlists"
	playlistList.SetShowTitle(true)
	playlistList.SetShowHelp(false)
	playlistList.SetShowStatusBar(false) // Disable built-in status bar
	playlistList.SetFilteringEnabled(false)
	playlistList.Styles.Title = titleStyle
	
	// Search input
	ti := textinput.New()
	ti.Placeholder = "Search for music..."
	ti.CharLimit = 50
	ti.Width = 30
	
	// Progress bar
	p := progress.New(progress.WithDefaultGradient())
	p.Width = 70 // Default width, will be updated
	
	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	
	// Player with debug mode
	musicPlayer := player.NewPlayer(debugMode)
	
	m := &Model{
		Api:           ytApi,
		Player:        musicPlayer,
		TrackList:     trackList,
		PlaylistList:  playlistList,
		SearchInput:   ti,
		Progress:      p,
		Spinner:       s,
		SearchMode:    false,
		LoginMode:     !ytApi.IsLoggedIn,
		ResetMode:     false,
		IsLoading:     false,
		DebugMode:     debugMode,
		SearchResults: 0,
		ViewMode:      ViewTracks,
		Width:         80,  // Default dimensions
		Height:        24,
	}
	
	// Set the active list to tracks by default
	m.ActiveList = &m.TrackList
	
	// Set up the next track callback
	m.Player.SetNextCallback(func() {
		// We need to send a message to the Bubble Tea program
		// This is done via a channel or using the Send method on the Program
		// For simplicity, we'll just automatically try to play the next track
		if err := m.Player.PlayNext(); err != nil {
			m.ErrorMsg = "Error playing next track: " + err.Error()
		}
	})
	
	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.Spinner.Tick,
		CheckLoginCmd(m.Api),
	)
}

// Messages
type loginStatusMsg struct {
	isLoggedIn bool
}

type searchResultMsg struct {
	tracks []api.Track
	err    error
}

type playlistsResultMsg struct {
	playlists []api.Playlist
	err       error
}

type playlistTracksResultMsg struct {
	tracks []api.Track
	err    error
}

type streamURLMsg struct {
	url string
	err error
}

type progressMsg struct{}

type cookieResetMsg struct {
	success bool
	err     error
}

// CheckLoginCmd checks if the user is logged in
func CheckLoginCmd(api *api.YouTubeMusicAPI) tea.Cmd {
	return func() tea.Msg {
		return loginStatusMsg{isLoggedIn: api.IsLoggedIn}
	}
}

// SearchCmd performs a search
func SearchCmd(api *api.YouTubeMusicAPI, query string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := api.Search(query)
		return searchResultMsg{tracks: tracks, err: err}
	}
}

// GetPlaylistsCmd fetches the user's playlists
func GetPlaylistsCmd(api *api.YouTubeMusicAPI) tea.Cmd {
	return func() tea.Msg {
		playlists, err := api.GetUserPlaylists()
		return playlistsResultMsg{playlists: playlists, err: err}
	}
}

// GetPlaylistTracksCmd fetches tracks from a playlist
func GetPlaylistTracksCmd(api *api.YouTubeMusicAPI, playlistID string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := api.GetPlaylistTracks(playlistID)
		return playlistTracksResultMsg{tracks: tracks, err: err}
	}
}

// GetStreamURLCmd gets a stream URL for a track
func GetStreamURLCmd(api *api.YouTubeMusicAPI, trackID string) tea.Cmd {
	return func() tea.Msg {
		url, err := api.GetStreamURL(trackID)
		return streamURLMsg{url: url, err: err}
	}
}

// ResetCookiesCmd resets cookies
func ResetCookiesCmd(api *api.YouTubeMusicAPI) tea.Cmd {
	return func() tea.Msg {
		err := api.ResetCookies()
		return cookieResetMsg{
			success: err == nil,
			err:     err,
		}
	}
}

// ProgressTickCmd ticks the progress bar
func ProgressTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return progressMsg{}
	})
}
