FROM golang:1.17.3-alpine

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./service cmd/service/main.go

FROM alpine
ENV PORT=8000

COPY --from=0 /src/service /usr/bin/service
CMD ["service"]