FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go test ./... && CGO_ENABLED=0 go build -ldflags="-s -w" -o /api .

FROM gcr.io/distroless/static-debian12
COPY --from=build /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
