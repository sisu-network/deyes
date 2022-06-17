FROM golang:1.18-alpine as builder

ENV GO111MODULE=on

WORKDIR /tmp/go-app

RUN apk add --no-cache make gcc musl-dev linux-headers git

# # Though the id_rsa file is removed at the end of this docker build, it's still dangerous to include
# # id_rsa in the build file since docker build steps are cached. Only do this while our repos are in
# # private mode.
# ADD tmp/id_rsa /root/.ssh/id_rsa

COPY go.mod .

COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/deyes main.go

# RUN rm /root/.ssh/id_rsa

# Start fresh from a smaller image
FROM alpine:3.9

WORKDIR /app

#Workaround: We shouldn't make .env mandatory, and the environment variables can be loaded from multiple places.
# RUN apk add ca-certificates \
#     && touch /app/.env && echo "#SAMPLE_KEY:SAMPLE_VALUE" > /app/.env

COPY --from=builder /tmp/go-app/out/deyes /app/deyes

CMD ["./deyes"]
