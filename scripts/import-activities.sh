#!/bin/sh
# Downloads activities for the current (and optionally previous) month
# and extracts them into /watched for Dawarich to pick up.
set -e

YEAR=$(date +%Y)
MONTH=$(date +%-m)
FORMAT="${FLEXCLI_FORMAT:-gpx}"
WATCHED="${FLEXCLI_WATCHED_DIR:-/watched}"

download_and_extract() {
  local year=$1
  local month=$2
  local tmpzip
  tmpzip=$(mktemp /tmp/activities-XXXXXX.zip)

  echo "[$(date)] Downloading $year/$month (format: $FORMAT)..."
  if flexcli profile data activity download-bulk \
      --format "$FORMAT" \
      --year "$year" \
      --month "$month" \
      --output "$tmpzip"; then
    echo "[$(date)] Extracting to $WATCHED..."
    unzip -o "$tmpzip" -d "$WATCHED/"
    echo "[$(date)] Done for $year/$month."
  else
    echo "[$(date)] No activities found for $year/$month, skipping."
  fi

  rm -f "$tmpzip"
}

download_and_extract "$YEAR" "$MONTH"

# On the 1st of the month also pull previous month to catch late syncs
if [ "$(date +%-d)" = "1" ]; then
  PREV_YEAR=$(date -d "last month" +%Y 2>/dev/null || date -v-1m +%Y)
  PREV_MONTH=$(date -d "last month" +%-m 2>/dev/null || date -v-1m +%-m)
  download_and_extract "$PREV_YEAR" "$PREV_MONTH"
fi
