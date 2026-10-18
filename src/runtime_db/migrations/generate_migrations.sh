#!/usr/bin/env bash
set -euo pipefail

if [ $# -ne 1 ]; then
    echo "Usage: $0 <migration-name>" >&2
    exit 1
fi

MIGRATION_NAME="$1"
MIGRATIONS_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if ! command -v atlas >/dev/null 2>&1; then
    echo "Error: atlas CLI not found in PATH" >&2
    exit 1
fi

TIMESTAMP=$(date +"%Y%m%d%H%M%S")

# Only files atlas creates in this run are renamed: renaming an existing migration would replay it as a new version.
for DB in mariadb mysql postgres sqlite; do
    DIR="$MIGRATIONS_ROOT/$DB"
    BEFORE=$(ls -1 "$DIR"/*.sql 2>/dev/null || true)

    echo "Generating migration for $DB..."
    atlas migrate diff "$MIGRATION_NAME" --env gorm-"$DB"

    NEW_FILE=$(comm -13 <(echo "$BEFORE") <(ls -1 "$DIR"/*.sql 2>/dev/null || true))
    if [ -z "$NEW_FILE" ]; then
        echo "No schema change for $DB, skipping."
        echo
        continue
    fi

    NAME_PART=$(basename "$NEW_FILE" | cut -d'_' -f2-)
    NEW_PATH="$DIR/${TIMESTAMP}_${NAME_PART}"
    mv "$NEW_FILE" "$NEW_PATH"
    echo "Generated $NEW_PATH"

    atlas migrate hash --dir file://"$DB"
    echo
done