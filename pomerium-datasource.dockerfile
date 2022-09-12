FROM golang:1.19.1-buster@sha256:d71125b5d1e299a8d9829eb09dbeb0cf576c461cbded210a9a6242d64072adc9  as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
RUN make build

FROM gcr.io/distroless/base-debian11@sha256:a08c76433d484340bd97013b5d868edfba797fbf83dc82174ebd0768d12f491d

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
