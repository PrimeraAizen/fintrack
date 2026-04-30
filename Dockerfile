FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app cmd/web/main.go

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/app /app/app
COPY config /app/config
COPY migrations /app/migrations
ENV APP_HTTP_HOST=0.0.0.0
EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/app"]

