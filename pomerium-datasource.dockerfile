FROM golang:1.21.3-bookworm@sha256:d0214956a9c50c300e430c1f6c0a820007ace238e5242c53762e61b344659e05 as build

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
