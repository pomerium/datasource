FROM golang:1.24.3-bookworm@sha256:29d97266c1d341b7424e2f5085440b74654ae0b61ecdba206bc12d6264844e21 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:cc8cf191ad9028e8a2d2a88cf4d4ac8711dcaf679471590e176e04b818463bcf

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
