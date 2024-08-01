FROM ubuntu:latest@sha256:2e863c44b718727c860746568e1d54afd13b2fa71b160f5cd9058fc436217b30 as curl

RUN apt-get update && apt-get install -y curl

WORKDIR /download
RUN --mount=type=secret,id=download_token \
    export DOWNLOAD_TOKEN=$(cat /run/secrets/download_token) && \
    curl --silent --show-error --fail \
    -o /download/IP2LOCATION-LITE-DB1.CSV.ZIP \
    "https://www.ip2location.com/download/?token=${DOWNLOAD_TOKEN}&file=DB1LITECSV"

FROM golang:1.22.5-bookworm@sha256:af9b40f2b1851be993763b85288f8434af87b5678af04355b1e33ff530b5765f as build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:af772ed0ce52d8994acedc3ec84a9d22e9366dda8767f17d1bb2213b06beaff5

COPY --from=build /build/bin/* /bin/
COPY --from=curl /download/IP2LOCATION-LITE-DB1.CSV.ZIP /usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP

ENTRYPOINT ["/bin/pomerium-datasource", "ip2location", "/usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP"]
