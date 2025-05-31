package main

import (
	"strings"
)

func filterSuitesByFilter(suites []Suite, filterTerms []string) []Suite {
	var filtered []Suite

	for _, suite := range suites {
		// If suite matches any negative filter, skip it entirely.
		if matchesNegativeFilter(suite.Title, suite.File, nil, filterTerms) {
			continue
		}

		// Now check if suite matches positive filter or if any child suites/specs do.
		suiteMatches := matchesFilter(suite.Title, suite.File, nil, filterTerms)

		if suiteMatches {
			// Recursively apply negative filter to child suites/specs
			suite.Suites = applyNegativeFilterToSuites(suite.Suites, filterTerms)

			var newSpecs []Spec
			for _, spec := range suite.Specs {
				if !matchesNegativeFilter(spec.Title, spec.File, spec.Tags, filterTerms) {
					newSpecs = append(newSpecs, spec)
				}
			}
			suite.Specs = newSpecs

			filtered = append(filtered, suite)
			continue
		}

		// If suite itself does not match positive filter,
		// recurse to filter child suites/specs for positive matches.
		suite.Suites = filterSuitesByFilter(suite.Suites, filterTerms)

		var newSpecs []Spec
		for _, spec := range suite.Specs {
			if matchesFilter(spec.Title, spec.File, spec.Tags, filterTerms) {
				newSpecs = append(newSpecs, spec)
			}
		}
		suite.Specs = newSpecs

		if len(suite.Suites) > 0 || len(suite.Specs) > 0 {
			filtered = append(filtered, suite)
		}
	}

	return filtered
}

func matchesNegativeFilter(title, file string, tags []string, filters []string) bool {
	for _, f := range filters {
		if strings.HasPrefix(f, "-") {
			term := strings.TrimPrefix(f, "-")
			if strings.Contains(title, term) || strings.Contains(file, term) {
				return true
			}
			for _, tag := range tags {
				if strings.Contains(tag, term) || strings.Contains("@"+tag, term) {
					return true
				}
			}
		}
	}
	return false
}

func applyNegativeFilterToSuites(suites []Suite, filterTerms []string) []Suite {
	var filtered []Suite
	for _, suite := range suites {
		if matchesNegativeFilter(suite.Title, suite.File, nil, filterTerms) {
			continue
		}

		suite.Suites = applyNegativeFilterToSuites(suite.Suites, filterTerms)

		var newSpecs []Spec
		for _, spec := range suite.Specs {
			if !matchesNegativeFilter(spec.Title, spec.File, spec.Tags, filterTerms) {
				newSpecs = append(newSpecs, spec)
			}
		}
		suite.Specs = newSpecs

		if len(suite.Suites) > 0 || len(suite.Specs) > 0 {
			filtered = append(filtered, suite)
		}
	}
	return filtered
}

func matchesFilter(title, file string, tags []string, filters []string) bool {
	var positive []string
	var negative []string

	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if strings.HasPrefix(f, "-") {
			negative = append(negative, strings.TrimPrefix(f, "-"))
		} else {
			positive = append(positive, f)
		}
	}

	for _, neg := range negative {
		if strings.Contains(title, neg) || strings.Contains(file, neg) {
			return false
		}
		for _, tag := range tags {
			if strings.Contains(tag, neg) || strings.Contains("@"+tag, neg) {
				return false
			}
		}
	}

	if len(positive) == 0 {
		return true
	}

	for _, pos := range positive {
		if strings.Contains(title, pos) || strings.Contains(file, pos) {
			return true
		}
		for _, tag := range tags {
			if strings.Contains(tag, pos) || strings.Contains("@"+tag, pos) {
				return true
			}
		}
	}

	return false
}

func filterSuitesByAnnotation(suites []Suite, showSkipped, showFixme, showFail bool) []Suite {
	var filtered []Suite

	for _, suite := range suites {
		suite.Suites = filterSuitesByAnnotation(suite.Suites, showSkipped, showFixme, showFail)

		var newSpecs []Spec
		for _, spec := range suite.Specs {
			var filteredTests []TestInstance
			for _, test := range spec.Tests {
				include := false
				for _, ann := range test.Annotations {
					if (showSkipped && ann.Type == "skip") ||
						(showFixme && ann.Type == "fixme") ||
						(showFail && ann.Type == "fail") {
						include = true
						break
					}
				}
				if include || (!showSkipped && !showFixme && !showFail) {
					filteredTests = append(filteredTests, test)
				}
			}
			if len(filteredTests) > 0 {
				spec.Tests = filteredTests
				newSpecs = append(newSpecs, spec)
			}
		}
		suite.Specs = newSpecs

		if len(suite.Suites) > 0 || len(suite.Specs) > 0 {
			filtered = append(filtered, suite)
		}
	}

	return filtered
}

func filterSuitesByProject(suites []Suite, projectSet map[string]bool) []Suite {
	var filtered []Suite

	for _, suite := range suites {
		suite.Suites = filterSuitesByProject(suite.Suites, projectSet)

		var newSpecs []Spec
		for _, spec := range suite.Specs {
			var filteredTests []TestInstance
			for _, test := range spec.Tests {
				if projectSet[test.ProjectName] {
					filteredTests = append(filteredTests, test)
				}
			}
			if len(filteredTests) > 0 {
				spec.Tests = filteredTests
				newSpecs = append(newSpecs, spec)
			}
		}

		suite.Specs = newSpecs

		if len(suite.Suites) > 0 || len(suite.Specs) > 0 {
			filtered = append(filtered, suite)
		}
	}

	return filtered
}

func filterSuites(suites []Suite, projectSet map[string]bool) []Suite {
	var filtered []Suite

	for _, suite := range suites {
		suite.Suites = filterSuites(suite.Suites, projectSet)

		var newSpecs []Spec
		for _, spec := range suite.Specs {
			var filteredTests []TestInstance
			for _, test := range spec.Tests {
				if projectSet[test.ProjectName] {
					filteredTests = append(filteredTests, test)
				}
			}
			if len(filteredTests) > 0 {
				spec.Tests = filteredTests
				newSpecs = append(newSpecs, spec)
			}
		}
		suite.Specs = newSpecs

		if len(suite.Suites) > 0 || len(suite.Specs) > 0 {
			filtered = append(filtered, suite)
		}
	}
	return filtered
}
