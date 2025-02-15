FROM golang:1.23.1-alpine

WORKDIR /app

RUN apk add --no-cache bash curl

COPY . .

RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PATH=$PATH:/go/bin

RUN go build -o merch-store ./cmd/api

COPY start.sh /start.sh
RUN chmod +x /start.sh

EXPOSE 8080

CMD ["./start.sh"]
# CMD ["sh"]