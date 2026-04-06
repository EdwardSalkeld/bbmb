FROM scratch

COPY server/bbmb-server-linux-amd64 /bbmb-server

EXPOSE 9876 9877

ENTRYPOINT ["/bbmb-server"]
