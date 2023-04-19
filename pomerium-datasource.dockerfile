FROM golang:1.20.3-buster@sha256:73c225bc5e2353f20dbe0466819b70a51a114a93bfe4af035a3bb9e1ecdd4107  as build

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
