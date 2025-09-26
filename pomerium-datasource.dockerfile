FROM golang:1.25.1-bookworm@sha256:c4bc0741e3c79c0e2d47ca2505a06f5f2a44682ada94e1dba251a3854e60c2bd AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:0d30cde8cd133aa8b8a967bb22bdb3fbfeb4daf413cce46b7149c4d4672dfe1e

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
