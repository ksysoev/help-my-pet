FROM golang:1.23-alpine AS builder

ARG VERSION=${VERSION}
ARG BUILD=${BUILD}

WORKDIR /app/

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 go build -o /help-my-pet -ldflags "-X main.version=$VERSION -X main.build=$BUILD" ./cmd/help-my-pet

FROM scratch

COPY --from=builder /help-my-pet /help-my-pet
COPY config.yaml /config.yaml

ENTRYPOINT ["/help-my-pet"]
CMD ["bot"]s
