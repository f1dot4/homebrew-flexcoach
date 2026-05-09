#!/bin/sh
# Downloads activities for the current (and optionally previous) month
# and extracts them into /watched for Dawarich to pick up.
set -e

YEAR=$(date +%Y)
MONTH=$(date +%-m)
DAY=$(date +%-d)
FORMAT="${FLEXCLI_FORMAT:-gpx}"
DAWARICH_USER="${FLEXCLI_DAWARICH_USER:?FLEXCLI_DAWARICH_USER env var is required}"
WATCHED="${FLEXCLI_WATCHED_DIR:-/watched}/${DAWARICH_USER}"
mkdir -p "$WATCHED"

tmpzip="/tmp/activities-$$-$(date +%s)"

echo "[$(date)] Downloading $YEAR/$MONTH/$DAY (format: $FORMAT)..."
if flexcli \
    --server "$FLEXCLI_SERVER" \
    --key "$FLEXCLI_API_KEY" \
    profile data activity download-bulk \
    --format "$FORMAT" \
    --year "$YEAR" \
    --month "$MONTH" \
    --day "$DAY" \
    --output "$tmpzip"; then
  echo "[$(date)] Extracting to $WATCHED..."
  unzip -o "$tmpzip" -d "$WATCHED/"
  chmod 644 "$WATCHED"/*.gpx 2>/dev/null || true
  echo "[$(date)] Done."
else
  echo "[$(date)] No activities found for $YEAR/$MONTH/$DAY, skipping."
fi

rm -f "$tmpzip"
