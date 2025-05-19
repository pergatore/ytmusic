package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	
	"ytmusic/internal/utils"
)

// loadCookies loads cookies from the config file
func (api *YouTubeMusicAPI) loadCookies() {
	cookiePath := filepath.Join(api.configPath, "cookies.json")
	
	if _, err := os.Stat(cookiePath); os.IsNotExist(err) {
		api.LogDebug("No cookies file found at %s", cookiePath)
		return
	}

	data, err := os.ReadFile(cookiePath)
	if err != nil {
		api.LogDebug("Error reading cookies file: %v", err)
		return
	}

	var cookies []*http.Cookie
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		api.LogDebug("Error unmarshalling cookies: %v", err)
		return
	}

	// Check if we have the required cookies
	for _, cookie := range cookies {
		if cookie.Name == "__Secure-3PSID" {
			api.IsLoggedIn = true
			api.LogDebug("Found valid __Secure-3PSID cookie, setting logged in")
			break
		}
	}

	// Set cookies on client
	if api.IsLoggedIn {
		ytMusicURL, _ := url.Parse("https://music.youtube.com")
		api.client.Jar.SetCookies(ytMusicURL, cookies)
		api.LogDebug("Loaded %d cookies into client", len(cookies))
	}
}

// saveCookies saves cookies to the config file
func (api *YouTubeMusicAPI) saveCookies() error {
	ytMusicURL, _ := url.Parse("https://music.youtube.com")
	cookies := api.client.Jar.Cookies(ytMusicURL)
	
	api.LogDebug("Saving %d cookies", len(cookies))
	
	data, err := json.Marshal(cookies)
	if err != nil {
		api.LogDebug("Error marshalling cookies: %v", err)
		return err
	}
	
	cookiePath := filepath.Join(api.configPath, "cookies.json")
	return os.WriteFile(cookiePath, data, 0644)
}

// ResetCookies removes saved cookies and resets login state
func (api *YouTubeMusicAPI) ResetCookies() error {
	api.LogDebug("Resetting cookies")
	
	// Clear cookies in the client
	api.client.Jar, _ = cookiejar.New(nil)
	api.IsLoggedIn = false
	
	// Remove the cookies file
	cookiePath := filepath.Join(api.configPath, "cookies.json")
	if _, err := os.Stat(cookiePath); !os.IsNotExist(err) {
		api.LogDebug("Removing cookies file at %s", cookiePath)
		err = os.Remove(cookiePath)
		if err != nil {
			api.LogDebug("Error removing cookies file: %v", err)
			return err
		}
	}
	
	return nil
}

// ManualLogin handles manual login with a provided cookie
func (api *YouTubeMusicAPI) ManualLogin(cookie string) error {
	if cookie == "" {
		return fmt.Errorf("no cookie provided")
	}
	
	api.LogDebug("Manual login attempt with cookie length: %d", len(cookie))
	
	// Set the cookie
	ytMusicURL, _ := url.Parse("https://music.youtube.com")
	api.client.Jar.SetCookies(ytMusicURL, []*http.Cookie{
		{
			Name:   "__Secure-3PSID",
			Value:  cookie,
			Domain: ".youtube.com",
			Path:   "/",
			Secure: true,
		},
	})
	
	api.IsLoggedIn = true
	return api.saveCookies()
}

// InitiateLogin starts the login process
func (api *YouTubeMusicAPI) InitiateLogin() error {
    api.LogDebug("Initiating login process")
    
    // Try to open browser to YouTube Music
    utils.ClearScreen()
    fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
    fmt.Println("│ Attempting to open YouTube Music in your browser...               │")
    fmt.Println("└───────────────────────────────────────────────────────────────────┘")
    
    browserOpened := utils.OpenBrowser("https://music.youtube.com")
    
    if !browserOpened {
        utils.ClearScreen()
        fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
        fmt.Println("│ Could not open browser automatically.                             │")
        fmt.Println("└───────────────────────────────────────────────────────────────────┘")
        fmt.Println("")
        fmt.Println("To get your cookie manually:")
        fmt.Println("  1. Open https://music.youtube.com in your browser")
        fmt.Println("  2. Log in if you're not already logged in")
        fmt.Println("  3. Open developer tools (F12 or right-click > Inspect)")
        fmt.Println("  4. Go to Application/Storage tab > Cookies > music.youtube.com")
        fmt.Println("  5. Find the '__Secure-3PSID' cookie and copy its value")
    } else {
        utils.ClearScreen()
        fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
        fmt.Println("│ Browser opened to YouTube Music.                                  │")
        fmt.Println("└───────────────────────────────────────────────────────────────────┘")
        fmt.Println("")
        fmt.Println("Please follow these steps:")
        fmt.Println("  1. Log in if you're not already logged in")
        fmt.Println("  2. Open developer tools (F12 or right-click > Inspect)")
        fmt.Println("  3. Go to Application/Storage tab > Cookies > music.youtube.com")
        fmt.Println("  4. Find the '__Secure-3PSID' cookie and copy its value")
    }
    
    fmt.Println("")
    fmt.Println("IMPORTANT: Make sure you're getting the cookie from music.youtube.com,")
    fmt.Println("not from google.com. The correct domain is .youtube.com")
    fmt.Println("")
    fmt.Println("┌───────────────────────────────────────────────────────────────────┐")
    fmt.Print("│ Paste the cookie value here: ")
	
    var cookie string
    fmt.Scanln(&cookie)
    
    if cookie == "" {
        api.LogDebug("No cookie provided during login")
        return fmt.Errorf("no cookie provided")
    }
    
    api.LogDebug("Received cookie input with length: %d", len(cookie))
    
    fmt.Println("└─────────────────────────────────────────────────────────┘")
    
    // Set the cookie
    ytMusicURL, _ := url.Parse("https://music.youtube.com")
    api.client.Jar.SetCookies(ytMusicURL, []*http.Cookie{
        {
            Name:   "__Secure-3PSID",
            Value:  cookie,
            Domain: ".youtube.com",
            Path:   "/",
            Secure: true,
        },
    })
    
    api.IsLoggedIn = true
    utils.ClearScreen()
    fmt.Println("┌─────────────────────────────────────────────────────────┐")
    fmt.Println("│ Login successful! Press any key to continue.            │")
    fmt.Println("└─────────────────────────────────────────────────────────┘")
    return api.saveCookies()
}
