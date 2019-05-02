FROM golang:latest AS build

RUN mkdir /home/app
ADD . /home/app

WORKDIR /home/app

RUN go build --mod=vendor ./cmd/forum/main.go

FROM serega753/bd_hw:latest AS release

COPY --from=build /home/app/main /home/app/main

#RUN true

COPY --from=build /home/app/config/config.json /home/app/

USER root

EXPOSE 5000

CMD service postgresql start && /home/app/main /home/app/config.json