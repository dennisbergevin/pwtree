package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	projects      multiFlag
	configFile    string
	onlyChanged   = flag.Bool("only-changed", false, "Show only tests related to changed files")
	lastFailed    = flag.Bool("last-failed", false, "Show only tests that failed last run")
	showSkipped   = flag.Bool("skipped", false, "Show only tests with [skipped] annotation")
	showFixme     = flag.Bool("fixme", false, "Show only tests with [fixme] annotation")
	showFail      = flag.Bool("fail", false, "Show only tests with [fail] annotation")
	titleStyle    = lipgloss.NewStyle().Bold(true)
	jsonDataPath  string
	ciMode        = flag.Bool("ci", false, "Disable colors and emojis for CI environments")
	filterString  string
	helpRequested = flag.Bool("help", false, "Show this help message")
)

func init() {
	flag.Var(&projects, "project", "Project(s) to filter (space-separated or repeatable)")
	flag.StringVar(&filterString, "filter", "", "Comma-separated list of filter terms. Use -prefix for exclusion.")
	flag.StringVar(&configFile, "config", "", "Path to Playwright config file")
	flag.StringVar(&configFile, "c", "", "Shorthand for --config")
	flag.StringVar(&jsonDataPath, "json-data-path", "", "Path to existing JSON file housing output of 'npx playwright test --list --reporter=json'")
	flag.BoolVar(helpRequested, "h", false, "Shorthand for --help")
}

func main() {
	flag.Parse()

	styles, display, emojis := loadStyleConfig()

	if *helpRequested {
		printHelp(emojis.Root)
		os.Exit(0)
	}

	var raw []byte
	var err error
	if jsonDataPath != "" {
		raw, err = os.ReadFile(jsonDataPath)
		if err != nil {
			fmt.Printf("Error reading JSON data from file: %v\n", err)
			os.Exit(1)
		}
	} else {
		raw = runPlaywrightList(projects, *onlyChanged, *lastFailed, configFile)
	}

	var pwData PlaywrightJSON
	if err := json.Unmarshal(raw, &pwData); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	if len(projects) > 0 {
		projectSet := map[string]bool{}
		for _, p := range projects {
			projectSet[p] = true
		}
		pwData.Suites = filterSuitesByProject(pwData.Suites, projectSet)
	}

	if filterString != "" {
		terms := strings.Split(filterString, ";")
		pwData.Suites = filterSuitesByFilter(pwData.Suites, terms)
	}

	if *showFail || *showSkipped || *showFixme {
		pwData.Suites = filterSuitesByAnnotation(pwData.Suites, *showSkipped, *showFixme, *showFail)
	}

	filteredRaw, err := json.Marshal(pwData)
	if err != nil {
		fmt.Printf("Error encoding filtered JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(buildTreeView(filteredRaw, styles, display, emojis))
}

func printHelp(rootEmoji string) {
	displayEmoji := rootEmoji
	if displayEmoji == "" {
		displayEmoji = ""
	}

	const helpText = `Usage:
  pwtree [flags]

Flags:
  --project [project-name]        Project(s) to filter (space-separated or repeatable)
  --filter [filter-string]        Semicolon separated list of filter terms. Use - for exclusion.
  --only-changed                  Show only tests related to changed files
  --last-failed                   Show only tests that failed last run
  --skipped                       Show only tests with [skipped] annotation
  --fixme                         Show only tests with [fixme] annotation
  --fail                          Show only tests with [fail] annotation
  --config, -c [file path]        Path to Playwright config file
  --json-data-path [file path]    Path to existing JSON file housing output of 'npx playwright test --list --reporter=json'
  --ci                            Disable colors and emojis for CI environments
  --help, -h                      Show this help message
`
	fmt.Printf("\n%s\n\n%s", titleStyle.Render(strings.TrimSpace(displayEmoji+" Playwright-tree")), helpText)
}
