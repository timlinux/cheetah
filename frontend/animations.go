// SPDX-FileCopyrightText: 2026 Tim Sutton / Kartoza
// SPDX-License-Identifier: MIT

package frontend

import (
	"time"

	"github.com/charmbracelet/harmonica"
)

// Animation configuration constants
const (
	AnimationInterval = 50 * time.Millisecond
)

// WordAnimator handles smooth spring-based animations for word transitions
type WordAnimator struct {
	// Springs for each element
	prevSpring    harmonica.Spring
	currentSpring harmonica.Spring
	nextSpring    harmonica.Spring

	// Positions (0.0 to 1.0 representing animation progress)
	PrevPos    float64
	CurrentPos float64
	NextPos    float64

	// Velocities for spring physics
	prevVel    float64
	currentVel float64
	nextVel    float64

	// Animation state
	IsAnimating bool
	frame       int
}

// NewWordAnimator creates a new animator for the reading screen word carousel
func NewWordAnimator() *WordAnimator {
	return &WordAnimator{
		// Snappy spring for smooth but quick animations
		prevSpring:    harmonica.NewSpring(harmonica.FPS(60), 8.0, 0.6),
		currentSpring: harmonica.NewSpring(harmonica.FPS(60), 7.0, 0.5),
		nextSpring:    harmonica.NewSpring(harmonica.FPS(60), 6.0, 0.6),
		// Start with everything in place
		PrevPos:     1.0,
		CurrentPos:  1.0,
		NextPos:     1.0,
		IsAnimating: false,
	}
}

// TriggerTransition starts the word animation when moving to next word
func (w *WordAnimator) TriggerTransition() {
	w.IsAnimating = true
	w.frame = 0

	// Previous word: start visible, will scroll up and fade out
	w.PrevPos = 0.0
	w.prevVel = 0.0

	// Current word: start below, will slide up into view
	w.CurrentPos = 0.0
	w.currentVel = 0.0

	// Next word: start hidden below, will fade in
	w.NextPos = 0.0
	w.nextVel = 0.0
}

// Update advances all springs by one frame
func (w *WordAnimator) Update() {
	if !w.IsAnimating {
		return
	}

	w.frame++

	// Update all springs toward target position of 1.0
	w.PrevPos, w.prevVel = w.prevSpring.Update(w.PrevPos, w.prevVel, 1.0)
	w.CurrentPos, w.currentVel = w.currentSpring.Update(w.CurrentPos, w.currentVel, 1.0)

	// Next word starts slightly delayed for stagger effect
	if w.frame > 2 {
		w.NextPos, w.nextVel = w.nextSpring.Update(w.NextPos, w.nextVel, 1.0)
	}

	// Check if animation is complete
	if w.PrevPos > 0.98 && w.CurrentPos > 0.98 && w.NextPos > 0.98 {
		if abs(w.prevVel) < 0.01 && abs(w.currentVel) < 0.01 && abs(w.nextVel) < 0.01 {
			w.IsAnimating = false
			w.PrevPos = 1.0
			w.CurrentPos = 1.0
			w.NextPos = 1.0
		}
	}
}

// GetPrevOffset returns the vertical offset for the previous word (scrolls up)
func (w *WordAnimator) GetPrevOffset() int {
	// Starts at bottom of its area (offset 2), scrolls up to final position (offset 0)
	maxOffset := 2
	return int(float64(maxOffset) * (1.0 - w.PrevPos))
}

// GetPrevOpacity returns opacity for previous word (0.0 to 1.0)
func (w *WordAnimator) GetPrevOpacity() float64 {
	// Fades from 0 to target opacity as it scrolls up
	return w.PrevPos * 0.5 // Max 50% opacity for dimmed previous word
}

// GetCurrentOffset returns the vertical offset for current word (slides up from below)
func (w *WordAnimator) GetCurrentOffset() int {
	// Starts below (offset 3), slides up to center (offset 0)
	maxOffset := 3
	return int(float64(maxOffset) * (1.0 - w.CurrentPos))
}

// GetCurrentScale returns a scale factor for the current word (grows as it enters)
func (w *WordAnimator) GetCurrentScale() float64 {
	// Starts at 0.7, grows to 1.0
	return 0.7 + (w.CurrentPos * 0.3)
}

// GetNextOffset returns the vertical offset for next word (fades in below)
func (w *WordAnimator) GetNextOffset() int {
	// Starts further below (offset 2), moves up slightly (offset 0)
	maxOffset := 2
	return int(float64(maxOffset) * (1.0 - w.NextPos))
}

// GetNextOpacity returns opacity for next word (0.0 to 1.0)
func (w *WordAnimator) GetNextOpacity() float64 {
	// Fades in to target opacity
	return w.NextPos * 0.6 // Max 60% opacity for next word preview
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetAnimationInterval returns the animation tick interval
func GetAnimationInterval() time.Duration {
	return AnimationInterval
}
