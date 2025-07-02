FROM golang:1.23.3-alpine

RUN apk add --no-cache git postgresql

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o app .

RUN go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

ENV PATH=$PATH:/go/bin

CMD ["./app"]
