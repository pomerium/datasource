FROM ubuntu:latest@sha256:f9d633ff6640178c2d0525017174a688e2c1aef28f0a0130b26bd5554491f0da as curl

RUN apt-get update && apt-get install -y curl

WORKDIR /download
RUN --mount=type=secret,id=download_token \
    export DOWNLOAD_TOKEN=$(cat /run/secrets/download_token) && \
    curl --silent --show-error --fail \
    -o /download/IP2LOCATION-LITE-DB1.CSV.ZIP \
    "https://www.ip2location.com/download/?token=${DOWNLOAD_TOKEN}&file=DB1LITECSV"

FROM golang:1.22.0-bookworm@sha256:925fe3fa28ba428cf67a7947ae838f8a1523117b40e3e6b5106c378e3f97fa29 as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:1d91d5f201f4be1ab11bdbcd3ef9d1a42bdfd29863bb007eba11eaa55172f85d

COPY --from=build /build/bin/* /bin/
COPY --from=curl /download/IP2LOCATION-LITE-DB1.CSV.ZIP /usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP

ENTRYPOINT ["/bin/pomerium-datasource", "ip2location", "/usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP"]
