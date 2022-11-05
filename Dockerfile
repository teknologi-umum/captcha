FROM golang:1.19.3-bullseye AS builder

ARG PORT=8080

WORKDIR /app

COPY . .

RUN go build -o captcha .

FROM debian:bullseye AS runtime

WORKDIR /app

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y curl ca-certificates openssl

COPY --from=builder /app/captcha /app/captcha

EXPOSE ${PORT}

CMD [ "/app/captcha" ]
