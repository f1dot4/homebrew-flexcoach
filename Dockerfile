FROM alpine:3.21

LABEL org.opencontainers.image.title="flexcli" \
      org.opencontainers.image.description="FlexCLI sidecar — downloads activities from the FlexCoach platform and places them into a Dawarich watched directory for automatic import. Runs on a configurable cron schedule." \
      org.opencontainers.image.source="https://github.com/f1dot4/homebrew-flexcli" \
      org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache unzip tzdata

COPY bin/flexcli-linux /usr/local/bin/flexcli
COPY scripts/import-activities.sh /usr/local/bin/import-activities

RUN chmod +x /usr/local/bin/flexcli /usr/local/bin/import-activities

COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENV TZ=Europe/Vienna

ENTRYPOINT ["/docker-entrypoint.sh"]
