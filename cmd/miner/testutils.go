package main

import (
	"log/slog"
	"testing"

	"github.com/zakkbob/go-blockchain/internal/gossip"
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

func CreateTestNode(t *testing.T, h slog.Handler) *gossip.Node {
	t.Helper()
	return &gossip.Node{
		Addr:   ":0",
		Logger: slog.New(h),
	}
}
