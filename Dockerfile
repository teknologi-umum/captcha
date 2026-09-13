FROM golang:1.27.1-trixie@sha256:9baa6b4187bbb98d240372a8a235ac0bb6b5ddd52bba1431dc2f7c0705862728 AS builder

WORKDIR /app

COPY . .

RUN go build -o captcha-bot -ldflags="-X main.version=$(git rev-parse HEAD)" ./cmd/captcha

FROM debian:trixie-20260824-slim@sha256:d7e12182ce18b85b93007c1dedf31f2d29e01ccf3182cc4017c709b6259bc132 AS runtime

WORKDIR /app

ARG PORT=8080

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl ca-certificates openssl --no-install-recommends && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir -p /var/lib/captcha/badger

COPY . .

COPY --from=builder /app/captcha-bot /usr/local/bin/captcha

ENV ENVIRONMENT=production
ENV BADGER_PATH=/var/lib/captcha/badger

EXPOSE ${PORT}

CMD [ "/usr/local/bin/captcha" ]
