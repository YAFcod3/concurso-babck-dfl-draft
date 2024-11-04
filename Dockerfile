FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /my-app

FROM gcr.io/distroless/base-debian11 AS runtime

COPY --from=builder /my-app /my-app

EXPOSE 8000

ENTRYPOINT ["/my-app"]
