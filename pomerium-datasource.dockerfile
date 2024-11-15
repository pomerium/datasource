FROM golang:1.23.2-bookworm@sha256:2341ddffd3eddb72e0aebab476222fbc24d4a507c4d490a51892ec861bdb71fc AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:29160befd6389ab5784858ad88459281bac1dcd739ab74e3e80d1e1a23e3987e

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
