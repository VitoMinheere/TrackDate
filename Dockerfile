FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

# CGO disabled: modernc.org/sqlite is pure Go, so the result is a static
# binary with no C toolchain needed at build or run time.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/trackdate ./cmd/trackdate

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/trackdate /app/trackdate

ENV TRACKDATE_ADDR=:8080
ENV TRACKDATE_DB_PATH=/data/trackdate.db

VOLUME ["/data"]
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/trackdate"]
