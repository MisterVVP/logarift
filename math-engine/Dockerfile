FROM debian:bookworm AS build

RUN apt-get update \
    && apt-get install -y --no-install-recommends g++ make ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src/math-engine
COPY math-engine/ ./
RUN make OUT=/out/logarift-math-engine LDFLAGS="-static-libstdc++ -static-libgcc"

FROM debian:bookworm-slim

RUN useradd --system --uid 10002 --create-home --home-dir /var/lib/logarift-math logarift-math

COPY --from=build /out/logarift-math-engine /usr/local/bin/logarift-math-engine

USER logarift-math
WORKDIR /var/lib/logarift-math
EXPOSE 8090
ENV LOGARIFT_MATH_ENGINE_PORT=8090

HEALTHCHECK --interval=5s --timeout=3s --retries=12 --start-period=5s \
  CMD ["logarift-math-engine", "--healthcheck"]

ENTRYPOINT ["logarift-math-engine", "--serve"]
