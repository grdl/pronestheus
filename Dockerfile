# This Dockerfile is intended to be used with goreleaser.
# It doesn't build the executable, it expects it to be already built by the goreleaser.
# Base image is based on official node-exporter Dockerfile.

FROM quay.io/prometheus/busybox:glibc
COPY pronestheus /
USER nobody
EXPOSE 9777
ENTRYPOINT ["/pronestheus"]
