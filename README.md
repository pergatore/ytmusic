<div align="center">

![YouTube Music TUI Header](https://raw.githubusercontent.com/pergatore/ytmusic/main/docs/header.svg)

[![Go](https://img.shields.io/badge/Go-1.18+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Python](https://img.shields.io/badge/Python-3.10+-3776AB?style=flat&logo=python)](https://python.org/)
[![ytmusicapi](https://img.shields.io/badge/ytmusicapi-latest-FF0000?style=flat&logo=youtube-music)](https://github.com/sigma67/ytmusicapi)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

# YouTube Music TUI

A terminal user interface for YouTube Music, built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea). This application uses a Python bridge to interact with the YouTube Music API via [ytmusicapi](https://github.com/sigma67/ytmusicapi).

## ✨ Features

- 🎵 Search and play music from YouTube Music
- 📱 Terminal-based UI with keyboard navigation  
- 🔐 Secure authentication via YouTube Music
- ⏯️ Full playback controls (play/pause/next/previous)
- 🔀 Shuffle and repeat modes
- 📋 Access your playlists and liked songs
- 🎚️ Queue management
- 🐛 Debug mode for troubleshooting

## 📋 Requirements

### System Dependencies
- **Go 1.18 or higher** - [Install Go](https://golang.org/doc/install)
- **Python 3.10+** - [Install Python](https://www.python.org/downloads/)
- **mpv** - For audio playback
- **pip** - Python package manager

### Install System Dependencies

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install mpv python3-pip

# Note: Ubuntu 22.04+ has Python 3.10+
# For older Ubuntu versions, you may need to install Python 3.10+ manually
python3 --version  # Check your Python version
```

#### macOS
```bash
# Using Homebrew
brew install mpv python3
```

#### Arch Linux
```bash
sudo pacman -S mpv python-pip
```

#### Windows
- Install [mpv](https://mpv.io/installation/)
- Install [Python](https://www.python.org/downloads/)

### Python Dependencies
```bash
# Install ytmusicapi
pip3 install ytmusicapi
```

## 🚀 Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/pergatore/ytmusic.git
   cd ytmusic
   ```

2. **Check Python version and install dependencies**
   ```bash
   # Check Python version (must be 3.10+)
   python3 --version
   
   # If you have Python 3.10+, install ytmusicapi
   pip3 install ytmusicapi
   
   # If your Python is older, you'll need to install Python 3.10+
   # Ubuntu 22.04+: sudo apt install python3.10 python3.10-pip
   # macOS: brew install python@3.10
   # Or download from: https://www.python.org/downloads/
   ```

3. **Build the application**
   ```bash
   go build -o ytmusic ./cmd/ytmusic
   ```

   Or install directly:
   ```bash
   go install ./cmd/ytmusic
   ```

## 🔐 Authentication Setup

**Important**: You need to authenticate with YouTube Music to access your playlists and use the full functionality. We recommend OAuth authentication for the most stable experience.

### Method 1: OAuth Authentication (Recommended)

OAuth provides the most stable and long-lasting authentication. It requires a one-time Google API setup.

#### Step 1: Create Google Cloud Project

1. **Go to [Google Cloud Console](https://console.cloud.google.com/)**
2. **Create a new project**:
   - Click "Select a project" → "New Project"
   - Name: `YouTube Music TUI` (or any name you prefer)
   - Click "Create"
3. **Select your project** from the dropdown

#### Step 2: Enable YouTube Data API

1. **Go to [API Library](https://console.cloud.google.com/apis/library)**
2. **Search for "YouTube Data API v3"**
3. **Click on it and press "Enable"**

#### Step 3: Create OAuth Credentials

1. **Go to [Credentials page](https://console.cloud.google.com/apis/credentials)**
2. **Click "Create Credentials" → "OAuth 2.0 Client IDs"**
3. **If prompted to configure OAuth consent screen**:
   - Choose "External" (unless you have a Google Workspace account)
   - Fill in required fields:
     - **App name**: `YouTube Music TUI`
     - **User support email**: Your email
     - **Developer contact information**: Your email
   - Save and continue through the steps
   - Add yourself as a test user in "Test users" section
4. **Create OAuth Client ID**:
   - **Application type**: Desktop application
   - **Name**: `YouTube Music TUI`
   - Click "Create"
5. **Download the JSON file**:
   - Click the download button next to your client ID
   - Save it as `client_secret.json` somewhere safe

#### Step 4: Set Up ytmusicapi OAuth

```bash
# Set environment variable to your downloaded file
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/client_secret.json"

# Or copy it to a standard location
cp /path/to/your/client_secret.json ~/.ytmusic/client_secret.json
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/.ytmusic/client_secret.json"

# Create OAuth authentication
ytmusicapi oauth --file ~/.ytmusic/oauth_auth.json
```

**Follow the prompts:**
1. A browser window will open
2. Sign in to your Google/YouTube Music account
3. Grant permission to the application
4. Copy the authorization code back to the terminal

#### Step 5: Verify OAuth Setup

```bash
# Test that OAuth authentication works
python3 scripts/ytmusic_bridge.py playlists --limit 5 --debug
```

If successful, you should see your playlists listed.

### Method 2: Browser Authentication (Alternative)

If OAuth setup is too complex, you can use browser authentication, though it may expire more frequently.

```bash
# Set up authentication using browser method
ytmusicapi browser --file ~/.ytmusic/headers_auth.json
```

#### Authentication Steps (Browser Method)

This method emulates your browser session by reusing its request headers.

1. **Open a new tab**
2. **Open the developer tools** (Ctrl-Shift-I) and select the **"Network"** tab
3. **Go to https://music.youtube.com** and ensure you are logged in
4. **Find an authenticated POST request**: The simplest way is to filter by `/browse` using the search bar of the developer tools. If you don't see the request, try scrolling down a bit or clicking on the library button in the top bar.

##### Firefox (Recommended)
1. Verify that the request looks like this:
   - **Status**: 200
   - **Method**: POST  
   - **Domain**: music.youtube.com
   - **File**: `browse?...`
2. **Copy the request headers**: Right click → Copy → Copy Request Headers

##### Chromium (Chrome/Edge)
1. Verify that the request looks like this:
   - **Status**: 200
   - **Name**: `browse?...`
2. **Click on the Name** of any matching request
3. In the **"Headers"** tab, scroll to the section **"Request headers"**
4. **Copy everything** starting from `"accept: */*"` to the end of the section

##### Complete the Setup
5. **Run the ytmusicapi setup command**:
   ```bash
   ytmusicapi browser --file ~/.ytmusic/headers_auth.json
   ```
6. **Paste the copied headers** when prompted
7. The authentication file will be saved automatically

### Verify Authentication
```bash
# Test that authentication works
python3 scripts/ytmusic_bridge.py playlists --limit 5 --debug
```

### Troubleshooting Authentication

#### OAuth Issues
- **"Access blocked"**: Make sure you added yourself as a test user in the OAuth consent screen
- **"Client secret not found"**: Check the `GOOGLE_APPLICATION_CREDENTIALS` environment variable
- **"Invalid client"**: Ensure you downloaded the correct JSON file from Google Cloud Console

#### Browser Issues  
- **Authentication expires quickly**: Try OAuth method instead
- **No playlists found**: You might not have any playlists, or authentication failed
- **Headers capture fails**: Make sure you're copying from music.youtube.com, not google.com

#### General Issues
```bash
# Check if ytmusicapi is installed
python3 -c "import ytmusicapi; print('OK')"

# Check authentication files
ls -la ~/.ytmusic/

# Test without authentication
python3 scripts/ytmusic_bridge.py search --query "test" --limit 3
```

If successful, you should see your playlists listed.

## 🎮 Usage

### Basic Usage
```bash
# Run the application
./ytmusic

# Enable debug mode (recommended for troubleshooting)
./ytmusic -debug

# Show help
./ytmusic -help
```

### Controls

#### Navigation
- `↑/↓` - Navigate up/down in lists
- `Enter` - Play selected track or open selected playlist
- `p` - Toggle between tracks and playlists view

#### Playback
- `Space` - Pause/resume playback
- `n` - Play next track
- `b` - Play previous track
- `r` - Cycle repeat modes (Off → One → All)
- `s` - Toggle shuffle mode

#### Other
- `/` - Search for music
- `Esc` - Exit search mode
- `R` - Reset authentication cookies
- `q` - Quit application

## 🏗️ Project Structure

```
ytmusic/
├── cmd/
│   └── ytmusic/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── auth.go              # Authentication handling
│   │   ├── bridge.go            # Python bridge communication
│   │   ├── client.go            # Main API client
│   │   ├── player.go            # Stream URL handling
│   │   ├── playlist.go          # Playlist data structures
│   │   └── track.go             # Track data structures
│   ├── player/
│   │   ├── player.go            # Music player (mpv interface)
│   │   └── queue.go             # Playback queue management
│   ├── ui/
│   │   ├── model.go             # TUI models and state
│   │   ├── update.go            # TUI update logic
│   │   └── view.go              # TUI rendering
│   └── utils/
│       └── utils.go             # Shared utilities
├── scripts/
│   └── ytmusic_bridge.py        # Python bridge to ytmusicapi
├── go.mod                       # Go dependencies
└── README.md                    # This file
```

## 🔧 Architecture

The application uses a hybrid Go + Python architecture:

```
Go TUI Application
       ↓
Python Bridge Script (scripts/ytmusic_bridge.py)
       ↓  
ytmusicapi Library
       ↓
YouTube Music API
```

This design allows us to:
- Use Go for the fast, responsive terminal UI
- Leverage the mature Python ytmusicapi library for API access
- Maintain separation between UI and API logic

## 🐛 Troubleshooting

### Common Issues

#### "Python bridge not available"
```bash
# Check if the script exists and is executable
ls -la scripts/ytmusic_bridge.py

# Test the bridge directly
python3 scripts/ytmusic_bridge.py search --query "test" --debug
```

#### "ytmusicapi not found" Error
```bash
# Check if ytmusicapi is installed
python3 -c "import ytmusicapi; print('OK')"

# If not found, reinstall
pip3 install --user ytmusicapi

# Or install system-wide
sudo pip3 install ytmusicapi
```

#### Authentication Issues
```bash
# Re-run authentication setup
ytmusicapi browser ~/.ytmusic/headers_auth.json

# Check if auth file exists
ls -la ~/.ytmusic/headers_auth.json

# Test authentication
python3 scripts/ytmusic_bridge.py playlists --debug
```

#### mpv Not Found
```bash
# Install mpv for your system
# Ubuntu/Debian: sudo apt install mpv
# macOS: brew install mpv
# Windows: Download from https://mpv.io/

# Test mpv
mpv --version
```

#### No Playlists Found
- Make sure you're logged into the correct YouTube Music account
- Try refreshing your browser authentication
- Some accounts may not have any created playlists

### Debug Mode

Always use debug mode when troubleshooting:
```bash
./ytmusic -debug
```

Debug logs are saved to:
- `~/.ytmusic/logs/ytmusic_YYYY-MM-DD.log`
- `~/.ytmusic/logs/player_YYYY-MM-DD.log`

### Getting Help

If you encounter issues:

1. **Check the debug logs** in `~/.ytmusic/logs/`
2. **Test the Python bridge directly**:
   ```bash
   python3 scripts/ytmusic_bridge.py search --query "test" --debug
   ```
3. **Verify all dependencies are installed**
4. **Re-run authentication setup**

## ⚠️ Important Notes

- **Rate Limiting**: YouTube Music has rate limits. Avoid making too many requests in a short time.
- **Authentication**: OAuth tokens are more stable than browser headers and rarely expire.
- **Google API Limits**: The free tier includes 10,000 YouTube API requests per day, which is plenty for personal use.
- **Browser Headers**: If using browser authentication, you may need to re-authenticate periodically if sessions expire.
- **Network**: You need an active internet connection to search and stream music.
- **Privacy**: Your authentication tokens are stored locally in `~/.ytmusic/` directory. Keep these files secure.

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 🙏 Acknowledgments

- [ytmusicapi](https://github.com/sigma67/ytmusicapi) - Unofficial API for YouTube Music
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [mpv](https://mpv.io/) - Media player
