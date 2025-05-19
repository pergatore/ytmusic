# YouTube Music TUI

A terminal user interface for YouTube Music, built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- Search and play music from YouTube Music
- Terminal-based UI with keyboard navigation
- Login support via YouTube Music cookie
- Playback controls (play/pause)
- Debug mode for troubleshooting

## Requirements

- Go 1.18 or higher
- mpv (for playback)

## Installation

```bash
# Clone the repository
git clone https://github.com/pergatore/ytmusic.git
cd ytmusic

# Build the binary
go build -o ytmusic ./cmd/ytmusic

# Or install directly
go install ./cmd/ytmusic
```

## Usage

```bash
# Run the application
ytmusic

# Enable debug mode
ytmusic -debug

# Show help
ytmusic -help
```

## Controls

- `q` - Quit
- `l` - Login (when not logged in)
- `r` - Reset cookies/credentials
- `/` - Search
- `Enter` - Play selected track
- `Space` - Pause/resume playback
- `↑/↓` - Navigate up/down

## Login

To use this application, you need to provide a YouTube Music cookie:

1. Open YouTube Music in your browser and log in
2. Open developer tools (F12 or right-click > Inspect)
3. Go to Application/Storage tab > Cookies > music.youtube.com
4. Find the `__Secure-3PSID` cookie and copy its value
5. Paste it into the application when prompted

## Project Structure

```
ytmusic/
├── cmd/
│   └── ytmusic/
│       └── main.go           # Entry point, command-line flags
├── internal/
│   ├── api/
│   │   └── ytmusic.go        # YouTube Music API client
│   ├── player/
│   │   └── player.go         # Music player functionality
│   ├── ui/
│   │   ├── model.go          # Main TUI model
│   │   ├── update.go         # Update logic for TUI
│   │   └── view.go           # View rendering for TUI
│   └── utils/
│       └── utils.go          # Shared utility functions
└── go.mod                    # Go module file
```

## License

MIT
