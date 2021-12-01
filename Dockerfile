FROM golang:1.17.1-buster

ARG CERT_URL

WORKDIR /usr/app

RUN curl --create-dirs -o ./.postgresql/root.crt -O ${CERT_URL}

COPY . .

RUN go mod download

RUN go build main.go

CMD [ "./main" ]
