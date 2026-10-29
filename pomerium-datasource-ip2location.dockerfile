FROM ubuntu:latest@sha256:da6fc2be547864451aa253836dd926da33623312df4a9a243e35dc877c378a78 AS curl

RUN apt-get update && apt-get install -y curl

WORKDIR /download
RUN --mount=type=secret,id=download_token \
    export DOWNLOAD_TOKEN=$(cat /run/secrets/download_token) && \
    curl --silent --show-error --fail \
    -o /download/IP2LOCATION-LITE-DB1.CSV.ZIP \
    "https://www.ip2location.com/download/?token=${DOWNLOAD_TOKEN}&file=DB1LITECSV"

FROM golang:1.27.3-bookworm@sha256:d9b6a1beb9474adc2a1ce842573639b44b2a0dd49b6669f51d3821baf9ec37c9 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:0a5b69081973d7069cbeab49c24851dd517887767e934fc8e6f808ea08770ef4

COPY --from=build /build/bin/* /bin/
COPY --from=curl /download/IP2LOCATION-LITE-DB1.CSV.ZIP /usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP

ENTRYPOINT ["/bin/pomerium-datasource", "ip2location", "/usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP"]
