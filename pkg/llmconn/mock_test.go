// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Replaced calls to the non-existent 'aeiou.RobustParse' with the correct V3 'aeiou.Parse' function.
// filename: pkg/llmconn/mock_test.go
// nlines: 75
// risk_rating: LOW

package llmconn

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
)

func TestMockConn_ScenarioPlayback(t *testing.T) {
	ctx := context.Background()
	dummyEnvelope := &aeiou.Envelope{}

	t.Run("Plays back a successful scenario", func(t *testing.T) {
		mock := NewMock(t,
			Continue("First turn"),
			Done("Second turn"),
		)
		_, err := mock.Converse(ctx, dummyEnvelope)
		if err != nil {
			t.Fatalf("Expected no error on turn 1, got %v", err)
		}
		_, err = mock.Converse(ctx, dummyEnvelope)
		if err != nil {
			t.Fatalf("Expected no error on turn 2, got %v", err)
		}
	})

	t.Run("Returns error at the end of the scenario", func(t *testing.T) {
		mock := NewMock(t, Done("Single turn"))
		_, err := mock.Converse(ctx, dummyEnvelope) // First call is fine
		if err != nil {
			t.Fatalf("Expected no error on turn 1, got %v", err)
		}
		_, err = mock.Converse(ctx, dummyEnvelope) // Second call should fail
		if err == nil {
			t.Fatal("Expected an error when calling Converse beyond the scenario length, but got nil")
		}
	})

	t.Run("Returns a predefined error", func(t *testing.T) {
		expectedErr := errors.New("provider failed")
		mock := NewMock(t, Error(expectedErr))
		_, err := mock.Converse(ctx, dummyEnvelope)
		if !errors.Is(err, expectedErr) {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestMockConn_ScenarioHelpers(t *testing.T) {
	t.Run("Continue helper creates valid envelope", func(t *testing.T) {
		turn := Continue("test message")
		if turn.Err != nil {
			t.Fatal("Continue() helper returned an error")
		}
		env, _, err := aeiou.Parse(strings.NewReader(turn.Response.TextContent))
		if err != nil {
			t.Fatalf("Failed to parse envelope from Continue(): %v", err)
		}
		if !strings.Contains(env.Actions, "emit \"test message\"") {
			t.Error("Envelope from Continue() does not contain the emitted message")
		}
	})

	t.Run("Done helper creates valid envelope", func(t *testing.T) {
		turn := Done("final message")
		if turn.Err != nil {
			t.Fatal("Done() helper returned an error")
		}
		env, _, err := aeiou.Parse(strings.NewReader(turn.Response.TextContent))
		if err != nil {
			t.Fatalf("Failed to parse envelope from Done(): %v", err)
		}
		if !strings.Contains(env.Actions, "emit \"final message\"") {
			t.Error("Envelope from Done() does not contain the emitted message")
		}
	})
}
