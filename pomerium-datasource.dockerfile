FROM golang:1.26.5-bookworm@sha256:18aedc16aa19b3fd7ded7245fc14b109e054d65d22ed53c355c899582bbb2113 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:b2a854c5f5b6d9441084b66628335fb9c66ae2ee93d719746b60ff1add99654a

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
