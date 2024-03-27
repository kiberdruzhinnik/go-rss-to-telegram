FROM golang:alpine as builder
WORKDIR /src/app
COPY . .
RUN go mod download && \
    go build -o go-rss-to-telegram cmd/go-rss-to-telegram/main.go

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /src/app/go-rss-to-telegram /app/go-rss-to-telegram
COPY --from=builder /src/app/config.yml.example /app/config.yml
ENTRYPOINT [ "./go-rss-to-telegram" ]