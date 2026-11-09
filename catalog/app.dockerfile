FROM golang:1.27-alpine AS build
RUN apk add --no-cache ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY catalog catalog
RUN go build -o /out/app ./catalog/cmd/catalog

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
WORKDIR /usr/bin
COPY --from=build /out/app ./app
EXPOSE 8080
CMD ["app"]