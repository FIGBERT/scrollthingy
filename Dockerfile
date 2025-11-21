FROM ghcr.io/gleam-lang/gleam:v1.13.0-erlang-alpine AS builder
COPY . /code
RUN apk --no-cache add just nodejs npm go linux-headers pkgconf x264-dev \
  && cd /code \
  && just build-prod

FROM scratch
COPY --from=builder /code/server/server /prod
ENTRYPOINT ["/prod"]
