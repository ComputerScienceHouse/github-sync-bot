FROM docker.io/golang:1.25-alpine AS build

RUN mkdir /opt/sync

WORKDIR /opt/sync

RUN apk add git 

COPY go.* .
RUN go mod download 
COPY . .
RUN go build -v -o sync

FROM docker.io/alpine
WORKDIR /opt/sync
RUN apk add --no-cache tzdata 
ENV TZ=America/New_York
RUN cp /usr/share/zoneinfo/America/New_York /etc/localtime 
COPY --from=build /opt/sync/sync sync

ENTRYPOINT [ "./sync" ]

