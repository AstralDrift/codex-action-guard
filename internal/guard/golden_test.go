package guard

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var updateGolden = flag.Bool("update", false, "update golden report files")

func TestGoldenReports(t *testing.T) {
	report, err := AuditPath(filepath.Join("..", "..", "fixtures", "secure"), AuditOptions{All: true, ToolVersion: "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	normalizeGoldenReport(&report)

	cases := []struct {
		name string
		path string
		data []byte
	}{
		{name: "markdown", path: filepath.Join("..", "..", "testdata", "golden", "markdown", "secure.md"), data: []byte(RenderMarkdown(report))},
		{name: "json", path: filepath.Join("..", "..", "testdata", "golden", "json", "secure.json"), data: mustGoldenJSON(t, report)},
		{name: "sarif", path: filepath.Join("..", "..", "testdata", "golden", "sarif", "secure.sarif"), data: mustGoldenSARIF(t, report)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if *updateGolden {
				if err := os.MkdirAll(filepath.Dir(tc.path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(tc.path, tc.data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(want, tc.data) {
				t.Fatalf("golden mismatch for %s\nwant %d bytes\ngot  %d bytes\nrun go test ./internal/guard -update to refresh intentionally", tc.path, len(want), len(tc.data))
			}
		})
	}
}

func normalizeGoldenReport(report *Report) {
	report.Root = "<ROOT>"
	report.Metadata.GeneratedAt = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
}

func mustGoldenJSON(t *testing.T, report Report) []byte {
	t.Helper()
	data, err := RenderJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func mustGoldenSARIF(t *testing.T, report Report) []byte {
	t.Helper()
	data, err := RenderSARIF(report)
	if err != nil {
		t.Fatal(err)
	}
	var shape any
	if err := json.Unmarshal(data, &shape); err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}
