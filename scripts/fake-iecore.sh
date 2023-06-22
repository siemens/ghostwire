#!/bin/bash
DIR=$(dirname "$0")

COREIMGNAME="fake-ied-core"
CORENAME="edge-iot-core"

FAKEAPPNAME="myapp_fake_1"


# Starts/stops a fake IE core runtime container (incl. fake platformbox.db) as
# well as a fake IE app project.

function start() {
    docker build -t ${COREIMGNAME} ${DIR}/../decorator/ieappicon/tests/fakeied
    stop
    docker run -d \
        --name ${CORENAME} \
        --entrypoint /bin/sh \
        ${COREIMGNAME} -c "sleep 365d"
    docker run -d \
        --name ${FAKEAPPNAME} \
        --label "com_mwp_conf_=" \
        --label "com.docker.compose.project=bbb" \
        --entrypoint /bin/sh \
        busybox -c "sleep 365d"
}

function stop() {
    docker rm -f ${CORENAME}
    docker rm -f ${FAKEAPPNAME}
}

case "$1" in
    start) start;;
    stop) stop;;
    restart) stop; start;;
    *)
        echo "usage: $0 start|stop|restart"
        exit 1
        ;;
esac
