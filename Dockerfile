FROM golang:1.21-bookworm AS builder

WORKDIR /app

COPY . .

RUN go build -o teknologi-umum-captcha .

FROM debian:bookworm-slim AS runtime

WORKDIR /app

ARG PORT=8080

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl ca-certificates openssl

COPY . .

COPY --from=builder /app/teknologi-umum-captcha .

EXPOSE ${PORT}

CMD [ "/app/teknologi-umum-captcha" ]
