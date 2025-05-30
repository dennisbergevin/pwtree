package main

import (
	"testing"
)

func TestFilterSuitesByAnnotation(t *testing.T) {
	suites := []Suite{
		{
			Title: "sanity/agentView.spec.ts",
			File:  "sanity/agentView.spec.ts",
			Suites: []Suite{
				{
					Title: "agent tests",
					File:  "sanity/agentView.spec.ts",
					Specs: []Spec{
						{
							Title: "get started link",
							Tags:  []string{"smoke", "sanity"},
							File:  "sanity/agentView.spec.ts",
							Line:  4,
							Tests: []TestInstance{
								{
									ProjectName: "chromium",
									Annotations: []Annotation{{Type: "fail"}},
								},
								{
									ProjectName: "webkit",
									Annotations: []Annotation{{Type: "fail"}},
								},
							},
						},
						{
							Title: "not annotated",
							File:  "sanity/agentView.spec.ts",
							Line:  10,
							Tests: []TestInstance{
								{
									ProjectName: "firefox",
									Annotations: nil,
								},
							},
						},
					},
				},
			},
		},
	}

	t.Run("OnlyFail", func(t *testing.T) {
		result := filterSuitesByAnnotation(suites, false, false, true)
		if len(result) != 1 || len(result[0].Suites) != 1 {
			t.Fatalf("Expected root suite and one child suite, got %+v", result)
		}
		specs := result[0].Suites[0].Specs
		if len(specs) != 1 || specs[0].Title != "get started link" {
			t.Errorf("Expected only 'get started link' spec, got %+v", specs)
		}
		if len(specs[0].Tests) != 2 {
			t.Errorf("Expected 2 test instances, got %d", len(specs[0].Tests))
		}
	})

	t.Run("SkipFixmeFalse_ShouldIncludeAll", func(t *testing.T) {
		result := filterSuitesByAnnotation(suites, false, false, false)
		specs := result[0].Suites[0].Specs
		if len(specs) != 2 {
			t.Errorf("Expected 2 specs (all), got %+v", specs)
		}
	})
}

func TestFilterSuitesByProject(t *testing.T) {
	suites := []Suite{
		{
			Title: "sanity/agentView.spec.ts",
			Suites: []Suite{
				{
					Title: "agent tests",
					Specs: []Spec{
						{
							Title: "get started link",
							File:  "sanity/agentView.spec.ts",
							Line:  4,
							Tests: []TestInstance{
								{ProjectName: "chromium"},
								{ProjectName: "webkit"},
							},
						},
						{
							Title: "firefox-only test",
							Tests: []TestInstance{
								{ProjectName: "firefox"},
							},
						},
					},
				},
			},
		},
	}

	t.Run("FilterToChromiumOnly", func(t *testing.T) {
		filtered := filterSuitesByProject(suites, map[string]bool{"chromium": true})
		specs := filtered[0].Suites[0].Specs
		if len(specs) != 1 || specs[0].Title != "get started link" {
			t.Errorf("Expected only 'get started link' spec, got %+v", specs)
		}
		if len(specs[0].Tests) != 1 || specs[0].Tests[0].ProjectName != "chromium" {
			t.Errorf("Expected only chromium test, got %+v", specs[0].Tests)
		}
	})
}

func TestFilterSuites(t *testing.T) {
	suites := []Suite{
		{
			Title: "sanity/agentView.spec.ts",
			Suites: []Suite{
				{
					Title: "agent tests",
					Specs: []Spec{
						{
							Title: "get started link",
							Tests: []TestInstance{
								{ProjectName: "webkit"},
								{ProjectName: "chromium"},
							},
						},
						{
							Title: "firefox-only test",
							Tests: []TestInstance{
								{ProjectName: "firefox"},
							},
						},
					},
				},
			},
		},
	}

	t.Run("ProjectWebkit", func(t *testing.T) {
		projectSet := map[string]bool{"webkit": true}
		result := filterSuites(suites, projectSet)
		specs := result[0].Suites[0].Specs

		if len(specs) != 1 || specs[0].Title != "get started link" {
			t.Errorf("Expected one spec for webkit, got %+v", specs)
		}
		if len(specs[0].Tests) != 1 || specs[0].Tests[0].ProjectName != "webkit" {
			t.Errorf("Expected webkit test only, got %+v", specs[0].Tests)
		}
	})

	t.Run("ProjectFirefox", func(t *testing.T) {
		projectSet := map[string]bool{"firefox": true}
		result := filterSuites(suites, projectSet)

		specs := result[0].Suites[0].Specs
		if len(specs) != 1 || specs[0].Title != "firefox-only test" {
			t.Errorf("Expected only firefox test spec, got %+v", specs)
		}
	})
}

func TestFilterSuitesByFilter_PositiveMatch(t *testing.T) {
	suites := []Suite{
		{
			Title: "root suite",
			File:  "root.spec.ts",
			Suites: []Suite{
				{
					Title: "child suite",
					File:  "child.spec.ts",
					Specs: []Spec{
						{
							Title: "should login successfully",
							File:  "child.spec.ts",
							Tags:  []string{"auth", "smoke"},
							Tests: []TestInstance{{ProjectName: "chromium"}},
						},
					},
				},
			},
		},
	}

	filtered := filterSuitesByFilter(suites, []string{"login"})
	if len(filtered) != 1 || len(filtered[0].Suites) != 1 {
		t.Fatalf("Expected 1 root suite and 1 matching child suite")
	}

	specs := filtered[0].Suites[0].Specs
	if len(specs) != 1 || specs[0].Title != "should login successfully" {
		t.Errorf("Expected 'should login successfully', got %+v", specs)
	}
}

func TestFilterSuitesByFilter_NegativeMatch(t *testing.T) {
	suites := []Suite{
		{
			Title: "root suite",
			File:  "root.spec.ts",
			Suites: []Suite{
				{
					Title: "auth suite",
					File:  "auth.spec.ts",
					Specs: []Spec{
						{
							Title: "should login successfully",
							File:  "auth.spec.ts",
							Tags:  []string{"auth", "smoke"},
							Tests: []TestInstance{{ProjectName: "chromium"}},
						},
						{
							Title: "should skip on timeout",
							File:  "auth.spec.ts",
							Tags:  []string{"timeout", "skip"},
							Tests: []TestInstance{{ProjectName: "firefox"}},
						},
					},
				},
			},
		},
	}

	filtered := filterSuitesByFilter(suites, []string{"auth", "-skip"})
	if len(filtered) != 1 || len(filtered[0].Suites) != 1 {
		t.Fatalf("Expected 1 matching root/child suite")
	}

	specs := filtered[0].Suites[0].Specs
	if len(specs) != 1 || specs[0].Title != "should login successfully" {
		t.Errorf("Expected 'should login successfully' to remain, got %+v", specs)
	}
}

func TestMatchesFilterAndNegativeFilter(t *testing.T) {
	t.Run("matchesFilter positive", func(t *testing.T) {
		match := matchesFilter("get started link", "tests/example.spec.ts", []string{"smoke"}, []string{"started"})
		if !match {
			t.Error("Expected positive match for 'started'")
		}
	})

	t.Run("matchesFilter negative", func(t *testing.T) {
		match := matchesFilter("get started link", "tests/example.spec.ts", []string{"smoke"}, []string{"-smoke"})
		if match {
			t.Error("Expected match to be false due to negative tag filter")
		}
	})

	t.Run("matchesNegativeFilter matches tag", func(t *testing.T) {
		if !matchesNegativeFilter("test", "file.ts", []string{"flaky"}, []string{"-flaky"}) {
			t.Error("Expected tag to match negative filter")
		}
	})
}

func TestApplyNegativeFilterToSuites(t *testing.T) {
	suites := []Suite{
		{
			Title: "top suite",
			Suites: []Suite{
				{
					Title: "child suite",
					Specs: []Spec{
						{
							Title: "should run",
							Tags:  []string{"run"},
						},
						{
							Title: "should be flaky",
							Tags:  []string{"flaky"},
						},
					},
				},
			},
		},
	}

	filtered := applyNegativeFilterToSuites(suites, []string{"-flaky"})
	if len(filtered) != 1 || len(filtered[0].Suites) != 1 {
		t.Fatalf("Expected filtered suite structure to be intact")
	}

	specs := filtered[0].Suites[0].Specs
	if len(specs) != 1 || specs[0].Title != "should run" {
		t.Errorf("Expected only 'should run' spec to remain, got %+v", specs)
	}
}
