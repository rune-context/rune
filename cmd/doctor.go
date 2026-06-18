package cmd

import (
	"fmt"

	"github.com/rune-context/rune/internal/doctor"
)

// Doctor validates repository health.
func Doctor(root string) error {
	issues := doctor.Check(root)

	if len(issues) == 0 {
		fmt.Println("✓ Repository is healthy.")
		return nil
	}

	icons := map[string]string{
		"error":   "✗",
		"warning": "⚠",
		"info":    "ℹ",
	}

	hasErrors := false
	for _, issue := range issues {
		icon := icons[issue.Level]
		if icon == "" {
			icon = "·"
		}
		fmt.Printf("  %s %s\n", icon, issue.Message)
		if issue.Level == "error" {
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("health check found errors")
	}
	return nil
}
