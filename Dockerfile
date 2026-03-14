FROM node:20-alpine AS web-build
WORKDIR /web
COPY ./web ./
RUN npm ci && npm run build

FROM golang:1.26-alpine AS build
WORKDIR /app
COPY . .
COPY --from=web-build /web/dist ./web/dist
#constantly embed frontend for now
RUN go build -tags=embed_frontend -o ./dist/app ./cmd/app

FROM alpine:3.23
COPY --from=build /app/dist/ .
EXPOSE 8080
ENV APP_DB=pgsql
CMD ["./app"]