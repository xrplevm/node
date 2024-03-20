FROM golang:1.20 AS base
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENTRYPOINT ["/bin/sh", "-c"]
CMD ["go test github.com/Peersyst/exrp/tools/contracts-tester"]