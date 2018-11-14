FROM golang:1.11-alpine as build

COPY . .

ENV GITLAB_TOKEN ""
ENV SENTRY_URL ""
ENV APP_HOST "0.0.0.0"
ENV APP_PORT 8000

RUN go build -i -o bin/releaser main.go

FROM alpine:latest as production
COPY --from=build /go/bin/releaser .
CMD ["./releaser"]
