package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"ytmusic/internal/ui"
	"ytmusic/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
)

// Global flag for debug mode
var debugMode bool

func main() {
	// Parse command line flags
	var showHelp bool
	flag.BoolVar(&debugMode, "debug", false, "Enable debug logging")
	flag.BoolVar(&showHelp, "help", false, "Show help information")
	flag.Parse()
	
	// Show help if requested
	if showHelp {
		fmt.Println("YouTube Music TUI")
		fmt.Println("----------------")
		fmt.Println("A terminal user interface for YouTube Music")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  ytmusic [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -debug    Enable debug logging")
		fmt.Println("  -help     Show this help message")
		fmt.Println("")
		fmt.Println("Controls:")
		fmt.Println("  q         Quit")
		fmt.Println("  l         Login (when not logged in)")
		fmt.Println("  r         Reset cookies/credentials")
		fmt.Println("  /         Search")
		fmt.Println("  Enter     Play selected track")
		fmt.Println("  Space     Pause/resume playback")
		fmt.Println("  ↑/↓       Navigate up/down")
		fmt.Println("")
		return
	}
	
	if debugMode {
		configDir, _ := os.UserHomeDir()
		logPath := filepath.Join(configDir, ".ytmusic", "logs")
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			os.MkdirAll(logPath, 0755)
		}
		
		logFile := filepath.Join(logPath, fmt.Sprintf("ytmusic_%s.log", time.Now().Format("2006-01-02")))
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Printf("Error opening log file: %v\nContinuing without logging...\n", err)
		} else {
			log.SetOutput(f)
			log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
			log.Printf("Starting YouTube Music TUI with debug mode enabled")
		}
	}
	
	// Clear terminal
	utils.ClearScreen()
	
	p := tea.NewProgram(ui.InitialModel(debugMode), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
