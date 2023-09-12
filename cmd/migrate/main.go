// Command migrate applies the store's embedded Postgres migrations without starting
// the rest of the bot. Useful for CI/ops to migrate ahead of a deploy.
package main

import (
	"fmt"
	"os"

	"github.com/jakeale/tonberry/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migrate:", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	return store.RunMigrations(dsn)
}
