# syntax=docker/dockerfile:1
FROM golang:1.22 as build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd server-analyst && go mod tidy && make build

FROM gcr.io/distroless/base-debian12
WORKDIR /opt/server-analyst
COPY --from=build /src/server-analyst/server-analyst /opt/server-analyst/server-analyst
COPY server-analyst/config /etc/server-analyst
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s CMD ["/opt/server-analyst/server-analyst","health"]
ENTRYPOINT ["/opt/server-analyst/server-analyst","-config","/etc/server-analyst/config.yaml"]
