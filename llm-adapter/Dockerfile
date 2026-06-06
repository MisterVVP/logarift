FROM golang:1.25-bookworm AS build
WORKDIR /src/llm-adapter
COPY llm-adapter/go.mod ./
RUN go mod download
COPY llm-adapter/ ./
RUN CGO_ENABLED=0 go build -o /out/llm-adapter ./cmd/llm-adapter

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/llm-adapter /llm-adapter
EXPOSE 8091
ENTRYPOINT ["/llm-adapter"]
