FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod .
COPY go.sum .

COPY . .

RUN go mod download

RUN go build -o go_docker ./cmd

EXPOSE 8080

CMD ["./go_docker"]