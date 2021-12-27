FROM golang:1.16-alpine as builder

ENV GO111MODULE=on \
    GOPRIVATE=github.com/sisu-network/*

WORKDIR /tmp/go-app

RUN apk add --no-cache make gcc musl-dev linux-headers git \
    && apk add openssh

RUN mkdir -p -m 0600 /root/.ssh \
    && ssh-keyscan github.com >> ~/.ssh/known_hosts \
    && git config --global url."git@github.com:".insteadOf "https://github.com/"


COPY go.mod go.sum ./

RUN --mount=type=ssh go mod download

COPY . .

RUN go build -o ./out/deyes main.go

# second stage
FROM alpine:3.9

WORKDIR /app

#Workaround: We shouldn't make .env mandatory, and the environment variables can be loaded from multiple places.
# RUN apk add ca-certificates \
#     && touch /app/.env && echo "#SAMPLE_KEY:SAMPLE_VALUE" > /app/.env

COPY --from=builder /tmp/go-app/out/deyes /app/deyes
COPY --from=builder /tmp/go-app/migrations /app/migrations

CMD ["./deyes"]
