FROM golang:1.18-bullseye AS builder

ARG CERT_URL

WORKDIR /app

RUN curl --create-dirs -o ./.postgresql/root.crt -O ${CERT_URL}

COPY . .

RUN go mod download

RUN go build main.go

FROM debian:bullseye

WORKDIR /app

COPY --from=builder /app .

ARG PORT=8080

EXPOSE ${PORT}

CMD [ "./main" ]
