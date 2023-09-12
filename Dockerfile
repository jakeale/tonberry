FROM golang:1.26 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/tonberry ./cmd/tonberry

FROM gcr.io/distroless/static-debian12

COPY --from=build /out/tonberry /usr/local/bin/tonberry

ENTRYPOINT ["/usr/local/bin/tonberry"]
