package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

var ForceJSON bool
var NoColor bool

func IsInteractive() bool {
	if ForceJSON {
		return false
	}
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if !NoColor {
		fmt.Fprintf(w, "\033[1m%s\033[0m\n", strings.Join(headers, "\t"))
	} else {
		fmt.Fprintln(w, strings.Join(headers, "\t"))
	}
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

func JSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func Section(title string) {
	fmt.Printf("\n--- %s ---\n", title)
}

func Error(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
}

func JSONError(msg string) {
	JSON(map[string]string{"error": msg})
}
