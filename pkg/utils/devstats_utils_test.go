package utils

import (
	"strings"
	"testing"
)

func TestParseContributionsRejectsEmptyFrames(t *testing.T) {
	_, err := parseContributions([]byte(`{"results":{"A":{"frames":[]}}}`))
	if err == nil {
		t.Fatal("expected error for empty devstats frames")
	}
	if !strings.Contains(err.Error(), "no frames") {
		t.Fatalf("expected empty frame error, got %q", err)
	}
}

func TestParseContributionsRejectsEmptyRows(t *testing.T) {
	_, err := parseContributions([]byte(`{"results":{"A":{"frames":[{"data":{"values":[[],[]]}}]}}}`))
	if err == nil {
		t.Fatal("expected error for empty devstats rows")
	}
	if !strings.Contains(err.Error(), "no contribution rows") {
		t.Fatalf("expected empty row error, got %q", err)
	}
}
