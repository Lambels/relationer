FROM golang:1.18

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build ./cmd/relationer-server

CMD [ "./relationer-server" ]

EXPOSE 8080