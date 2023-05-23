#Builder
FROM golang:1.20 AS build
WORKDIR /go/src
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app .

#APP DEPLOYED
FROM alpine:latest
ARG version=DEV
ENV APP_VERSION=$version
RUN echo "Bulding Docker image version: $APP_VERSION"
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=build /go/src/app .
EXPOSE 3000
CMD ["./app"]
