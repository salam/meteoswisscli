package output

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// InteractiveFrame holds a single radar frame for interactive viewing.
type InteractiveFrame struct {
	Timestamp string
	Grid      *RadarGrid
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
		fmt.Printf("Precipitation Radar — %s  [%d/%d]\r\n", f.Timestamp, idx+1, len(frames))
		fmt.Printf("← prev  → next  q quit\r\n\r\n")
		// In raw mode, we need \r\n instead of \n
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
				fmt.Print("\033[2J\033[H") // clear screen on exit
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
