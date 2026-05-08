FROM golang:1.26.3-bookworm@sha256:b09e568dcf2a1ff3ce09a230ea234193fc014dc195472fe63316e50238453d96 AS build

WORKDIR /build

COPY Makefile ./Makefile

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY ./pkg/ ./pkg/
RUN make build

FROM gcr.io/distroless/base-debian12:debug@sha256:fd8e2df55ce1400d6f651f14a030fd38868fca0e1ab6989bc1641cd5fc0f3335

COPY --from=build /build/bin/* /bin/

ENTRYPOINT ["/bin/pomerium-datasource"]
