#!/bin/bash
# Export user_reports data to CSV for backup or migration

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DB_PATH="${1:-$SCRIPT_DIR/../data/whatsapp.db}"
OUTPUT="${2:-$SCRIPT_DIR/../data/user_reports_export.csv}"

if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database not found at $DB_PATH"
    exit 1
fi

echo "Exporting from: $DB_PATH"
echo "Output: $OUTPUT"

sqlite3 -header -csv "$DB_PATH" "SELECT user_id, name, streak, activity_count, last_report_date FROM user_reports ORDER BY activity_count DESC;" > "$OUTPUT"

COUNT=$(wc -l < "$OUTPUT")
echo "✅ Exported $((COUNT - 1)) records to $OUTPUT"
