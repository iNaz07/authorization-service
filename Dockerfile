# syntax=docker/dockerfile:1
FROM golang:1.17-alpine as build

WORKDIR /app

COPY ./ ./

RUN go mod download

RUN --mount=type=cache,target=/root/.cache CGO_ENABLED=0 go build -o main ./app/main.go

FROM alpine:latest

WORKDIR /cmd

COPY --from=build /app/main ./
COPY --from=build /app/config.json ./

EXPOSE 8080

CMD ["./main"]