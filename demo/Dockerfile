FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./demo/main

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=builder /app/server .

USER nonroot:nonroot

ENTRYPOINT ["./server"]