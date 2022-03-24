FROM golang:1.18.0-bullseye

ARG CERT_URL

ARG PORT=8080

WORKDIR /app

RUN curl --create-dirs -o ./.postgresql/root.crt -O ${CERT_URL}

COPY . .

RUN go mod download

RUN go build main.go

EXPOSE ${PORT}

CMD [ "./main" ]
