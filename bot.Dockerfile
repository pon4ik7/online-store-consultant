FROM golang:1.24-alpine

RUN apk add --no-cache bash git postgresql-client

WORKDIR /app

COPY bot/go.mod bot/go.sum ./bot/
RUN cd bot && go mod download

COPY . .

RUN cd bot && go build -o /app/telegram-bot .

COPY entrypoint-bot.sh /entrypoint-bot.sh
RUN chmod +x /entrypoint-bot.sh

ENTRYPOINT ["/entrypoint-bot.sh"]