//nolint:gomnd
package prettylogger

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	reset   = "\033[0m"
	red     = "\033[0;31m"
	green   = "\033[0;32m"
	yellow  = "\033[0;33m"
	blue    = "\033[0;34m"
	magenta = "\033[0;35m"
	cyan    = "\033[0;36m"
)

// colorize returns the string s wrapped in the appropriate color code, followed by a reset.
func colorize(s, color string) string {
	if !strings.HasPrefix(color, "\033[0;3") {
		panic("color must be a valid ANSI color code")
	}

	return color + s + reset
}

// colorizeStatus returns the status code string wrapped in an appropriate color code.
func colorizeStatus(statusCode int) string {
	var color string
	switch {
	case statusCode >= 200 && statusCode < 300:
		color = green
	case statusCode >= 300 && statusCode < 400:
		color = cyan
	case statusCode >= 400 && statusCode < 500:
		color = red
	case statusCode >= 500:
		color = magenta
	default:
		color = yellow // Unknown or informational status codes
	}

	return colorize(strconv.Itoa(statusCode), color)
}

// fitString modifies a string to fit a fixed width by padding or trimming it.
// If the string is shorter than the desired length, it will be padded on the specified side (left or right).
// If the string is longer than the maxAllowedLength, it will be truncated with an ellipsis in the middle.
func fitString(s string, desiredLength int, padLeft bool, maxAllowedLength int) string {
	if maxAllowedLength > 0 && len(s) > maxAllowedLength {
		// Adjusted calculation to ensure symmetry around the ellipsis
		firstHalfLength := (maxAllowedLength) / 2
		secondHalfLength := firstHalfLength + (maxAllowedLength)%2 // Add 1 if odd to make the second half potentially longer

		return s[:firstHalfLength] + "..." + s[len(s)-secondHalfLength:]
	}

	currentLength := len(s)
	if currentLength < desiredLength {
		padding := strings.Repeat(" ", desiredLength-currentLength)
		if padLeft {
			return padding + s
		}

		return s + padding
	}

	return s
}

// formatPath uses fitString to ensure consistent log formatting.
func formatPath(path string) string {
	if path == "" {
		path = "/"
	}

	return fitString(path, 40, false, 37) // Truncate with ellipsis if longer than 37 characters
}

// getBytesIn returns the number of bytes in the request.
func getBytesIn(req *http.Request) int {
	bytesIn, _ := strconv.Atoi(req.Header.Get(echo.HeaderContentLength))

	return bytesIn
}

// formatBytes formats the byte count to a human readable format (b, Kb, Mb, Gb).

func formatBytes(bytes int) string {
	var unit string
	var value float64

	switch {
	case bytes >= 1<<30:
		unit = "Gb"
		value = float64(bytes) / (1 << 30)
	case bytes >= 1<<20:
		unit = "Mb"
		value = float64(bytes) / (1 << 20)
	case bytes >= 1<<10:
		unit = "Kb"
		value = float64(bytes) / (1 << 10)
	default:
		unit = "b"
		value = float64(bytes)
	}

	return fmt.Sprintf("%.2f%s", value, unit)
}

// Logger is a middleware that logs the request and response in a pretty format.
func Logger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		res := c.Response()
		start := time.Now()
		err := next(c)
		if err != nil {
			c.Error(err)
		}

		now := time.Now().Format("15:04:05")
		duration := time.Since(start)
		durationMs := fmt.Sprintf("%dms", duration.Milliseconds())
		method := colorize(fitString(req.Method, 7, false, 0), yellow)
		path := formatPath(req.URL.Path)
		status := colorizeStatus(res.Status)
		bytesIn := "In: " + colorize(fitString(formatBytes(getBytesIn(req)), 9, true, 0), magenta)
		bytesOut := "Out: " + colorize(fitString(formatBytes(int(res.Size)), 9, true, 0), cyan)
		durationColored := colorize(fitString(durationMs, 7, true, 0), blue)

		// Format and print the log message
		logMessage := fmt.Sprintf(
			"%s %s → %s (%s) %s [ %s | %s ]",
			now,             // Timestamp
			method,          // Method with color and fit
			path,            // Path
			status,          // Status code with color
			durationColored, // Duration in ms with color and fit
			bytesIn,         // Bytes in
			bytesOut,        // Bytes out
		)

		log.Println(logMessage)

		return nil
	}
}
