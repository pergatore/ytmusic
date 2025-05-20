#!/usr/bin/env python3
"""
YouTube Music API Bridge
A Python bridge that uses ytmusicapi to provide YouTube Music functionality
to the Go application via JSON communication.
"""

import json
import sys
import os
import argparse
from typing import Dict, List, Any, Optional
import traceback

try:
    from ytmusicapi import YTMusic
except ImportError:
    print(json.dumps({
        "error": "ytmusicapi not installed. Install with: pip install ytmusicapi"
    }))
    sys.exit(1)


class YouTubeMusicBridge:
    def __init__(self, auth_file: Optional[str] = None):
        """Initialize the YouTube Music API bridge."""
        self.ytmusic = None
        self.auth_file = auth_file
        self.is_authenticated = False
        
    def authenticate(self, headers_auth: Optional[str] = None) -> Dict[str, Any]:
        """Authenticate with YouTube Music."""
        try:
            if headers_auth:
                # Use provided headers for authentication
                self.ytmusic = YTMusic(headers_auth)
            elif self.auth_file and os.path.exists(self.auth_file):
                # Use existing auth file
                self.ytmusic = YTMusic(self.auth_file)
            else:
                # Try to authenticate without headers (may not work for all features)
                self.ytmusic = YTMusic()
            
            # Test authentication by trying to get library playlists
            try:
                self.ytmusic.get_library_playlists(limit=1)
                self.is_authenticated = True
                return {"success": True, "authenticated": True}
            except Exception as e:
                # If library access fails, we might still be able to search
                self.is_authenticated = False
                return {"success": True, "authenticated": False, "message": "Limited access - search only"}
                
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    def search(self, query: str, filter_type: str = "songs", limit: int = 20) -> Dict[str, Any]:
        """Search for music."""
        try:
            if not self.ytmusic:
                return {"success": False, "error": "Not authenticated"}
            
            results = self.ytmusic.search(query, filter=filter_type, limit=limit)
            
            tracks = []
            for item in results:
                if filter_type == "songs" and item.get('category') == 'Songs':
                    track = {
                        "id": item.get('videoId', ''),
                        "title": item.get('title', ''),
                        "artist": self._get_artist_name(item),
                        "duration": self._get_duration_seconds(item.get('duration')),
                        "thumbnail": self._get_thumbnail_url(item)
                    }
                    tracks.append(track)
            
            return {"success": True, "tracks": tracks}
            
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    def get_playlists(self, limit: int = 25) -> Dict[str, Any]:
        """Get user's library playlists."""
        try:
            if not self.ytmusic:
                return {"success": False, "error": "Not authenticated"}
            
            if not self.is_authenticated:
                return {"success": False, "error": "Authentication required for playlists"}
            
            playlists_data = self.ytmusic.get_library_playlists(limit=limit)
            
            playlists = []
            for item in playlists_data:
                playlist = {
                    "id": item.get('playlistId', ''),
                    "title": item.get('title', ''),
                    "description": item.get('description', ''),
                    "track_count": item.get('count', 0),
                    "author": item.get('author', {}).get('name', 'Unknown') if item.get('author') else 'You'
                }
                playlists.append(playlist)
            
            return {"success": True, "playlists": playlists}
            
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    def get_playlist_tracks(self, playlist_id: str, limit: int = 100) -> Dict[str, Any]:
        """Get tracks from a playlist."""
        try:
            if not self.ytmusic:
                return {"success": False, "error": "Not authenticated"}
            
            playlist_data = self.ytmusic.get_playlist(playlist_id, limit=limit)
            
            tracks = []
            for item in playlist_data.get('tracks', []):
                if item.get('videoId'):
                    track = {
                        "id": item.get('videoId', ''),
                        "title": item.get('title', ''),
                        "artist": self._get_artist_name(item),
                        "duration": self._get_duration_seconds(item.get('duration')),
                        "thumbnail": self._get_thumbnail_url(item)
                    }
                    tracks.append(track)
            
            return {"success": True, "tracks": tracks}
            
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    def get_liked_songs(self, limit: int = 100) -> Dict[str, Any]:
        """Get user's liked songs."""
        try:
            if not self.ytmusic:
                return {"success": False, "error": "Not authenticated"}
            
            if not self.is_authenticated:
                return {"success": False, "error": "Authentication required for liked songs"}
            
            liked_songs = self.ytmusic.get_liked_songs(limit=limit)
            
            tracks = []
            for item in liked_songs.get('tracks', []):
                if item.get('videoId'):
                    track = {
                        "id": item.get('videoId', ''),
                        "title": item.get('title', ''),
                        "artist": self._get_artist_name(item),
                        "duration": self._get_duration_seconds(item.get('duration')),
                        "thumbnail": self._get_thumbnail_url(item)
                    }
                    tracks.append(track)
            
            return {"success": True, "tracks": tracks}
            
        except Exception as e:
            return {"success": False, "error": str(e)}
    
    def _get_artist_name(self, item: Dict[str, Any]) -> str:
        """Extract artist name from item."""
        artists = item.get('artists', [])
        if artists and len(artists) > 0:
            return artists[0].get('name', 'Unknown Artist')
        
        # Fallback for different structures
        artist = item.get('artist')
        if isinstance(artist, list) and len(artist) > 0:
            return artist[0].get('name', 'Unknown Artist')
        elif isinstance(artist, dict):
            return artist.get('name', 'Unknown Artist')
        elif isinstance(artist, str):
            return artist
        
        return 'Unknown Artist'
    
    def _get_duration_seconds(self, duration: Optional[str]) -> int:
        """Convert duration string to seconds."""
        if not duration:
            return 180  # Default 3 minutes
        
        try:
            # Duration format is usually "MM:SS" or "H:MM:SS"
            parts = duration.split(':')
            if len(parts) == 2:
                minutes, seconds = int(parts[0]), int(parts[1])
                return minutes * 60 + seconds
            elif len(parts) == 3:
                hours, minutes, seconds = int(parts[0]), int(parts[1]), int(parts[2])
                return hours * 3600 + minutes * 60 + seconds
        except (ValueError, AttributeError):
            pass
        
        return 180  # Default fallback
    
    def _get_thumbnail_url(self, item: Dict[str, Any]) -> str:
        """Extract thumbnail URL from item."""
        thumbnails = item.get('thumbnails', [])
        if thumbnails and len(thumbnails) > 0:
            return thumbnails[0].get('url', '')
        return ''


def main():
    parser = argparse.ArgumentParser(description='YouTube Music API Bridge')
    parser.add_argument('command', help='Command to execute')
    parser.add_argument('--auth-file', help='Path to authentication file')
    parser.add_argument('--headers-auth', help='Authentication headers JSON string')
    parser.add_argument('--query', help='Search query')
    parser.add_argument('--playlist-id', help='Playlist ID')
    parser.add_argument('--filter', default='songs', help='Search filter type')
    parser.add_argument('--limit', type=int, default=20, help='Result limit')
    
    args = parser.parse_args()
    
    bridge = YouTubeMusicBridge(args.auth_file)
    
    try:
        if args.command == 'auth':
            result = bridge.authenticate(args.headers_auth)
        elif args.command == 'search':
            if not args.query:
                result = {"success": False, "error": "Query required for search"}
            else:
                bridge.authenticate(args.headers_auth)
                result = bridge.search(args.query, args.filter, args.limit)
        elif args.command == 'playlists':
            bridge.authenticate(args.headers_auth)
            result = bridge.get_playlists(args.limit)
        elif args.command == 'playlist_tracks':
            if not args.playlist_id:
                result = {"success": False, "error": "Playlist ID required"}
            else:
                bridge.authenticate(args.headers_auth)
                result = bridge.get_playlist_tracks(args.playlist_id, args.limit)
        elif args.command == 'liked_songs':
            bridge.authenticate(args.headers_auth)
            result = bridge.get_liked_songs(args.limit)
        else:
            result = {"success": False, "error": f"Unknown command: {args.command}"}
        
        print(json.dumps(result))
        
    except Exception as e:
        error_result = {
            "success": False, 
            "error": str(e),
            "traceback": traceback.format_exc()
        }
        print(json.dumps(error_result))


if __name__ == '__main__':
    main()

