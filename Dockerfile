FROM golang:1.20-bullseye AS builder

ARG PORT=8080

WORKDIR /app

COPY . .

RUN go build .

FROM debian:bullseye AS runtime

WORKDIR /app

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl ca-certificates openssl

COPY --from=builder /app/ .

EXPOSE ${PORT}

CMD [ "/app/teknologi-umum-bot" ]
