FROM golang:1.26-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o ./dist/app .

FROM alpine:3.23
WORKDIR /app
COPY --from=build /app/dist/app .
EXPOSE 8080
CMD ["./app"]