FROM golang:1.23-bookworm AS build

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum* ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o /out/logarift-api ./cmd/api

FROM debian:bookworm-slim

RUN useradd --system --uid 10001 --create-home --home-dir /var/lib/logarift logarift \
    && mkdir -p /var/lib/logarift/exports \
    && chown -R logarift:logarift /var/lib/logarift

COPY --from=build /out/logarift-api /usr/local/bin/logarift-api

USER logarift
WORKDIR /var/lib/logarift
EXPOSE 8080

ENTRYPOINT ["logarift-api"]
