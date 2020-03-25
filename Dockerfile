#
# Builder container
#

FROM golang:1.14 as builder
WORKDIR /go/src/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -installsuffix cgo .


#
# Runtime container
#

FROM scratch
# Certificates are needed to be able to use https from inside the container
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /go/src/pronestheus /
ENTRYPOINT ["/pronestheus"]