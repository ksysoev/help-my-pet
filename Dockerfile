FROM golang:1.21-alpine AS builder

ARG VERSION=${VERSION}
ARG BUILD=${BUILD}

WORKDIR /app/

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 go build -o /help-my-pet -ldflags "-X main.version=$VERSION -X main.build=$BUILD" ./cmd/help-my-pet

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

COPY --from=builder /help-my-pet /app/help-my-pet
COPY config.yaml /app/config.yaml

ENTRYPOINT ["/app/help-my-pet"]
CMD ["bot"]
