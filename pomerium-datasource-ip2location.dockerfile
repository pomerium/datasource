FROM ubuntu:latest@sha256:b59d21599a2b151e23eea5f6602f4af4d7d31c4e236d22bf0b62b86d2e386b8f AS curl

RUN apt-get update && apt-get install -y curl

WORKDIR /download
RUN --mount=type=secret,id=download_token \
    export DOWNLOAD_TOKEN=$(cat /run/secrets/download_token) && \
    curl --silent --show-error --fail \
    -o /download/IP2LOCATION-LITE-DB1.CSV.ZIP \
    "https://www.ip2location.com/download/?token=${DOWNLOAD_TOKEN}&file=DB1LITECSV"

FROM golang:1.24.4-bookworm@sha256:7b25b1ea217e0a56060953b3d4859134ecbe757d7434f7ce4756e0c25aad1ef0 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:7d1d72086ccf7b5c7e0f612dd59ae064765a529daafaecac97ea4a8b48b69e93

COPY --from=build /build/bin/* /bin/
COPY --from=curl /download/IP2LOCATION-LITE-DB1.CSV.ZIP /usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP

ENTRYPOINT ["/bin/pomerium-datasource", "ip2location", "/usr/share/IP2LOCATION-LITE-DB1.CSV.ZIP"]
