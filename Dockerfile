FROM golang:1.16-alpine AS builder

RUN apk --no-cache add make

WORKDIR /opt/consuldemo
COPY . .

RUN make build

# ---

FROM alpine:3.13

WORKDIR /opt/consuldemo

COPY entrypoint.sh /
COPY --from=builder /opt/consuldemo/bin/ /opt/consuldemo/bin/
COPY --from=builder /opt/consuldemo/tls_ca_cert.pem /opt/consuldemo/
COPY --from=builder /opt/consuldemo/tls_cert.pem /opt/consuldemo/
COPY --from=builder /opt/consuldemo/tls_key.key /opt/consuldemo/

RUN apk --no-cache add tini tzdata && \
        chmod +x /entrypoint.sh /opt/consuldemo/bin/*

ENV PATH /opt/consuldemo/bin:$PATH

ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]
