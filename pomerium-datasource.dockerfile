FROM golang:1.23.2-bookworm@sha256:18d2f940cc20497f85466fdbe6c3d7a52ed2db1d5a1a49a4508ffeee2dff1463 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:662eaa2606087124ac1fc108291ecad341f6376ce6fa28ac7e1655ec76c6e6d9

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
