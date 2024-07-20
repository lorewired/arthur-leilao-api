FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o out ./cmd

FROM ubuntu:jammy

WORKDIR /app

COPY --from=build /app/out /app/out

CMD ["/app/out"]
