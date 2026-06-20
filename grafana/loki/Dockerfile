FROM grafana/loki:latest

COPY loki-config.yaml /etc/loki/loki.yaml
USER root

ENTRYPOINT ["/usr/bin/loki"]
CMD ["-config.file=/etc/loki/loki.yaml"]
