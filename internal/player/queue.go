package player

import (
	"math/rand"
	"time"
	"ytmusic/internal/api"
)

// PlaybackMode represents the different playback modes
type PlaybackMode int

const (
	RepeatNone PlaybackMode = iota
	RepeatOne
	RepeatAll
)

// Queue manages tracks for playback
type Queue struct {
	Tracks       []api.Track
	CurrentIndex int
	ShuffleMode  bool
	RepeatMode   PlaybackMode
	History      []int // Keeps track of play history for navigation
	ShuffleOrder []int // Stores the shuffle order
	logger       func(format string, v ...interface{})
}

// NewQueue creates a new queue
func NewQueue(logFn func(format string, v ...interface{})) *Queue {
	return &Queue{
		Tracks:       []api.Track{},
		CurrentIndex: -1,
		ShuffleMode:  false,
		RepeatMode:   RepeatNone,
		History:      []int{},
		ShuffleOrder: []int{},
		logger:       logFn,
	}
}

// log helper function
func (q *Queue) log(format string, v ...interface{}) {
	if q.logger != nil {
		q.logger(format, v...)
	}
}

// GetCurrentTrack returns the current track or nil if queue is empty
func (q *Queue) GetCurrentTrack() *api.Track {
	if len(q.Tracks) == 0 || q.CurrentIndex < 0 || q.CurrentIndex >= len(q.Tracks) {
		return nil
	}
	return &q.Tracks[q.CurrentIndex]
}

// Clear empties the queue
func (q *Queue) Clear() {
	q.log("Clearing queue")
	q.Tracks = []api.Track{}
	q.CurrentIndex = -1
	q.History = []int{}
	q.ShuffleOrder = []int{}
}

// Add adds a track to the queue
func (q *Queue) Add(track api.Track) {
	q.log("Adding track to queue: %s - %s", track.TrackTitle, track.Artist)
	q.Tracks = append(q.Tracks, track)
	
	// Update shuffle order if shuffle is enabled
	if q.ShuffleMode {
		q.ShuffleOrder = append(q.ShuffleOrder, len(q.Tracks)-1)
		if len(q.Tracks) == 1 {
			q.CurrentIndex = 0
		}
	} else if q.CurrentIndex == -1 && len(q.Tracks) == 1 {
		// If this is the first track, set it as current
		q.CurrentIndex = 0
	}
}

// AddTracks adds multiple tracks to the queue
func (q *Queue) AddTracks(tracks []api.Track) {
	q.log("Adding %d tracks to queue", len(tracks))
	
	if len(tracks) == 0 {
		return
	}
	
	originalLength := len(q.Tracks)
	q.Tracks = append(q.Tracks, tracks...)
	
	// Update shuffle order if shuffle is enabled
	if q.ShuffleMode {
		// Generate new indices for the added tracks
		for i := originalLength; i < len(q.Tracks); i++ {
			q.ShuffleOrder = append(q.ShuffleOrder, i)
		}
		// Shuffle only the newly added tracks
		q.shuffleSegment(originalLength, len(q.Tracks)-1)
	}
	
	// If the queue was empty, set the current index
	if q.CurrentIndex == -1 {
		q.CurrentIndex = 0
	}
}

// SetTracks replaces the queue with the provided tracks
func (q *Queue) SetTracks(tracks []api.Track) {
	q.log("Setting queue to %d tracks", len(tracks))
	q.Clear()
	q.AddTracks(tracks)
}

// PlayTrack sets the current track to the specified index
func (q *Queue) PlayTrack(index int) bool {
	if index < 0 || index >= len(q.Tracks) {
		q.log("Cannot play track with index %d, out of bounds", index)
		return false
	}
	
	q.log("Playing track at index %d", index)
	
	// Add current track to history if we have one
	if q.CurrentIndex != -1 {
		q.History = append(q.History, q.CurrentIndex)
	}
	
	q.CurrentIndex = index
	return true
}

// NextTrack advances to the next track
func (q *Queue) NextTrack() (track *api.Track, ok bool) {
	if len(q.Tracks) == 0 {
		q.log("Cannot play next track, queue is empty")
		return nil, false
	}
	
	if q.CurrentIndex != -1 {
		q.History = append(q.History, q.CurrentIndex)
	}
	
	// Handle different repeat modes
	if q.RepeatMode == RepeatOne && q.CurrentIndex != -1 {
		// With repeat one, we just replay the current track
		q.log("Repeat One mode: replaying current track")
		return &q.Tracks[q.CurrentIndex], true
	}
	
	var nextIndex int
	
	if q.ShuffleMode {
		// In shuffle mode, use the shuffle order
		currentShufflePos := -1
		
		// Find the position of the current track in the shuffle order
		for i, idx := range q.ShuffleOrder {
			if idx == q.CurrentIndex {
				currentShufflePos = i
				break
			}
		}
		
		if currentShufflePos == -1 || currentShufflePos == len(q.ShuffleOrder)-1 {
			// We're at the end of the shuffle order
			if q.RepeatMode == RepeatAll {
				// Reset to beginning of shuffle order
				nextIndex = q.ShuffleOrder[0]
				q.log("Repeat All mode (shuffle): returning to first track in shuffle order")
			} else {
				// No more tracks
				q.log("End of shuffle order reached with no repeat")
				return nil, false
			}
		} else {
			// Move to the next track in shuffle order
			nextIndex = q.ShuffleOrder[currentShufflePos+1]
			q.log("Playing next track in shuffle order: %d", nextIndex)
		}
	} else {
		// Normal playback
		if q.CurrentIndex == -1 || q.CurrentIndex == len(q.Tracks)-1 {
			// We're at the end of the queue
			if q.RepeatMode == RepeatAll {
				// Loop back to the beginning
				nextIndex = 0
				q.log("Repeat All mode: returning to first track")
			} else {
				// No more tracks
				q.log("End of queue reached with no repeat")
				return nil, false
			}
		} else {
			// Move to the next track
			nextIndex = q.CurrentIndex + 1
			q.log("Playing next track: %d", nextIndex)
		}
	}
	
	q.CurrentIndex = nextIndex
	return &q.Tracks[q.CurrentIndex], true
}

// PreviousTrack goes back to the previous track
func (q *Queue) PreviousTrack() (track *api.Track, ok bool) {
	if len(q.Tracks) == 0 {
		q.log("Cannot play previous track, queue is empty")
		return nil, false
	}
	
	if len(q.History) > 0 {
		// Use history to go back
		prevIndex := q.History[len(q.History)-1]
		q.History = q.History[:len(q.History)-1]
		q.CurrentIndex = prevIndex
		q.log("Going back to previous track from history: %d", prevIndex)
		return &q.Tracks[q.CurrentIndex], true
	}
	
	// No history, try to go back in sequence
	if q.ShuffleMode {
		// In shuffle mode, going back is complex without history
		q.log("Cannot go back in shuffle mode without history")
		return &q.Tracks[q.CurrentIndex], true // Just replay the current track
	} else {
		// Normal playback
		if q.CurrentIndex <= 0 {
			if q.RepeatMode == RepeatAll {
				// Wrap around to the end
				q.CurrentIndex = len(q.Tracks) - 1
				q.log("Repeat All mode: wrapping to last track")
				return &q.Tracks[q.CurrentIndex], true
			}
			// Already at the beginning
			q.log("Already at the first track")
			return &q.Tracks[q.CurrentIndex], true
		}
		
		// Move to the previous track
		q.CurrentIndex--
		q.log("Playing previous track: %d", q.CurrentIndex)
		return &q.Tracks[q.CurrentIndex], true
	}
}

// ToggleShuffleMode toggles shuffle mode on/off
func (q *Queue) ToggleShuffleMode() {
	q.ShuffleMode = !q.ShuffleMode
	q.log("Shuffle mode toggled to: %v", q.ShuffleMode)
	
	if q.ShuffleMode {
		// Enable shuffle
		
		// Store original position
		originalTrack := q.GetCurrentTrack()
		
		// Initialize shuffle order with sequential indices
		q.ShuffleOrder = make([]int, len(q.Tracks))
		for i := range q.Tracks {
			q.ShuffleOrder[i] = i
		}
		
		// Shuffle the order
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(q.ShuffleOrder), func(i, j int) {
			q.ShuffleOrder[i], q.ShuffleOrder[j] = q.ShuffleOrder[j], q.ShuffleOrder[i]
		})
		
		// If there's a current track, make sure it stays as the current one
		if originalTrack != nil {
			// Find the current track in the shuffle order and swap it to the current position
			for i, idx := range q.ShuffleOrder {
				if idx == q.CurrentIndex {
					q.ShuffleOrder[i], q.ShuffleOrder[0] = q.ShuffleOrder[0], q.ShuffleOrder[i]
					break
				}
			}
			q.CurrentIndex = q.ShuffleOrder[0]
		}
	} else {
		// Disable shuffle - revert to sequential playback
		// Keep current track
		if q.CurrentIndex != -1 {
			track := q.GetCurrentTrack()
			
			// Find the actual index of the current track
			for i, t := range q.Tracks {
				if t.ID == track.ID {
					q.CurrentIndex = i
					break
				}
			}
		}
		
		// Clear the shuffle order
		q.ShuffleOrder = []int{}
	}
	
	// Reset history
	q.History = []int{}
}

// shuffleSegment shuffles a segment of the shuffle order
func (q *Queue) shuffleSegment(start, end int) {
	if start >= end || end >= len(q.ShuffleOrder) {
		return
	}
	
	segment := q.ShuffleOrder[start : end+1]
	
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(segment), func(i, j int) {
		segment[i], segment[j] = segment[j], segment[i]
	})
	
	// Copy back
	for i, val := range segment {
		q.ShuffleOrder[start+i] = val
	}
}

// CycleRepeatMode cycles through the repeat modes
func (q *Queue) CycleRepeatMode() PlaybackMode {
	switch q.RepeatMode {
	case RepeatNone:
		q.RepeatMode = RepeatOne
	case RepeatOne:
		q.RepeatMode = RepeatAll
	case RepeatAll:
		q.RepeatMode = RepeatNone
	}
	
	q.log("Repeat mode changed to: %d", q.RepeatMode)
	return q.RepeatMode
}

