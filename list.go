package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

type PlaywrightError struct {
	Message string `json:"message"`
}

type PlaywrightOutput struct {
	Errors []PlaywrightError `json:"errors"`
}

func (m *multiFlag) Set(value string) error {
	parts := strings.Fields(value)
	*m = append(*m, parts...)
	return nil
}

func runPlaywrightList(projects multiFlag, onlyChanged, lastFailed bool, config string) []byte {
	args := []string{"playwright", "test", "--list", "--reporter=json"}

	if config != "" {
		args = append(args, "--config", config)
	}

	for _, p := range projects {
		args = append(args, "--project", p)
	}

	if onlyChanged {
		args = append(args, "--only-changed")
	}
	if lastFailed {
		args = append(args, "--last-failed")
	}

	cmd := exec.Command("npx", args...)
	fmt.Println("Running command:", "npx", strings.Join(args, " "))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	var parsed PlaywrightOutput
	if err := cmd.Run(); err != nil {
		if err := json.Unmarshal(out.Bytes(), &parsed); err == nil {
			if len(parsed.Errors) > 0 {
				fmt.Println(parsed.Errors[0].Message)
				os.Exit(1)
			}
		} else {
			fmt.Println("Error running Playwright:", err)
			os.Exit(1)
		}
	}

	return out.Bytes()
}
