package cmd

import (
	"fmt"

	"github.com/rune-context/rune/internal/context"
)

// Context produces minimal context for a request.
func Context(root, query string) error {
	result, err := context.Query(root, query)
	if err != nil {
		return fmt.Errorf("querying context: %w", err)
	}

	fmt.Print(result.Format())
	return nil
}
