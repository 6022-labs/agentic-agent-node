FROM grafana/mimir:latest

COPY mimir-config.yaml /etc/mimir/mimir.yml

CMD ["--config.file=/etc/mimir/mimir.yml", "--target=all"]
