FROM golang:1.24.4-bookworm@sha256:7b25b1ea217e0a56060953b3d4859134ecbe757d7434f7ce4756e0c25aad1ef0 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:7d1d72086ccf7b5c7e0f612dd59ae064765a529daafaecac97ea4a8b48b69e93

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
