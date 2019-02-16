
FROM alpine:3.9

RUN apk add --no-cache ca-certificates
RUN set -eux; apk add --no-cache --virtual .build-deps  bash tzdata curl
RUN apk update
RUN apk upgrade
RUN rm -rf /var/cache/apk/*


RUN mkdir /app
COPY Readme.md /app

COPY bin/alpine/do-upper /app/

EXPOSE 3012

HEALTHCHECK --interval=1m --timeout=3s CMD curl -f http://0.0.0.0:3012/health || exit 1

WORKDIR /app

ENTRYPOINT ["/app/do-upper","serve"]
