FROM golang:1.21.3-bookworm@sha256:deebfdaab16f8508e2590048d416b451e76ebe22e487cafd68990ff28f93ab1c as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:d2890b2740037c95fca7fe44c27e09e91f2e557c62cf0910d2569b0dedc98ddc

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
