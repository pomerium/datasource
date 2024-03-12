FROM golang:1.22.1-bookworm@sha256:5d60d1b62db209d4fe8a060f8d15374482e5258cc0f92491b3901e20aa5dda8f as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:e0cc8fa0ed6c46f7f019678218f8b7efdc7df09638ee49f586fb4f0fdf8b09ae

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
