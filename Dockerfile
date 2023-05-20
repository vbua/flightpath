FROM alpine:3.17

RUN apk update && apk upgrade libcrypto3 libssl3

COPY ./flightpath /usr/bin/flightpath

CMD ["flightpath"]
