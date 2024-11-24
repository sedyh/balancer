FROM golang:1.23.3-alpine AS builder
WORKDIR /app

COPY go.mod go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o program ./cmd/balancer

FROM scratch

COPY --from=builder /app/program /
COPY balancer.env* .

ENTRYPOINT ["./program"]
