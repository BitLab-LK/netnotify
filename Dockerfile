FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/netnotify ./cmd/netnotify
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/netnotify /usr/local/bin/netnotify
COPY templates /templates
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/netnotify"]
