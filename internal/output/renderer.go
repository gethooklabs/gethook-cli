package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ── Palette ───────────────────────────────────────────────────────────────────

var (
	colorPrimary = lipgloss.Color("#7C3AED") // violet
	colorSuccess = lipgloss.Color("#22C55E") // green
	colorWarn    = lipgloss.Color("#F59E0B") // amber
	colorError   = lipgloss.Color("#EF4444") // red
	colorMuted   = lipgloss.Color("#6B7280") // gray
	colorWhite   = lipgloss.Color("#F9FAFB")
	colorBlue    = lipgloss.Color("#3B82F6")
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	StyleBold    = lipgloss.NewStyle().Bold(true)
	StyleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	StyleSuccess = lipgloss.NewStyle().Foreground(colorSuccess).Bold(true)
	StyleError   = lipgloss.NewStyle().Foreground(colorError).Bold(true)
	StyleWarn    = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
	StylePrimary = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
	StyleURL     = lipgloss.NewStyle().Foreground(colorBlue).Underline(true)

	StyleBanner = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	StyleSection = lipgloss.NewStyle().
			Foreground(colorMuted).
			Bold(true)
)

// ── Output helpers ────────────────────────────────────────────────────────────

func Success(msg string) {
	fmt.Fprintln(os.Stderr, StyleSuccess.Render("✓")+" "+msg)
}

func Info(msg string) {
	fmt.Fprintln(os.Stderr, StylePrimary.Render("→")+" "+msg)
}

func Warn(msg string) {
	fmt.Fprintln(os.Stderr, StyleWarn.Render("!")+" "+msg)
}

func Error(msg string) {
	fmt.Fprintln(os.Stderr, StyleError.Render("✗")+" "+msg)
}

func Muted(msg string) {
	fmt.Fprintln(os.Stderr, StyleMuted.Render(msg))
}

// ── Banner ────────────────────────────────────────────────────────────────────

func Banner(lines ...string) {
	content := strings.Join(lines, "\n")
	fmt.Fprintln(os.Stderr, StyleBanner.Render(content))
}

// ── Event log line ────────────────────────────────────────────────────────────

func EventLine(w io.Writer, t time.Time, method, eventType, status string, durationMs int) {
	ts := StyleMuted.Render(t.Format("15:04:05"))
	meth := StyleBold.Render(method)
	evType := lipgloss.NewStyle().Foreground(colorWhite).Render(eventType)
	statusStr := renderStatus(status)
	dur := ""
	if durationMs > 0 {
		dur = StyleMuted.Render(fmt.Sprintf("(%dms)", durationMs))
	}
	fmt.Fprintf(w, "  %s  %s  %s  %s  %s\n", ts, meth, evType, statusStr, dur)
}

func renderStatus(status string) string {
	switch {
	case status == "200" || status == "201" || status == "204" || status == "delivered" || status == "success":
		return StyleSuccess.Render(status)
	case strings.HasPrefix(status, "5") || status == "dead_letter" || status == "timeout" || status == "network_error":
		return StyleError.Render(status)
	case strings.HasPrefix(status, "4") || status == "http_4xx":
		return StyleWarn.Render(status)
	default:
		return StyleMuted.Render(status)
	}
}

// ── Table ─────────────────────────────────────────────────────────────────────

type TableRow []string

func PrintTable(headers []string, rows []TableRow) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	headerStyle := lipgloss.NewStyle().Foreground(colorMuted).Bold(true)
	sep := strings.Repeat("─", totalWidth(widths, len(headers)))
	fmt.Println(StyleMuted.Render(sep))

	headerCells := make([]string, len(headers))
	for i, h := range headers {
		headerCells[i] = headerStyle.Render(pad(h, widths[i]))
	}
	fmt.Println("  " + strings.Join(headerCells, "  "))
	fmt.Println(StyleMuted.Render(sep))

	for _, row := range rows {
		cells := make([]string, len(headers))
		for i := range headers {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			cells[i] = pad(val, widths[i])
		}
		fmt.Println("  " + strings.Join(cells, "  "))
	}
	fmt.Println(StyleMuted.Render(sep))
}

func totalWidth(widths []int, n int) int {
	total := (n - 1) * 2 // separators
	for _, w := range widths {
		total += w
	}
	return total + 4 // padding
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// ── JSON pretty print ─────────────────────────────────────────────────────────

func PrettyJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// ── Section header ────────────────────────────────────────────────────────────

func Section(title string) {
	fmt.Printf("\n%s\n%s\n",
		StyleSection.Render("── "+title+" "+strings.Repeat("─", max(0, 48-len(title)))),
		"",
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
