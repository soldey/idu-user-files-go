FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV APP_ENV production

RUN apk update --no-cache

WORKDIR /build

ADD go.mod .
ADD go.sum .

RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /app/main ./main.go

FROM alpine

ENV APP_ENV production

RUN apk update --no-cache

WORKDIR /app

COPY ./.env.${APP_ENV} .
COPY --from=builder /app/main /app/main

EXPOSE 7001

CMD ["./main"]