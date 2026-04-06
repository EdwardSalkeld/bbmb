FROM debian:trixie-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && rm -rf /var/lib/apt/lists/*

COPY server/bbmb-server-linux-amd64 /usr/local/bin/bbmb-server

EXPOSE 9876 9877

HEALTHCHECK --interval=5s --timeout=2s --retries=3 \
    CMD curl -fsS http://127.0.0.1:9877/metrics > /dev/null || exit 1

ENTRYPOINT ["bbmb-server"]
