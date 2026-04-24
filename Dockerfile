FROM golang:1.24 AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/winston .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/winston /app/winston
ENTRYPOINT ["/app/winston"]
