package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

type StyleEntry struct {
	Name   string `json:"name"`
	Color  string `json:"color"`
	Bold   bool   `json:"bold"`
	Italic bool   `json:"italic"`
	Faint  bool   `json:"faint"`
}

type EmojiConfig struct {
	Root  *string `json:"root,omitempty"`
	File  *string `json:"file,omitempty"`
	Suite *string `json:"suite,omitempty"`
}

type DisplayEmojis struct {
	Root  string
	File  string
	Suite string
}

type FullConfig struct {
	Styles         []StyleEntry `json:"styles"`
	ShowProjects   *bool        `json:"showProjects,omitempty"`
	ShowTags       *bool        `json:"showTags,omitempty"`
	ShowFileLines  *bool        `json:"showFileLines,omitempty"`
	EmojiOverrides EmojiConfig  `json:"emojis,omitempty"`
}

type DisplayOptions struct {
	ShowProjects  bool
	ShowTags      bool
	ShowFileLines bool
}

func defaultStyles() map[string]lipgloss.Style {
	return map[string]lipgloss.Style{
		"enumerator": lipgloss.NewStyle().Foreground(lipgloss.Color("")).PaddingRight(1),
		"root":       lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"item":       lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"tag":        lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"project":    lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"fileLine":   lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"skipped":    lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"fixme":      lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"fail":       lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"test":       lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"counter":    lipgloss.NewStyle(),
		"file":       lipgloss.NewStyle().Foreground(lipgloss.Color("")),
		"suite":      lipgloss.NewStyle().Foreground(lipgloss.Color("")),
	}
}

func loadStyleConfig() (map[string]lipgloss.Style, DisplayOptions, DisplayEmojis) {
	paths := []string{
		"./.pwtree.json",
		filepath.Join(os.Getenv("HOME"), ".config", "pwtree", "config.json"),
	}

	var configData []byte
	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			configData = data
			break
		}
	}

	defaultDisplay := DisplayOptions{
		ShowProjects:  true,
		ShowTags:      true,
		ShowFileLines: true,
	}
	defaultEmojis := DisplayEmojis{
		Root:  "",
		File:  "",
		Suite: "",
	}

	if *ciMode {
		// Return empty styles and emojis in CI mode
		return map[string]lipgloss.Style{}, defaultDisplay, DisplayEmojis{}
	}

	if len(configData) == 0 {
		return defaultStyles(), defaultDisplay, defaultEmojis
	}

	var cfg FullConfig
	if err := json.Unmarshal(configData, &cfg); err != nil {
		fmt.Printf("Error parsing config: %v\n", err)
		return defaultStyles(), defaultDisplay, defaultEmojis
	}

	styles := map[string]lipgloss.Style{}
	for _, entry := range cfg.Styles {
		style := lipgloss.NewStyle()
		if entry.Color != "" {
			style = style.Foreground(lipgloss.Color(entry.Color))
		}
		if entry.Bold {
			style = style.Bold(true)
		}
		if entry.Italic {
			style = style.Italic(true)
		}
		if entry.Faint {
			style = style.Faint(true)
		}
		styles[entry.Name] = style
	}

	emojis := DisplayEmojis{
		Root:  defaultEmojis.Root,
		File:  defaultEmojis.File,
		Suite: defaultEmojis.Suite,
	}

	if cfg.EmojiOverrides.Root != nil {
		emojis.Root = *cfg.EmojiOverrides.Root
	}
	if cfg.EmojiOverrides.File != nil {
		emojis.File = *cfg.EmojiOverrides.File
	}
	if cfg.EmojiOverrides.Suite != nil {
		emojis.Suite = *cfg.EmojiOverrides.Suite
	}

	return styles, DisplayOptions{
		ShowProjects:  cfg.ShowProjects == nil || *cfg.ShowProjects,
		ShowTags:      cfg.ShowTags == nil || *cfg.ShowTags,
		ShowFileLines: cfg.ShowFileLines == nil || *cfg.ShowFileLines,
	}, emojis
}
