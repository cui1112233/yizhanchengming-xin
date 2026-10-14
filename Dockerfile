# The public application image is the native Go service. The Go binary embeds
# the built React assets from backend/web/dist, so this image contains no Node
# application runtime or Express source.
FROM golang:1.23-bookworm AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
ARG QIANTIE_RELEASE_SHA
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/qiantie ./cmd/qiantie

FROM debian:bookworm-slim
ARG QIANTIE_RELEASE_SHA
LABEL org.opencontainers.image.revision=$QIANTIE_RELEASE_SHA \
      org.opencontainers.image.source="https://github.com/cui1112233/-"
ENV QIANTIE_RELEASE_SHA=$QIANTIE_RELEASE_SHA
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/qiantie /qiantie
EXPOSE 3000
ENTRYPOINT ["/qiantie"]
