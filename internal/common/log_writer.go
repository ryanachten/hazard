package common

import "strings"

// LogWriter redirects logs to a channel for custom rendering
type LogWriter struct {
	LogChannel chan string
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))

	select {
	case lw.LogChannel <- msg:
	default:
		// Drop log or handle overflow if the TUI event loop is stuck
	}

	return len(p), nil
}
