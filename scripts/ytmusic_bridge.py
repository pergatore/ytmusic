#!/usr/bin/env python3
"""
YouTube Music API Bridge Script - Fixed Import Issues
Command-line interface for the Go application to interact with ytmusicapi
"""

import argparse
import json
import sys
import os
import logging
import subprocess
from typing import List, Dict, Optional, Any

# Add the current directory to path to import our module
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

# Try multiple ways to import ytmusicapi
ytmusicapi_available = False
YTMusic = None

# Method 1: Direct import
try:
    from ytmusicapi import YTMusic
    ytmusicapi_available = True
    logging.info("ytmusicapi imported successfully (direct)")
except ImportError as e:
    logging.warning(f"Direct import failed: {e}")

# Method 2: Try adding user site-packages to path
if not ytmusicapi_available:
    try:
        import site
        user_site = site.getusersitepackages()
        if user_site not in sys.path:
            sys.path.insert(0, user_site)
        
        from ytmusicapi import YTMusic
        ytmusicapi_available = True
        logging.info("ytmusicapi imported successfully (user site-packages)")
    except ImportError as e:
        logging.warning(f"User site-packages import failed: {e}")

# Method 3: Find ytmusicapi installation location
if not ytmusicapi_available:
    try:
        # Try to find where pip installed it
        result = subprocess.run(['pip3', 'show', 'ytmusicapi'], 
                               capture_output=True, text=True)
        if result.returncode == 0:
            lines = result.stdout.split('\n')
            location = None
            for line in lines:
                if line.startswith('Location: '):
                    location = line.replace('Location: ', '').strip()
                    break
            
            if location and location not in sys.path:
                sys.path.insert(0, location)
                from ytmusicapi import YTMusic
                ytmusicapi_available = True
                logging.info(f"ytmusicapi imported successfully (found at {location})")
    except Exception as e:
        logging.warning(f"pip show method failed: {e}")

# Final check
if not ytmusicapi_available:
    print(json.dumps({
        "success": False,
        "error": "ytmusicapi not found. Please check installation.",
        "traceback": f"Tried paths: {sys.path[:5]}...",
        "suggestions": [
            "Try: pip3 install --user ytmusicapi",
            "Or: sudo pip3 install ytmusicapi",
            "Check with: python3 -c 'import ytmusicapi; print(ytmusicapi.__file__)'"
        ]
    }))
    sys.exit(1)


class YouTubeMusicBridge:
    def __init__(self, cookie: str = None):
        """Initialize the bridge with optional cookie authentication"""
        self.ytmusic = None
        self.authenticated = False
        
        if cookie:
            try:
                # Try to use cookie-based authentication
                # Note: ytmusicapi doesn't directly support cookies, so we'll try headers auth
                self._authenticate_with_cookie(cookie)
            except Exception as e:
                logging.error(f"Cookie authentication failed: {e}")
        
        # Try to authenticate with existing headers file
        if not self.authenticated:
            self._authenticate_with_headers()
        
        # Fall back to no authentication
        if not self.authenticated:
            try:
                self.ytmusic = YTMusic()
                logging.warning("Running without authentication - limited functionality")
            except Exception as e:
                raise Exception(f"Failed to initialize YTMusic: {e}")
    
    def _authenticate_with_cookie(self, cookie: str):
        """Try to authenticate with a cookie (this is a simplified approach)"""
        # For now, we'll skip cookie auth since ytmusicapi prefers headers
        # In a real implementation, you'd convert the cookie to headers format
        pass
    
    def _authenticate_with_headers(self):
        """Try to authenticate with headers file"""
        # Try OAuth first (more stable)
        oauth_path = os.path.expanduser("~/.ytmusic/oauth_auth.json")
        if os.path.exists(oauth_path):
            try:
                self.ytmusic = YTMusic(oauth_path)
                self.authenticated = True
                logging.info(f"Authenticated using OAuth: {oauth_path}")
                return
            except Exception as e:
                logging.error(f"OAuth authentication failed: {e}")
        
        # Fall back to browser headers
        headers_path = os.path.expanduser("~/.ytmusic/headers_auth.json")
        if os.path.exists(headers_path):
            try:
                self.ytmusic = YTMusic(headers_path)
                self.authenticated = True
                logging.info(f"Authenticated using headers: {headers_path}")
            except Exception as e:
                logging.error(f"Headers authentication failed: {e}")
    
    def search_tracks(self, query: str, limit: int = 20) -> List[Dict[str, Any]]:
        """Search for tracks"""
        try:
            if not self.ytmusic:
                raise Exception("YTMusic client not initialized")
            
            logging.info(f"Searching for: {query}")
            results = self.ytmusic.search(query, filter="songs", limit=limit)
            
            tracks = []
            for item in results:
                track = self._format_track(item)
                if track:
                    tracks.append(track)
            
            logging.info(f"Found {len(tracks)} tracks")
            return tracks
        except Exception as e:
            logging.error(f"Search error: {e}")
            raise
    
    def get_playlists(self, limit: int = 25) -> List[Dict[str, Any]]:
        """Get user playlists"""
        try:
            if not self.ytmusic:
                raise Exception("YTMusic client not initialized")
            
            if not self.authenticated:
                logging.warning("Not authenticated - cannot fetch personal playlists")
                return []
            
            logging.info("Fetching library playlists...")
            playlists = self.ytmusic.get_library_playlists(limit=limit)
            
            # Debug: Check what we actually got
            logging.info(f"Raw playlists type: {type(playlists)}")
            if playlists and len(playlists) > 0:
                logging.info(f"First playlist type: {type(playlists[0])}")
                logging.info(f"First playlist keys: {list(playlists[0].keys()) if isinstance(playlists[0], dict) else 'Not a dict'}")

            formatted_playlists = []
            for i, playlist in enumerate(playlists):
                try:
                    # Handle case where playlist might not be a dict
                    if not isinstance(playlist, dict):
                        logging.warning(f"Playlist {i} is not a dict: {type(playlist)}")
                        continue
                        
                    # Extract playlist info more carefully
                    playlist_id = playlist.get('playlistId') or playlist.get('id', '')
                    title = playlist.get('title', 'Unknown Playlist')
                    description = playlist.get('description', '')
                    
                    # Handle count field which might be different types
                    count = playlist.get('count', 0)
                    if isinstance(count, str):
                        try:
                            count = int(count.replace(',', ''))
                        except:
                            count = 0
                    
                    # Handle author field more carefully
                    author = 'Unknown'
                    if 'author' in playlist:
                        author_data = playlist['author']
                        if isinstance(author_data, dict):
                            author = author_data.get('name', 'Unknown')
                        elif isinstance(author_data, str):
                            author = author_data
                    
                    formatted_playlist = {
                        'id': playlist_id,
                        'title': title,
                        'description': description,
                        'track_count': count,
                        'author': author
                    }
                    formatted_playlists.append(formatted_playlist)
                    logging.debug(f"Formatted playlist {i}: {title}")
                    
                except Exception as e:
                    logging.error(f"Error formatting playlist {i}: {e}")
                    continue
            
            logging.info(f"Found {len(formatted_playlists)} playlists")
            return formatted_playlists
        except Exception as e:
            logging.error(f"Get playlists error: {e}")
            raise
    
    def get_playlist_tracks(self, playlist_id: str, limit: int = 100) -> List[Dict[str, Any]]:
        """Get tracks from a playlist"""
        try:
            if not self.ytmusic:
                raise Exception("YTMusic client not initialized")
            
            if not self.authenticated:
                logging.warning("Not authenticated - cannot fetch playlist tracks")
                return []
            
            logging.info(f"Fetching tracks for playlist: {playlist_id}")
            
            # Handle special playlists
            if playlist_id == 'LM':  # Liked songs
                result = self.ytmusic.get_liked_songs(limit=limit)
                tracks = result.get('tracks', []) if isinstance(result, dict) else result
            else:
                result = self.ytmusic.get_playlist(playlist_id, limit=limit)
                if isinstance(result, dict):
                    tracks = result.get('tracks', [])
                else:
                    tracks = result
            
            formatted_tracks = []
            for track in tracks:
                formatted_track = self._format_track(track)
                if formatted_track:
                    formatted_tracks.append(formatted_track)
            
            logging.info(f"Found {len(formatted_tracks)} tracks")
            return formatted_tracks
        except Exception as e:
            logging.error(f"Get playlist tracks error: {e}")
            raise
    
    def get_liked_songs(self, limit: int = 100) -> List[Dict[str, Any]]:
        """Get user's liked songs"""
        try:
            if not self.ytmusic:
                raise Exception("YTMusic client not initialized")
            
            if not self.authenticated:
                logging.warning("Not authenticated - cannot fetch liked songs")
                return []
            
            logging.info("Fetching liked songs...")
            result = self.ytmusic.get_liked_songs(limit=limit)
            tracks = result.get('tracks', []) if isinstance(result, dict) else result
            
            formatted_tracks = []
            for track in tracks:
                formatted_track = self._format_track(track)
                if formatted_track:
                    formatted_tracks.append(formatted_track)
            
            logging.info(f"Found {len(formatted_tracks)} liked songs")
            return formatted_tracks
        except Exception as e:
            logging.error(f"Get liked songs error: {e}")
            raise
    
    def _format_track(self, track: Dict) -> Optional[Dict[str, Any]]:
        """Format a track with proper duration parsing"""
        try:
            # Debug: log the track structure
            logging.debug(f"Formatting track: {track.keys() if isinstance(track, dict) else type(track)}")
            
            # Handle case where track might not be a dict
            if not isinstance(track, dict):
                logging.warning(f"Track is not a dict: {type(track)}")
                return None
            
            # Extract video ID - try multiple possible fields
            video_id = None
            for id_field in ['videoId', 'id', 'videoID']:
                if id_field in track and track[id_field]:
                    video_id = track[id_field]
                    break
            
            if not video_id:
                logging.debug(f"Track missing video ID, available keys: {list(track.keys())}")
                return None
            
            # Extract title
            title = track.get('title', 'Unknown Title')
            
            # Extract artists - handle multiple possible structures
            artists = []
            if 'artists' in track and track['artists']:
                artist_list = track['artists']
                if isinstance(artist_list, list):
                    for artist in artist_list:
                        if isinstance(artist, dict):
                            artists.append(artist.get('name', ''))
                        elif isinstance(artist, str):
                            artists.append(artist)
                elif isinstance(artist_list, str):
                    artists.append(artist_list)
            elif 'artist' in track:
                artist_data = track['artist']
                if isinstance(artist_data, list):
                    for artist in artist_data:
                        if isinstance(artist, dict):
                            artists.append(artist.get('name', ''))
                        else:
                            artists.append(str(artist))
                else:
                    artists.append(str(artist_data))
            
            artist_str = ', '.join(filter(None, artists)) or 'Unknown Artist'
            
            # Parse duration
            duration_seconds = self._parse_duration(track)
            
            # Get thumbnail
            thumbnail = ""
            if 'thumbnails' in track and track['thumbnails']:
                thumbnails = track['thumbnails']
                if isinstance(thumbnails, list) and len(thumbnails) > 0:
                    thumbnail = thumbnails[0].get('url', '') if isinstance(thumbnails[0], dict) else ''
            
            formatted_track = {
                'id': video_id,
                'title': title,
                'artist': artist_str,
                'duration': duration_seconds,
                'thumbnail': thumbnail
            }
            
            logging.debug(f"Successfully formatted track: {title} - {artist_str}")
            return formatted_track
            
        except Exception as e:
            logging.error(f"Error formatting track: {e}")
            logging.debug(f"Track data: {track}")
            return None
    
    def _parse_duration(self, track: Dict) -> int:
        """Parse duration from track data"""
        try:
            # Try different duration fields
            if 'duration_seconds' in track:
                return int(track['duration_seconds'])
            
            duration_text = track.get('duration') or track.get('lengthText')
            if duration_text:
                return self._parse_duration_string(duration_text)
        except Exception:
            pass
        
        return 180  # Default fallback
    
    def _parse_duration_string(self, duration_str: str) -> int:
        """Parse duration string like '3:45' into seconds"""
        try:
            if not duration_str or not isinstance(duration_str, str):
                return 180
            
            # Clean the string
            clean_duration = ''.join(c for c in duration_str if c.isdigit() or c == ':').strip()
            
            if ':' in clean_duration:
                parts = clean_duration.split(':')
                if len(parts) == 2:  # MM:SS
                    minutes, seconds = int(parts[0]), int(parts[1])
                    return minutes * 60 + seconds
                elif len(parts) == 3:  # HH:MM:SS
                    hours, minutes, seconds = int(parts[0]), int(parts[1]), int(parts[2])
                    return hours * 3600 + minutes * 60 + seconds
            else:
                return int(clean_duration)
        except Exception:
            pass
        
        return 180


def main():
    """Main command-line interface"""
    parser = argparse.ArgumentParser(description='YouTube Music API Bridge')
    parser.add_argument('command', choices=['search', 'playlists', 'playlist_tracks', 'liked_songs'],
                       help='Command to execute')
    parser.add_argument('--query', help='Search query (for search command)')
    parser.add_argument('--playlist-id', help='Playlist ID (for playlist_tracks command)')
    parser.add_argument('--filter', default='songs', help='Search filter (default: songs)')
    parser.add_argument('--limit', type=int, default=20, help='Result limit (default: 20)')
    parser.add_argument('--cookie', help='Authentication cookie')
    parser.add_argument('--debug', action='store_true', help='Enable debug logging')
    
    args = parser.parse_args()
    
    # Set up logging
    log_level = logging.DEBUG if args.debug else logging.WARNING
    logging.basicConfig(
        level=log_level,
        format='%(asctime)s - %(levelname)s - %(message)s'
    )
    
    # Create response structure
    response = {
        "success": False,
        "error": None,
        "traceback": None
    }
    
    try:
        # Initialize the bridge
        bridge = YouTubeMusicBridge(cookie=args.cookie)
        
        # Execute the command
        if args.command == 'search':
            if not args.query:
                raise ValueError("Search query is required")
            
            tracks = bridge.search_tracks(args.query, args.limit)
            response["success"] = True
            response["tracks"] = tracks
            
        elif args.command == 'playlists':
            playlists = bridge.get_playlists(args.limit)
            response["success"] = True
            response["playlists"] = playlists
            
        elif args.command == 'playlist_tracks':
            if not args.playlist_id:
                raise ValueError("Playlist ID is required")
            
            tracks = bridge.get_playlist_tracks(args.playlist_id, args.limit)
            response["success"] = True
            response["tracks"] = tracks
            
        elif args.command == 'liked_songs':
            tracks = bridge.get_liked_songs(args.limit)
            response["success"] = True
            response["tracks"] = tracks
    
    except Exception as e:
        response["success"] = False
        response["error"] = str(e)
        response["traceback"] = str(e)
        logging.error(f"Command failed: {e}")
    
    # Output JSON response
    print(json.dumps(response, indent=2 if args.debug else None))


if __name__ == "__main__":
    main()
