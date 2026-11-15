FROM golang:1.26.4 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/chat .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/chat ./chat
COPY web/ ./web/
EXPOSE 8080
CMD ["./chat"]