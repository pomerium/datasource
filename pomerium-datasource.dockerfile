FROM golang:1.20.2-buster@sha256:b089a60efbf123ecfa2b0d8b955c9f441d83d08918dd358a7a8713e71c17c1d7  as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian11@sha256:a08c76433d484340bd97013b5d868edfba797fbf83dc82174ebd0768d12f491d

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
