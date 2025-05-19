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

// Styling
var (
	appStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#ff0000")).
		Padding(1, 2)

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
)

// Model is the main application model
type Model struct {
	Api          *api.YouTubeMusicAPI
	Player       *player.Player
	List         list.Model
	SearchInput  textinput.Model
	Progress     progress.Model
	Spinner      spinner.Model
	CurrentTrack api.Track
	Width        int
	Height       int
	SearchMode   bool
	LoginMode    bool
	ResetMode    bool
	IsLoading    bool
	ErrorMsg     string
	DebugMode    bool
}

// InitialModel creates the initial application model
func InitialModel(debugMode bool) Model {
	// Initialize API
	ytApi := api.NewYouTubeMusicAPI(debugMode)
	
	// Initialize list with delegate for our custom Track type
	delegate := list.NewDefaultDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "YouTube Music"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	
	// Search input
	ti := textinput.New()
	ti.Placeholder = "Search for music..."
	ti.CharLimit = 50
	ti.Width = 30
	
	// Progress bar
	p := progress.New(progress.WithDefaultGradient())
	
	// Spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	
	return Model{
		Api:         ytApi,
		Player:      player.NewPlayer(debugMode),
		List:        l,
		SearchInput: ti,
		Progress:    p,
		Spinner:     s,
		SearchMode:  false,
		LoginMode:   !ytApi.IsLoggedIn,
		ResetMode:   false,
		IsLoading:   false,
		DebugMode:   debugMode,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
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
