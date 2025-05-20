#!/usr/bin/env python3
"""
Fixed YouTube Music API implementation focusing on playlist and duration issues
"""

import json
import os
import logging
from typing import List, Dict, Optional, Any
from ytmusicapi import YTMusic

class YouTubeMusicClient:
    def __init__(self, auth_file: str = None):
        """
        Initialize YouTube Music client with proper authentication
        
        Args:
            auth_file: Path to headers_auth.json file
        """
        self.logger = logging.getLogger(__name__)
        self.ytmusic = None
        self.authenticated = False
        
        # Try to authenticate
        self._authenticate(auth_file)
    
    def _authenticate(self, auth_file: str = None):
        """Authenticate with YouTube Music"""
        
        # Try provided auth file
        if auth_file and os.path.exists(auth_file):
            try:
                self.ytmusic = YTMusic(auth_file)
                self.authenticated = True
                self.logger.info(f"Authenticated using {auth_file}")
                return
            except Exception as e:
                self.logger.error(f"Auth file {auth_file} failed: {e}")
        
        # Try default location
        default_auth = os.path.expanduser("~/.ytmusic/headers_auth.json")
        if os.path.exists(default_auth):
            try:
                self.ytmusic = YTMusic(default_auth)
                self.authenticated = True
                self.logger.info(f"Authenticated using {default_auth}")
                return
            except Exception as e:
                self.logger.error(f"Default auth file failed: {e}")
        
        # Initialize without auth (limited functionality)
        try:
            self.ytmusic = YTMusic()
            self.authenticated = False
            self.logger.warning("Running without authentication - limited functionality")
        except Exception as e:
            self.logger.error(f"Failed to initialize YTMusic: {e}")
            raise
    
    def get_playlists(self) -> List[Dict[str, Any]]:
        """Get user's playlists with better error handling"""
        if not self.ytmusic:
            raise Exception("YTMusic client not initialized")
        
        if not self.authenticated:
            self.logger.error("Not authenticated - cannot fetch personal playlists")
            self.logger.info("To authenticate, run: ytmusicapi headers ~/.ytmusic/headers_auth.json")
            return []
        
        try:
            # Get library playlists
            self.logger.info("Fetching library playlists...")
            playlists = self.ytmusic.get_library_playlists(limit=25)
            
            if playlists:
                self.logger.info(f"Found {len(playlists)} library playlists")
                return self._format_playlists(playlists)
            
            # If no library playlists, try to get some default ones
            self.logger.info("No library playlists found, trying other methods...")
            
            # Try getting liked songs playlist
            try:
                history = self.ytmusic.get_history()
                if history:
                    self.logger.info("Found history, creating history playlist")
                    return [{
                        'playlistId': 'HISTORY',
                        'title': 'Recently Played',
                        'description': 'Your recently played tracks',
                        'count': len(history),
                        'thumbnails': []
                    }]
            except Exception as e:
                self.logger.warning(f"Could not get history: {e}")
            
            # Try searching for popular playlists as fallback
            try:
                self.logger.info("Trying to find public playlists...")
                search_results = self.ytmusic.search("top hits", filter="playlists", limit=5)
                if search_results:
                    return self._format_search_playlists(search_results)
            except Exception as e:
                self.logger.warning(f"Could not search playlists: {e}")
            
            self.logger.warning("No playlists found")
            return []
            
        except Exception as e:
            self.logger.error(f"Error fetching playlists: {e}")
            return []
    
    def _format_playlists(self, playlists: List[Dict]) -> List[Dict[str, Any]]:
        """Format playlists for consistency"""
        formatted = []
        for playlist in playlists:
            formatted_playlist = {
                'playlistId': playlist.get('playlistId', ''),
                'title': playlist.get('title', 'Unknown Playlist'),
                'description': playlist.get('description', ''),
                'count': playlist.get('count', 0),
                'thumbnails': playlist.get('thumbnails', [])
            }
            formatted.append(formatted_playlist)
        return formatted
    
    def _format_search_playlists(self, search_results: List[Dict]) -> List[Dict[str, Any]]:
        """Format search results as playlists"""
        formatted = []
        for item in search_results:
            if item.get('resultType') == 'playlist':
                formatted_playlist = {
                    'playlistId': item.get('browseId', ''),
                    'title': item.get('title', 'Unknown Playlist'),
                    'description': item.get('description', 'Public playlist'),
                    'count': item.get('count', '?'),
                    'thumbnails': item.get('thumbnails', [])
                }
                formatted.append(formatted_playlist)
        return formatted
    
    def get_playlist_tracks(self, playlist_id: str, limit: int = 100) -> List[Dict[str, Any]]:
        """Get tracks from a specific playlist with proper error handling"""
        if not self.ytmusic:
            raise Exception("YTMusic client not initialized")
        
        if not self.authenticated:
            self.logger.error("Not authenticated - cannot fetch playlist tracks")
            return []
        
        try:
            self.logger.info(f"Fetching tracks for playlist: {playlist_id}")
            
            # Handle special playlists
            if playlist_id == 'HISTORY':
                tracks = self.ytmusic.get_history()
            elif playlist_id == 'LM':  # Liked songs
                result = self.ytmusic.get_liked_songs(limit=limit)
                tracks = result.get('tracks', []) if isinstance(result, dict) else result
            else:
                # Regular playlist - use the correct method
                result = self.ytmusic.get_playlist(playlist_id, limit=limit)
                if isinstance(result, dict):
                    tracks = result.get('tracks', [])
                else:
                    tracks = result
            
            if not tracks:
                self.logger.warning(f"No tracks found in playlist {playlist_id}")
                return []
            
            # Format tracks with proper duration extraction
            formatted_tracks = []
            for track in tracks:
                formatted_track = self._format_track(track)
                if formatted_track:
                    formatted_tracks.append(formatted_track)
            
            self.logger.info(f"Successfully fetched {len(formatted_tracks)} tracks")
            return formatted_tracks
            
        except Exception as e:
            self.logger.error(f"Error fetching playlist tracks: {e}")
            # Log more details for debugging
            self.logger.debug(f"Playlist ID: {playlist_id}")
            self.logger.debug(f"Error type: {type(e).__name__}")
            return []
    
    def _format_track(self, track: Dict) -> Optional[Dict[str, Any]]:
        """Format a track with proper duration parsing"""
        try:
            # Extract video ID
            video_id = None
            if 'videoId' in track:
                video_id = track['videoId']
            elif 'id' in track:
                video_id = track['id']
            
            if not video_id:
                self.logger.warning("Track missing video ID")
                return None
            
            # Extract title and artist
            title = track.get('title', 'Unknown Title')
            
            # Handle different artist formats
            artists = []
            if 'artists' in track and track['artists']:
                for artist in track['artists']:
                    if isinstance(artist, dict):
                        artists.append(artist.get('name', ''))
                    else:
                        artists.append(str(artist))
            elif 'artist' in track:
                if isinstance(track['artist'], list):
                    artists = [a.get('name', str(a)) if isinstance(a, dict) else str(a) 
                             for a in track['artist']]
                else:
                    artists = [str(track['artist'])]
            
            artist_str = ', '.join(filter(None, artists)) or 'Unknown Artist'
            
            # Parse duration properly
            duration_seconds = self._parse_duration(track)
            
            # Create formatted track
            formatted_track = {
                'id': video_id,
                'title': title,
                'artist': artist_str,
                'duration': duration_seconds,
                'url': f"https://www.youtube.com/watch?v={video_id}"
            }
            
            return formatted_track
            
        except Exception as e:
            self.logger.warning(f"Error formatting track: {e}")
            return None
    
    def _parse_duration(self, track: Dict) -> int:
        """Parse duration from various possible formats"""
        duration_seconds = 180  # Default fallback
        
        try:
            # Try different duration fields
            duration_text = None
            
            if 'duration' in track:
                duration_text = track['duration']
            elif 'duration_seconds' in track:
                return int(track['duration_seconds'])
            elif 'lengthText' in track:
                duration_text = track['lengthText']
            
            if duration_text:
                duration_seconds = self._parse_duration_string(duration_text)
            
        except Exception as e:
            self.logger.debug(f"Could not parse duration, using default: {e}")
        
        return duration_seconds
    
    def _parse_duration_string(self, duration_str: str) -> int:
        """Parse duration string like '3:45' into seconds"""
        try:
            if not duration_str or not isinstance(duration_str, str):
                return 180
            
            # Remove any non-numeric/colon characters and whitespace
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
                # Try to parse as just seconds
                return int(clean_duration)
                
        except Exception:
            pass
        
        return 180  # Default fallback
    
    def search_tracks(self, query: str, limit: int = 20) -> List[Dict[str, Any]]:
        """Search for tracks with improved duration handling"""
        if not self.ytmusic:
            raise Exception("YTMusic client not initialized")
        
        try:
            self.logger.info(f"Searching for: {query}")
            
            # Search for songs specifically
            results = self.ytmusic.search(query, filter="songs", limit=limit)
            
            if not results:
                self.logger.warning(f"No results found for query: {query}")
                return []
            
            # Format search results
            formatted_tracks = []
            for track in results:
                formatted_track = self._format_track(track)
                if formatted_track:
                    formatted_tracks.append(formatted_track)
            
            self.logger.info(f"Found {len(formatted_tracks)} tracks")
            return formatted_tracks
            
        except Exception as e:
            self.logger.error(f"Error searching tracks: {e}")
            return []


# Example usage and testing
if __name__ == "__main__":
    # Set up logging
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    # Initialize client
    client = YouTubeMusicClient()
    
    # Test playlist fetching
    print("=== Testing Playlist Fetching ===")
    playlists = client.get_playlists()
    for playlist in playlists[:3]:  # Show first 3
        print(f"Playlist: {playlist['title']} (ID: {playlist['playlistId']})")
    
    # Test search functionality
    print("\n=== Testing Search ===")
    tracks = client.search_tracks("gekirin", limit=5)
    for track in tracks:
        print(f"Track: {track['title']} - {track['artist']} ({track['duration']}s)")
    
    # Test playlist tracks if we have playlists
    if playlists:
        print(f"\n=== Testing Playlist Tracks ===")
        playlist_id = playlists[0]['playlistId']
        tracks = client.get_playlist_tracks(playlist_id, limit=5)
        for track in tracks:
            print(f"Track: {track['title']} - {track['artist']} ({track['duration']}s)")

