FROM golang:1.20.4-alpine AS build

ARG SVC_PATH

WORKDIR /src/
COPY . /src/
RUN CGO_ENABLED=0 go build -o /bin/svc $SVC_PATH/main.go

FROM scratch
COPY --from=build /bin/svc /bin/svc
ENTRYPOINT ["/bin/svc"]
