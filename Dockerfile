FROM golang:1.17.1-buster

ARG CERT_URL

RUN curl --create-dirs -o $HOME/.postgresql/root.crt -O ${CERT_URL}

WORKDIR /usr/app

COPY . .

RUN go mod download

RUN go build main.go

CMD [ "./main" ]
