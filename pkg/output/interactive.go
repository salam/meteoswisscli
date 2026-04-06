package output

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

// InteractiveFrame holds a single radar frame for interactive viewing.
type InteractiveFrame struct {
	Timestamp string
	Grid      *RadarGrid
}

// relativeTime formats a timestamp relative to now (e.g. "-30min", "now", "+1h").
func relativeTime(timestamp string) string {
	t, err := time.Parse("2006-01-02 15:04", timestamp)
	if err != nil {
		return ""
	}
	diff := time.Since(t)
	if diff < -30*time.Second {
		// Future
		mins := int((-diff).Minutes())
		if mins < 60 {
			return fmt.Sprintf("+%dmin", mins)
		}
		return fmt.Sprintf("+%dh%02dmin", mins/60, mins%60)
	}
	if diff < 5*time.Minute {
		return "now"
	}
	mins := int(diff.Minutes())
	if mins < 60 {
		return fmt.Sprintf("-%dmin", mins)
	}
	return fmt.Sprintf("-%dh%02dmin", mins/60, mins%60)
}

// timelineBar renders a visual timeline showing all frames and the current position.
func timelineBar(frames []InteractiveFrame, idx, width int) string {
	if len(frames) == 0 || width < 10 {
		return ""
	}
	barWidth := width - 2 // for [ and ]
	if barWidth > len(frames) {
		barWidth = len(frames)
	}

	bar := make([]byte, barWidth)
	for i := range bar {
		bar[i] = '-'
	}

	// Mark current position
	pos := idx * (barWidth - 1) / (len(frames) - 1)
	if pos < 0 {
		pos = 0
	}
	if pos >= barWidth {
		pos = barWidth - 1
	}
	bar[pos] = '#'

	return "[" + string(bar) + "]"
}

// InteractiveRadar shows radar frames with keyboard navigation.
func InteractiveRadar(frames []InteractiveFrame, width int, noColor, showBorder, showLakes bool) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames to display")
	}

	// Enter raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("enter raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	idx := len(frames) - 1 // start at most recent

	for {
		// Clear screen
		fmt.Print("\033[2J\033[H")

		f := frames[idx]
		rel := relativeTime(f.Timestamp)
		timeLabel := f.Timestamp
		if rel != "" {
			timeLabel += "  (" + rel + ")"
		}
		fmt.Printf("Precipitation Radar — %s  [%d/%d]\r\n", timeLabel, idx+1, len(frames))

		// Timeline bar
		timeline := timelineBar(frames, idx, width)
		oldest := relativeTime(frames[0].Timestamp)
		newest := relativeTime(frames[len(frames)-1].Timestamp)
		fmt.Printf("%s  %s\r\n", oldest, newest)
		fmt.Printf("%s\r\n", timeline)
		fmt.Printf("← prev  → next  q quit\r\n\r\n")

		rendered := RenderRadarASCII(f.Grid, width, noColor, showBorder, showLakes)
		for _, line := range splitLines(rendered) {
			fmt.Printf("%s\r\n", line)
		}

		// Read key
		buf := make([]byte, 3)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		if n == 1 {
			switch buf[0] {
			case 'q', 27: // q or ESC
				fmt.Print("\033[2J\033[H")
				return nil
			case 'h': // left
				if idx > 0 {
					idx--
				}
			case 'l': // right
				if idx < len(frames)-1 {
					idx++
				}
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 68: // left arrow
				if idx > 0 {
					idx--
				}
			case 67: // right arrow
				if idx < len(frames)-1 {
					idx++
				}
			}
		}
	}

	return nil
}

// splitLines splits a string into lines without the trailing newline.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
