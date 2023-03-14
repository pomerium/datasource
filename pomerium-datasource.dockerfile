FROM golang:1.20.2-buster@sha256:57dbdd5c8fe24e357b15a4ed045b0b1607a99a1465b5101304ea39e11547be27  as build

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
