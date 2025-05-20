package player

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Player handles music playback
type Player struct {
	cmd          *exec.Cmd
	Queue        *Queue
	IsPlaying    bool
	CurrentPos   int
	Duration     int
	logger       *log.Logger
	nextCallback func() // Callback for when a track ends
}

// NewPlayer creates a new Player instance
func NewPlayer(debugMode bool) *Player {
	var logger *log.Logger
	if debugMode {
		configDir, _ := os.UserHomeDir()
		logPath := filepath.Join(configDir, ".ytmusic", "logs")
		logFile := filepath.Join(logPath, fmt.Sprintf("player_%s.log", time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Error opening player log file: %v\n", err)
		} else {
			logger = log.New(f, "Player: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}
	
	p := &Player{
		IsPlaying:  false,
		CurrentPos: 0,
		Duration:   0,
		logger:     logger,
	}
	
	// Create queue with logging function
	p.Queue = NewQueue(p.LogDebug)
	
	return p
}

// LogDebug logs messages if in debug mode
func (p *Player) LogDebug(format string, v ...interface{}) {
	if p.logger != nil {
		p.logger.Printf(format, v...)
	}
}

// SetNextCallback sets a callback to be called when a track ends
func (p *Player) SetNextCallback(callback func()) {
	p.nextCallback = callback
}

// Play starts playback of a URL
func (p *Player) Play(url string, duration int) error {
	if p.IsPlaying {
		p.Stop()
	}
	
	p.LogDebug("Playing URL: %s, initial duration: %d", url, duration)
	
	// Use yt-dlp to get the actual duration
	p.LogDebug("Trying to get accurate duration with yt-dlp")
	cmdGetDuration := exec.Command("yt-dlp", "--get-duration", url)
	output, err := cmdGetDuration.Output()
	if err == nil {
		durationStr := strings.TrimSpace(string(output))
		p.LogDebug("Got duration string from yt-dlp: %s", durationStr)
		
		// Parse duration like "3:45" or "1:23:45"
		parts := strings.Split(durationStr, ":")
		newDuration := 0
		
		if len(parts) == 2 {
			// MM:SS format
			minutes, _ := strconv.Atoi(parts[0])
			seconds, _ := strconv.Atoi(parts[1])
			newDuration = minutes*60 + seconds
		} else if len(parts) == 3 {
			// HH:MM:SS format
			hours, _ := strconv.Atoi(parts[0])
			minutes, _ := strconv.Atoi(parts[1])
			seconds, _ := strconv.Atoi(parts[2])
			newDuration = hours*3600 + minutes*60 + seconds
		}
		
		if newDuration > 0 {
			p.LogDebug("Setting new duration: %d seconds (was %d seconds)", newDuration, duration)
			duration = newDuration
		}
	} else {
		p.LogDebug("Failed to get duration with yt-dlp: %v", err)
	}
	
	// Now play with mpv
	p.cmd = exec.Command("mpv", "--no-video", "--no-terminal", url)
	err = p.cmd.Start()
	if err != nil {
		p.LogDebug("Error starting mpv: %v", err)
		return err
	}
	
	p.IsPlaying = true
	p.CurrentPos = 0
	p.Duration = duration
	
	// Start a goroutine to monitor playback end
	go p.monitorPlayback()
	
	return nil
}

// monitorPlayback waits for the current track to end
func (p *Player) monitorPlayback() {
	if p.cmd == nil || p.cmd.Process == nil {
		return
	}
	
	// Wait for the process to finish
	p.cmd.Wait()
	
	// Only proceed if the track actually finished (not stopped manually)
	if p.IsPlaying && p.CurrentPos >= p.Duration-1 {
		p.LogDebug("Track finished naturally, advancing to next")
		p.IsPlaying = false
		
		// Call the next callback if set
		if p.nextCallback != nil {
			p.nextCallback()
		}
	} else {
		p.LogDebug("Track was stopped manually or still playing")
	}
}

// Stop stops the current playback
func (p *Player) Stop() {
	p.LogDebug("Stopping playback")
	if p.IsPlaying && p.cmd != nil && p.cmd.Process != nil {
		p.cmd.Process.Kill()
		p.cmd.Wait()
	}
	p.IsPlaying = false
}

// TogglePause toggles the pause state of the player
func (p *Player) TogglePause() {
	p.LogDebug("Toggling pause state, current state: %v", p.IsPlaying)
	if p.cmd != nil && p.cmd.Process != nil {
		// Send SIGTSTP to pause/unpause mpv
		// Note: This is a simplified approach, ideally you'd use an mpv IPC socket
		if runtime.GOOS != "windows" {
			if p.IsPlaying {
				exec.Command("kill", "-SIGTSTP", fmt.Sprintf("%d", p.cmd.Process.Pid)).Run()
			} else {
				exec.Command("kill", "-SIGCONT", fmt.Sprintf("%d", p.cmd.Process.Pid)).Run()
			}
		}
	}
	
	p.IsPlaying = !p.IsPlaying
}

// PlayTrack plays a specific track from the queue
func (p *Player) PlayTrack(index int) error {
	if !p.Queue.PlayTrack(index) {
		return fmt.Errorf("invalid track index: %d", index)
	}
	
	track := p.Queue.GetCurrentTrack()
	if track == nil {
		return fmt.Errorf("no track to play")
	}
	
	// Get stream URL and play
	return p.PlayCurrentTrack()
}

// PlayCurrentTrack plays the current track in the queue
func (p *Player) PlayCurrentTrack() error {
	track := p.Queue.GetCurrentTrack()
	if track == nil {
		return fmt.Errorf("no track to play")
	}
	
	// Here you would get the stream URL from the API
	// For now, we'll use a simplified approach
	url := "https://www.youtube.com/watch?v=" + track.ID
	
	return p.Play(url, track.Duration)
}

// PlayNext plays the next track in the queue
func (p *Player) PlayNext() error {
	track, ok := p.Queue.NextTrack()
	if !ok || track == nil {
		return fmt.Errorf("no next track available")
	}
	
	// Get stream URL and play
	url := "https://www.youtube.com/watch?v=" + track.ID
	return p.Play(url, track.Duration)
}

// PlayPrevious plays the previous track in the queue
func (p *Player) PlayPrevious() error {
	track, ok := p.Queue.PreviousTrack()
	if !ok || track == nil {
		return fmt.Errorf("no previous track available")
	}
	
	// Get stream URL and play
	url := "https://www.youtube.com/watch?v=" + track.ID
	return p.Play(url, track.Duration)
}

// ToggleShuffle toggles shuffle mode
func (p *Player) ToggleShuffle() {
	p.Queue.ToggleShuffleMode()
}

// CycleRepeatMode cycles through repeat modes
func (p *Player) CycleRepeatMode() PlaybackMode {
	return p.Queue.CycleRepeatMode()
}
