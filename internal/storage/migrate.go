package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// RunMigrations applies SQL files in dir (skips 001 — use only on fresh DB via postgres init).
func (ps *PostgresStorage) RunMigrations(dir string) error {
	entries, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(entries)
	for _, path := range entries {
		base := filepath.Base(path)
		if strings.HasPrefix(base, "001_") {
			continue
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", base, err)
		}
		if _, err := ps.db.Exec(string(body)); err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "42P07" {
				continue // relation already exists
			}
			if strings.Contains(err.Error(), "already exists") {
				continue
			}
			return fmt.Errorf("apply %s: %w", base, err)
		}
	}
	return nil
}
