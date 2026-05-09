#!/bin/sh
# Downloads activities for the current (and optionally previous) month
# and extracts them into /watched for Dawarich to pick up.
set -e

YEAR=$(date +%Y)
MONTH=$(date +%-m)
DAY=$(date +%-d)
FORMAT="${FLEXCLI_FORMAT:-gpx}"
WATCHED="${FLEXCLI_WATCHED_DIR:-/watched}"

tmpzip=$(mktemp /tmp/activities-XXXXXX)

echo "[$(date)] Downloading $YEAR/$MONTH/$DAY (format: $FORMAT)..."
if flexcli profile data activity download-bulk \
    --format "$FORMAT" \
    --year "$YEAR" \
    --month "$MONTH" \
    --day "$DAY" \
    --output "$tmpzip"; then
  echo "[$(date)] Extracting to $WATCHED..."
  unzip -o "$tmpzip" -d "$WATCHED/"
  echo "[$(date)] Done."
else
  echo "[$(date)] No activities found for $YEAR/$MONTH/$DAY, skipping."
fi

rm -f "$tmpzip"
