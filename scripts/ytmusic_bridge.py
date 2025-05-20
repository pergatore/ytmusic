#!/usr/bin/env python3
"""
YouTube Music API Bridge Script
Command-line interface for the Go application to interact with ytmusicapi
"""

import argparse
import json
import sys
import os
import logging
from typing import List, Dict, Optional, Any

# Add the current directory to path to import our module
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

try:
    from ytmusicapi import YTMusic
except ImportError:
    print(json.dumps({
        "success": False,
        "error": "ytmusicapi not installed. Run: pip install ytmusicapi",
        "traceback": "ImportError: No module named 'ytmusicapi'"
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
        headers_path = os.path.expanduser("~/.ytmusic/headers_auth.json")
        if os.path.exists(headers_path):
            try:
                self.ytmusic = YTMusic(headers_path)
                self.authenticated = True
                logging.info(f"Authenticated using {headers_path}")
            except Exception as e:
                logging.error(f"Headers authentication failed: {e}")
    
    def search_tracks(self, query: str, limit: int = 20) -> List[Dict[str, Any]]:
        """Search for tracks"""
        try:
            if not self.ytmusic:
                raise Exception("YTMusic client not initialized")
            
            results = self.ytmusic.search(query, filter="songs", limit=limit)
            
            tracks = []
            for item in results:
                track = self._format_track(item)
                if track:
                    tracks.append(track)
            
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
            
            playlists = self.ytmusic.get_library_playlists(limit=limit)
            
            formatted_playlists = []
            for playlist in playlists:
                formatted_playlist = {
                    'id': playlist.get('playlistId', ''),
                    'title': playlist.get('title', 'Unknown Playlist'),
                    'description': playlist.get('description', ''),
                    'track_count': playlist.get('count', 0),
                    'author': playlist.get('author', {}).get('name', 'Unknown') if playlist.get('author') else 'Unknown'
                }
                formatted_playlists.append(formatted_playlist)
            
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
            
            result = self.ytmusic.get_liked_songs(limit=limit)
            tracks = result.get('tracks', []) if isinstance(result, dict) else result
            
            formatted_tracks = []
            for track in tracks:
                formatted_track = self._format_track(track)
                if formatted_track:
                    formatted_tracks.append(formatted_track)
            
            return formatted_tracks
        except Exception as e:
            logging.error(f"Get liked songs error: {e}")
            raise
    
    def _format_track(self, track: Dict) -> Optional[Dict[str, Any]]:
        """Format a track with proper duration parsing"""
        try:
            # Extract video ID
            video_id = track.get('videoId') or track.get('id')
            if not video_id:
                return None
            
            # Extract title
            title = track.get('title', 'Unknown Title')
            
            # Extract artists
            artists = []
            if 'artists' in track and track['artists']:
                for artist in track['artists']:
                    if isinstance(artist, dict):
                        artists.append(artist.get('name', ''))
                    else:
                        artists.append(str(artist))
            
            artist_str = ', '.join(filter(None, artists)) or 'Unknown Artist'
            
            # Parse duration
            duration_seconds = self._parse_duration(track)
            
            # Get thumbnail
            thumbnail = ""
            if 'thumbnails' in track and track['thumbnails']:
                thumbnail = track['thumbnails'][0].get('url', '')
            
            return {
                'id': video_id,
                'title': title,
                'artist': artist_str,
                'duration': duration_seconds,
                'thumbnail': thumbnail
            }
        except Exception as e:
            logging.warning(f"Error formatting track: {e}")
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

