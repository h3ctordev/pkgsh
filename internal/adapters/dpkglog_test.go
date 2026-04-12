package adapters

import (
	"strings"
	"testing"
	"time"
)

func TestParseDpkgLog_ParsesStatusInstalled(t *testing.T) {
	content := "2024-01-15 10:30:45 status installed vim:amd64 2:9.1.0016-1\n"
	content += "2024-01-16 11:00:00 status installed bash:amd64 5.2.21-2\n"
	content += "2024-01-10 08:00:00 status installed vim:amd64 2:9.0.0000-1\n" // older vim entry

	result := parseDpkgLogFile(strings.NewReader(content))

	vimDate, ok := result["vim"]
	if !ok {
		t.Fatal("expected vim to be present in result")
	}
	expected := time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)
	if !vimDate.Equal(expected) {
		t.Errorf("expected oldest vim date %v, got %v", expected, vimDate)
	}

	bashDate, ok := result["bash"]
	if !ok {
		t.Fatal("expected bash to be present in result")
	}
	expectedBash := time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC)
	if !bashDate.Equal(expectedBash) {
		t.Errorf("expected bash date %v, got %v", expectedBash, bashDate)
	}
}

func TestParseDpkgLog_IgnoresNonInstallLines(t *testing.T) {
	content := "2024-01-15 10:30:45 status half-configured vim:amd64 2:9.1.0016-1\n"
	content += "2024-01-15 10:30:50 configure vim:amd64 2:9.1.0016-1 <none>\n"

	result := parseDpkgLogFile(strings.NewReader(content))
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d entries", len(result))
	}
}

func TestParseDpkgLog_EmptyInput(t *testing.T) {
	result := parseDpkgLogFile(strings.NewReader(""))
	if len(result) != 0 {
		t.Errorf("expected empty result for empty input, got %d entries", len(result))
	}
}
