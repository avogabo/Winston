FROM node:22-bookworm AS webbuild
WORKDIR /src/web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

FROM golang:1.24 AS build
WORKDIR /src
COPY . .
COPY --from=webbuild /src/web/dist /src/web/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/winston .

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/winston /app/winston
EXPOSE 8091
ENTRYPOINT ["/app/winston"]
