FROM golang:1.25 AS build

WORKDIR /src/backend

# Copy the full backend source before downloading modules.
# The repository may not contain a pre-generated go.sum in early archives,
# so downloading after the full copy prevents an empty host go.sum from
# overwriting the generated one before the build step.
COPY backend/ ./
RUN go mod download
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
