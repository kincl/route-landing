FROM golang:1.18 as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY assets assets

RUN ls -la

RUN go build -o route-landing

FROM registry.access.redhat.com/ubi8-minimal:8.5

COPY --from=build /app/route-landing /route-landing

EXPOSE 8080

ENTRYPOINT ["/route-landing"]
