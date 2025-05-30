package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/tree"
)

type PlaywrightJSON struct {
	Suites []Suite `json:"suites"`
}

type Annotation struct {
	Type string `json:"type"`
}

type TestInstance struct {
	ProjectName string       `json:"projectName"`
	Annotations []Annotation `json:"annotations"`
}

type Spec struct {
	Title string         `json:"title"`
	Tags  []string       `json:"tags"`
	Tests []TestInstance `json:"tests"`
	File  string         `json:"file"`
	Line  int            `json:"line"`
}

type Suite struct {
	Title  string  `json:"title"`
	File   string  `json:"file"`
	Line   int     `json:"line"`
	Suites []Suite `json:"suites"`
	Specs  []Spec  `json:"specs"`
}

func loadPlaywrightJSON(raw []byte) (PlaywrightJSON, error) {
	var pwData PlaywrightJSON
	err := json.Unmarshal(raw, &pwData)
	return pwData, err
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func suiteLine(s Suite) int {
	if s.Line != 0 {
		return s.Line
	}
	if len(s.Specs) > 0 {
		return s.Specs[0].Line
	}
	for _, child := range s.Suites {
		if line := suiteLine(child); line != 0 {
			return line
		}
	}
	return 0
}

func buildTreeView(jsonData []byte, styles map[string]lipgloss.Style, display DisplayOptions, emojis DisplayEmojis) string {
	var pwData PlaywrightJSON
	if err := json.Unmarshal(jsonData, &pwData); err != nil {
		return fmt.Sprintf("Error parsing JSON: %v", err)
	}

	// Styles
	enumeratorStyle := styles["enumerator"]
	rootStyle := styles["root"]
	tagStyle := styles["tag"]
	projectStyle := styles["project"]
	fileLineStyle := styles["fileLine"]
	skippedStyle := styles["skipped"]
	fixmeStyle := styles["fixme"]
	failStyle := styles["fail"]
	testStyle := styles["test"]
	counterStyle := styles["counter"]
	fileNodeStyle := styles["file"]
	suiteNodeStyle := styles["suite"]

	title := strings.TrimSpace(emojis.Root + " Playwright-tree")
	root := tree.Root(title).
		Enumerator(tree.RoundedEnumerator).
		EnumeratorStyle(enumeratorStyle).
		RootStyle(rootStyle)

	seenTests := map[string]bool{}
	totalTests := 0
	totalFiles := 0

	var processSuite func(s Suite, parent *tree.Tree, parentFile string) (*tree.Tree, bool)

	processSuite = func(suite Suite, parent *tree.Tree, parentFile string) (*tree.Tree, bool) {
		currentFile := suite.File
		if currentFile == "" {
			currentFile = parentFile
		}

		var suiteNode *tree.Tree = parent
		if suite.Title != "" && suite.Title != suite.File {
			line := suiteLine(suite)
			fileLineStr := ""
			if display.ShowFileLines {
				fileLineStr = fileLineStyle.Render(fmt.Sprintf("(%s:%d)", currentFile, line))
			}
			label := strings.TrimSpace(fmt.Sprintf("%s %s %s", emojis.Suite, suite.Title, fileLineStr))
			suiteLabel := suiteNodeStyle.Render(label)
			suiteNode = tree.Root(suiteLabel)
		}

		type aggSpec struct {
			Title    string
			File     string
			Line     int
			Tags     map[string]bool
			Projects map[string]bool
			Skipped  bool
			Fixme    bool
			Fail     bool
		}
		aggSpecs := map[string]*aggSpec{}

		for _, spec := range suite.Specs {
			key := fmt.Sprintf("%s:%d:%s", spec.File, spec.Line, spec.Title)
			as, exists := aggSpecs[key]
			if !exists {
				as = &aggSpec{
					Title:    spec.Title,
					File:     spec.File,
					Line:     spec.Line,
					Tags:     map[string]bool{},
					Projects: map[string]bool{},
				}
				aggSpecs[key] = as
			}
			for _, tag := range spec.Tags {
				as.Tags[tag] = true
			}
			for _, test := range spec.Tests {
				as.Projects[test.ProjectName] = true
				for _, ann := range test.Annotations {
					switch ann.Type {
					case "skip":
						as.Skipped = true
					case "fixme":
						as.Fixme = true
					case "fail":
						as.Fail = true
					}
				}
			}
		}

		var hasVisibleSpecs bool

		for _, as := range aggSpecs {
			showAny := *showSkipped || *showFixme || *showFail
			matchesAnnotation := (!showAny) ||
				(*showSkipped && as.Skipped) ||
				(*showFixme && as.Fixme) ||
				(*showFail && as.Fail)

			if !matchesAnnotation {
				continue
			}

			key := fmt.Sprintf("%s:%d:%s", as.File, as.Line, as.Title)
			if seenTests[key] {
				continue
			}
			seenTests[key] = true
			hasVisibleSpecs = true

			projectCount := len(as.Projects)
			totalTests += projectCount

			var tags []string
			for t := range as.Tags {
				tags = append(tags, t)
			}
			sort.Strings(tags)
			tagStr := ""
			if display.ShowTags && len(tags) > 0 {
				tagStr = tagStyle.Render(" [" + strings.Join(tags, ", ") + "]")
			}

			var projects []string
			for p := range as.Projects {
				projects = append(projects, p)
			}
			sort.Strings(projects)
			projectStr := ""
			if display.ShowProjects && len(projects) > 0 {
				projectStr = projectStyle.Render(" (" + strings.Join(projects, ", ") + ")")
			}

			titleLabel := as.Title
			if as.Skipped {
				titleLabel += " [skipped]"
			}
			if as.Fixme {
				titleLabel += " [fixme]"
			}
			if as.Fail {
				titleLabel += " [fail]"
			}

			var title string
			switch {
			case as.Skipped:
				title = skippedStyle.Render(titleLabel)
			case as.Fixme:
				title = fixmeStyle.Render(titleLabel)
			case as.Fail:
				title = failStyle.Render(titleLabel)
			default:
				title = testStyle.Render(titleLabel)
			}

			fileLineStr := ""
			if display.ShowFileLines {
				fileLineStr = fileLineStyle.Render(fmt.Sprintf("(%s:%d)", as.File, as.Line))
			}
			specLabel := fmt.Sprintf("%s%s%s %s", title, projectStr, tagStr, fileLineStr)
			specNode := tree.Root(specLabel)
			suiteNode.Child(specNode)
		}

		var hasVisibleChildren bool
		for _, child := range suite.Suites {
			if childNode, ok := processSuite(child, suiteNode, currentFile); ok {
				suiteNode.Child(childNode)
				hasVisibleChildren = true
			}
		}

		if hasVisibleSpecs || hasVisibleChildren {
			if suiteNode != parent {
				return suiteNode, true
			}
			return parent, true
		}
		return nil, false
	}

	for _, topSuite := range pwData.Suites {
		currentFile := topSuite.File
		if currentFile == "" {
			continue
		}
		label := strings.TrimSpace(emojis.File + " " + currentFile)
		fileNode := tree.Root(fileNodeStyle.Render(label))
		if node, ok := processSuite(topSuite, fileNode, currentFile); ok {
			root.Child(node)
			totalFiles++
		}
	}

	counter := fmt.Sprintf("Total: %d test%s in %d file%s",
		totalTests, pluralize(totalTests), totalFiles, pluralize(totalFiles))

	return "\n" + root.String() + "\n\n" + counterStyle.Render(counter) + "\n"
}
