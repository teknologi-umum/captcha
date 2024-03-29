FROM golang:1.22-bookworm AS builder

WORKDIR /app

COPY . .

RUN go build -o captcha-bot -ldflags="-X main.version=$(git rev-parse HEAD)" ./cmd/captcha

FROM debian:bookworm-slim AS runtime

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
