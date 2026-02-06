#!/bin/bash
# Import user_reports data from CSV

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DB_PATH="${1:-$SCRIPT_DIR/../data/whatsapp.db}"
INPUT="${2:-$SCRIPT_DIR/../data/user_reports_export.csv}"

if [ ! -f "$INPUT" ]; then
    echo "Error: CSV file not found at $INPUT"
    exit 1
fi

echo "Importing to: $DB_PATH"
echo "From: $INPUT"

# Create table if not exists
sqlite3 "$DB_PATH" "CREATE TABLE IF NOT EXISTS user_reports (
    user_id TEXT PRIMARY KEY,
    name TEXT,
    streak INTEGER,
    activity_count INTEGER DEFAULT 0,
    last_report_date TEXT
);"

# Clear existing data (optional - comment out to merge instead)
read -p "Clear existing data before import? (y/N): " confirm
if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
    sqlite3 "$DB_PATH" "DELETE FROM user_reports;"
    echo "Cleared existing data"
fi

# Import CSV
sqlite3 "$DB_PATH" <<EOF
.mode csv
.import --skip 1 "$INPUT" user_reports
EOF

COUNT=$(sqlite3 "$DB_PATH" "SELECT COUNT(*) FROM user_reports;")
echo "✅ Imported successfully. Total records: $COUNT"
