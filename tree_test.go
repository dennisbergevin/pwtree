package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestLoadPlaywrightJSON(t *testing.T) {
	jsonData := []byte(`{
		"suites": [{
			"title": "sanity/agentView.spec.ts",
			"file": "sanity/agentView.spec.ts",
			"suites": [{
				"title": "agent tests",
				"file": "sanity/agentView.spec.ts",
				"line": 3,
				"specs": [{
					"title": "get started link",
					"file": "sanity/agentView.spec.ts",
					"line": 4,
					"tags": ["smoke"],
					"tests": [{
						"projectName": "chromium",
						"annotations": [{"type": "fail"}],
						"status": "skipped"
					}]
				}]
			}]
		}]
	}`)

	data, err := loadPlaywrightJSON(jsonData)
	if err != nil {
		t.Fatalf("Unexpected error loading JSON: %v", err)
	}
	if len(data.Suites) != 1 {
		t.Fatalf("Expected 1 top-level suite, got %d", len(data.Suites))
	}
	if data.Suites[0].Title != "sanity/agentView.spec.ts" {
		t.Errorf("Unexpected top-level suite title: %s", data.Suites[0].Title)
	}
}

func TestSuiteLine(t *testing.T) {
	s1 := Suite{Line: 10}
	if suiteLine(s1) != 10 {
		t.Errorf("Expected line 10, got %d", suiteLine(s1))
	}

	s2 := Suite{
		Specs: []Spec{{Line: 20}},
	}
	if suiteLine(s2) != 20 {
		t.Errorf("Expected line 20 from spec, got %d", suiteLine(s2))
	}

	s3 := Suite{
		Suites: []Suite{
			{Specs: []Spec{{Line: 30}}},
		},
	}
	if suiteLine(s3) != 30 {
		t.Errorf("Expected line 30 from nested suite, got %d", suiteLine(s3))
	}

	s4 := Suite{}
	if suiteLine(s4) != 0 {
		t.Errorf("Expected line 0, got %d", suiteLine(s4))
	}
}

func TestBuildTreeView_BasicRender(t *testing.T) {
	jsonData := []byte(`{
		"suites": [{
			"title": "sanity/agentView.spec.ts",
			"file": "sanity/agentView.spec.ts",
			"suites": [{
				"title": "agent tests",
				"file": "sanity/agentView.spec.ts",
				"line": 3,
				"specs": [{
					"title": "get started link",
					"file": "sanity/agentView.spec.ts",
					"line": 4,
					"tags": ["smoke"],
					"tests": [{
						"projectName": "chromium",
						"annotations": [],
						"status": "skipped"
					}]
				}]
			}]
		}]
	}`)

	styles := map[string]lipgloss.Style{
		"enumerator": {},
		"root":       {},
		"tag":        {},
		"project":    {},
		"fileLine":   {},
		"skipped":    {},
		"fixme":      {},
		"fail":       {},
		"test":       {},
		"counter":    {},
		"file":       {},
		"suite":      {},
	}

	display := DisplayOptions{
		ShowTags:      true,
		ShowProjects:  true,
		ShowFileLines: true,
	}
	emojis := DisplayEmojis{
		Root:  "🌳",
		File:  "📄",
		Suite: "📁",
	}

	output := buildTreeView(jsonData, styles, display, emojis)

	if !strings.Contains(output, "get started link") {
		t.Errorf("Expected test title 'get started link'\nGot:\n%s", output)
	}
	if !strings.Contains(output, "chromium") {
		t.Errorf("Expected output to include project name 'chromium'\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Total: 1 test") {
		t.Errorf("Expected total test count to be correct\nGot:\n%s", output)
	}
}

func TestBuildTreeView_DeeplyNestedSuites(t *testing.T) {
	jsonData := []byte(`{
		"suites": [{
			"title": "Root",
			"file": "deep.go",
			"suites": [{
				"title": "Level 1",
				"suites": [{
					"title": "Level 2",
					"suites": [{
						"title": "Level 3",
						"specs": [{
							"title": "Deep Test",
							"file": "deep.go",
							"line": 99,
							"tests": [{
								"projectName": "chromium",
								"annotations": []
							}]
						}]
					}]
				}]
			}]
		}]
	}`)

	styles := map[string]lipgloss.Style{
		"enumerator": lipgloss.NewStyle(),
		"root":       lipgloss.NewStyle(),
		"tag":        lipgloss.NewStyle(),
		"project":    lipgloss.NewStyle(),
		"fileLine":   lipgloss.NewStyle(),
		"skipped":    lipgloss.NewStyle(),
		"fixme":      lipgloss.NewStyle(),
		"fail":       lipgloss.NewStyle(),
		"test":       lipgloss.NewStyle(),
		"counter":    lipgloss.NewStyle(),
		"file":       lipgloss.NewStyle(),
		"suite":      lipgloss.NewStyle(),
	}

	display := DisplayOptions{
		ShowTags:      true,
		ShowProjects:  true,
		ShowFileLines: true,
	}
	emojis := DisplayEmojis{
		Root:  "🌳",
		File:  "📄",
		Suite: "📁",
	}

	output := buildTreeView(jsonData, styles, display, emojis)

	if !strings.Contains(output, "Level 3") {
		t.Errorf("Expected deeply nested suite 'Level 3' to be included\nOutput:\n%s", output)
	}
	if !strings.Contains(output, "Deep Test") {
		t.Errorf("Expected test 'Deep Test' to be rendered\nOutput:\n%s", output)
	}
	if !strings.Contains(output, "chromium") {
		t.Errorf("Expected project name 'chromium' to appear\nOutput:\n%s", output)
	}
}

func TestBuildTreeView_WithAnnotations(t *testing.T) {
	jsonData := []byte(`{
		"suites": [{
			"title": "Suite",
			"file": "annot.go",
			"specs": [{
				"title": "Skipped Test",
				"file": "annot.go",
				"line": 10,
				"tests": [{
					"projectName": "firefox",
					"annotations": [{"type": "skip"}]
				}]
			}, {
				"title": "Fixme Test",
				"file": "annot.go",
				"line": 11,
				"tests": [{
					"projectName": "firefox",
					"annotations": [{"type": "fixme"}]
				}]
			}, {
				"title": "Failing Test",
				"file": "annot.go",
				"line": 12,
				"tests": [{
					"projectName": "firefox",
					"annotations": [{"type": "fail"}]
				}]
			}]
		}]
	}`)

	styles := map[string]lipgloss.Style{
		"enumerator": lipgloss.NewStyle(),
		"root":       lipgloss.NewStyle(),
		"tag":        lipgloss.NewStyle(),
		"project":    lipgloss.NewStyle(),
		"fileLine":   lipgloss.NewStyle(),
		"skipped":    lipgloss.NewStyle(),
		"fixme":      lipgloss.NewStyle(),
		"fail":       lipgloss.NewStyle(),
		"test":       lipgloss.NewStyle(),
		"counter":    lipgloss.NewStyle(),
		"file":       lipgloss.NewStyle(),
		"suite":      lipgloss.NewStyle(),
	}

	display := DisplayOptions{
		ShowTags:      true,
		ShowProjects:  true,
		ShowFileLines: true,
	}
	emojis := DisplayEmojis{
		Root:  "🌳",
		File:  "📄",
		Suite: "📁",
	}

	output := buildTreeView(jsonData, styles, display, emojis)

	if !strings.Contains(output, "Skipped Test") || !strings.Contains(output, "[skipped]") {
		t.Errorf("Expected 'Skipped Test' with '[skipped]' annotation\nOutput:\n%s", output)
	}
	if !strings.Contains(output, "Fixme Test") || !strings.Contains(output, "[fixme]") {
		t.Errorf("Expected 'Fixme Test' with '[fixme]' annotation\nOutput:\n%s", output)
	}
	if !strings.Contains(output, "Failing Test") || !strings.Contains(output, "[fail]") {
		t.Errorf("Expected 'Failing Test' with '[fail]' annotation\nOutput:\n%s", output)
	}
}
