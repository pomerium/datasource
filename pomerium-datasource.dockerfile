FROM golang:1.20.5-buster@sha256:e8c88a10ec0e0be5677c6c2ebd12a321a5b7493d3d67c0d5c588884ed5a8794a  as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian11@sha256:df13a91fd415eb192a75e2ef7eacf3bb5877bb05ce93064b91b83feef5431f37

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
