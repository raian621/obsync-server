# build stage
FROM golang:1.22-alpine AS build
RUN apk add --no-cache --update go gcc g++
WORKDIR /app
COPY . .
RUN ./scripts/download_redoc_bundle.sh
RUN go build .

# packaging stage
FROM golang:1.22-alpine AS packaging
WORKDIR /app
COPY --from=build /app/obsync-server .
RUN mkdir /var/obsync
COPY <<EOF ./config.yaml
type: FileSystem
root: /var/obsync
host: 0.0.0.0
port: 8000
EOF
CMD ["./obsync-server"]