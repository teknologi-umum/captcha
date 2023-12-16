FROM golang:1.21-bookworm AS builder

ARG PORT=8080

WORKDIR /app

COPY . .

RUN go build .

FROM debian:bookworm AS runtime

WORKDIR /app

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl ca-certificates openssl

COPY --from=builder /app/ .

EXPOSE ${PORT}

CMD [ "/app/teknologi-umum-bot" ]
