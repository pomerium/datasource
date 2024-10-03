FROM ubuntu:latest@sha256:dfc10878be8d8fc9c61cbff33166cb1d1fe44391539243703c72766894fa834a AS curl

RUN apt-get update && apt-get install -y curl

WORKDIR /download
RUN --mount=type=secret,id=download_token \
    export DOWNLOAD_TOKEN=$(cat /run/secrets/download_token) && \
    curl --silent --show-error --fail \
    -o /download/IP2LOCATION-LITE-DB1.CSV.ZIP \
    "https://www.ip2location.com/download/?token=${DOWNLOAD_TOKEN}&file=DB1LITECSV"

FROM golang:1.23.2-bookworm@sha256:18d2f940cc20497f85466fdbe6c3d7a52ed2db1d5a1a49a4508ffeee2dff1463 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:662eaa2606087124ac1fc108291ecad341f6376ce6fa28ac7e1655ec76c6e6d9

COPY --from=build /build/bin/* /bin/
COPY --from=curl /download/IP2LOCATION-LITE-DB1.CSV.ZIP /usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP

ENTRYPOINT ["/bin/pomerium-datasource", "ip2location", "/usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP"]
