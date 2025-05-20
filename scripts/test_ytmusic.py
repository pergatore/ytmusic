#!/usr/bin/env python3
"""
Simple test script to debug ytmusicapi authentication and basic functionality
"""

import logging
import os
from ytmusicapi import YTMusic

def test_ytmusicapi():
    """Test ytmusicapi step by step"""
    
    logging.basicConfig(level=logging.DEBUG, format='%(levelname)s: %(message)s')
    
    print("=== YTMusicAPI Test ===")
    
    # Check if headers file exists
    headers_path = os.path.expanduser("~/.ytmusic/headers_auth.json")
    print(f"1. Checking headers file: {headers_path}")
    print(f"   Exists: {os.path.exists(headers_path)}")
    
    if os.path.exists(headers_path):
        # Check file size
        size = os.path.getsize(headers_path)
        print(f"   Size: {size} bytes")
        
        if size < 100:
            print("   ⚠️  File seems too small - might be empty or corrupted")
    else:
        print("   ❌ Headers file not found!")
        print("   Run: ytmusicapi browser --file ~/.ytmusic/headers_auth.json")
        return
    
    # Test initialization with headers
    print("\n2. Testing YTMusic initialization with headers...")
    try:
        ytmusic = YTMusic(headers_path)
        print("   ✅ YTMusic initialized successfully")
    except Exception as e:
        print(f"   ❌ YTMusic initialization failed: {e}")
        return
    
    # Test without authentication first
    print("\n3. Testing YTMusic without authentication...")
    try:
        ytmusic_unauth = YTMusic()
        print("   ✅ Unauthenticated YTMusic works")
        
        # Try a simple search
        search_results = ytmusic_unauth.search("test", filter="songs", limit=1)
        print(f"   ✅ Search works: found {len(search_results)} results")
    except Exception as e:
        print(f"   ❌ Unauthenticated test failed: {e}")
    
    # Test authenticated search
    print("\n4. Testing authenticated search...")
    try:
        search_results = ytmusic.search("test", filter="songs", limit=3)
        print(f"   ✅ Authenticated search works: found {len(search_results)} results")
        
        if search_results:
            first_result = search_results[0]
            print(f"   First result: {first_result.get('title', 'No title')}")
            print(f"   Result keys: {list(first_result.keys())}")
    except Exception as e:
        print(f"   ❌ Authenticated search failed: {e}")
        import traceback
        print(f"   Traceback: {traceback.format_exc()}")
    
    # Test get_library_playlists (the problem area)
    print("\n5. Testing get_library_playlists...")
    try:
        print("   Calling get_library_playlists...")
        playlists = ytmusic.get_library_playlists(limit=5)
        print(f"   Raw result type: {type(playlists)}")
        print(f"   Raw result: {playlists}")
        
        if playlists is None:
            print("   ⚠️  Returned None - you might not have any playlists")
        elif isinstance(playlists, list):
            print(f"   ✅ Returned list with {len(playlists)} items")
            if len(playlists) > 0:
                print(f"   First playlist: {playlists[0]}")
        else:
            print(f"   ❓ Unexpected return type: {type(playlists)}")
            
    except Exception as e:
        print(f"   ❌ get_library_playlists failed: {e}")
        import traceback
        print(f"   Traceback: {traceback.format_exc()}")
    
    # Test alternative methods
    print("\n6. Testing alternative methods...")
    
    # Try get_liked_songs
    try:
        print("   Testing get_liked_songs...")
        liked = ytmusic.get_liked_songs(limit=1)
        print(f"   Liked songs type: {type(liked)}")
        print(f"   Liked songs result: {liked}")
    except Exception as e:
        print(f"   get_liked_songs failed: {e}")
    
    # Try get_history
    try:
        print("   Testing get_history...")
        history = ytmusic.get_history()
        print(f"   History type: {type(history)}")
        if history:
            print(f"   History length: {len(history) if isinstance(history, list) else 'Not a list'}")
    except Exception as e:
        print(f"   get_history failed: {e}")
    
    print("\n=== Test Complete ===")

if __name__ == "__main__":
    test_ytmusicapi()
