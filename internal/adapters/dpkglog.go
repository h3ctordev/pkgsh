package adapters

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ParseDpkgLog lee las rotaciones de dpkg.log y devuelve un mapa
// nombre → fecha de instalación más antigua encontrada.
func ParseDpkgLog() map[string]time.Time {
	result := make(map[string]time.Time)

	files := []string{
		"/var/log/dpkg.log",
		"/var/log/dpkg.log.1",
	}

	// Find gzip rotations
	gzMatches, _ := filepath.Glob("/var/log/dpkg.log.*.gz")
	files = append(files, gzMatches...)

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		func() {
			defer f.Close()
			var r io.Reader
			if strings.HasSuffix(path, ".gz") {
				gz, err := gzip.NewReader(f)
				if err != nil {
					return
				}
				defer gz.Close()
				r = gz
			} else {
				r = f
			}
			mergeDpkgLogDates(result, r)
		}()
	}

	return result
}

// parseDpkgLogFile parses lines from a dpkg.log reader and returns a map of
// package name → oldest install date. Exported for testing.
func parseDpkgLogFile(r io.Reader) map[string]time.Time {
	result := make(map[string]time.Time)
	mergeDpkgLogDates(result, r)
	return result
}

func mergeDpkgLogDates(result map[string]time.Time, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, " status installed ") {
			continue
		}
		// Format: "2024-01-15 10:30:45 status installed vim:amd64 2:9.1.0016-1"
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		dateStr := fields[0] + " " + fields[1]
		t, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			continue
		}
		pkgName := fields[4]
		if idx := strings.Index(pkgName, ":"); idx != -1 {
			pkgName = pkgName[:idx]
		}
		if existing, ok := result[pkgName]; !ok || t.Before(existing) {
			result[pkgName] = t
		}
	}
}
