package utils

import (
	"net/url"
	"strings"
)

var sqliteDefaultPragmas = []string{
	"foreign_keys(1)",
	"busy_timeout(10000)",
	"journal_mode(WAL)",
}

// ApplySqliteDsnDefaults appends FK enforcement, busy timeout and WAL pragmas;
// a pragma the DSN already sets (any value) is left to the user.
func ApplySqliteDsnDefaults(dsn string) string {
	base, rawQuery, _ := strings.Cut(dsn, "?")

	query, err := url.ParseQuery(rawQuery)
	if err != nil {
		return dsn
	}

	existing := map[string]bool{}
	for _, pragma := range query["_pragma"] {
		existing[sqlitePragmaName(pragma)] = true
	}

	missing := []string{}
	for _, pragma := range sqliteDefaultPragmas {
		if !existing[sqlitePragmaName(pragma)] {
			missing = append(missing, "_pragma="+pragma)
		}
	}
	if len(missing) == 0 {
		return dsn
	}

	parts := []string{}
	if rawQuery != "" {
		parts = append(parts, rawQuery)
	}
	parts = append(parts, missing...)
	return base + "?" + strings.Join(parts, "&")
}

// sqlitePragmaName extracts the pragma name from "name(value)" or "name=value" forms.
func sqlitePragmaName(pragma string) string {
	name := pragma
	if i := strings.IndexAny(name, "(="); i >= 0 {
		name = name[:i]
	}
	return strings.ToLower(strings.TrimSpace(name))
}
