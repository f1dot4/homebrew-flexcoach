FROM alpine:3.21

RUN apk add --no-cache unzip tzdata

COPY bin/flexcli-linux /usr/local/bin/flexcli
COPY scripts/import-activities.sh /usr/local/bin/import-activities

RUN chmod +x /usr/local/bin/flexcli /usr/local/bin/import-activities

COPY scripts/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

ENV TZ=Europe/Vienna

ENTRYPOINT ["/docker-entrypoint.sh"]
