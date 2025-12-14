package main

import (
	"log/slog"
	"testing"
)

type testLogger struct {
	t *testing.T
}

func (l testLogger) Write(p []byte) (int, error) {
	l.t.Log("Log from test: '" + string(p) + "'")
	return len(p), nil
}

func CreateTestConfig(t *testing.T) config {
	return config{
		debug: true,
	}
}

func CreateTestLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(testLogger{t}, nil))
}
