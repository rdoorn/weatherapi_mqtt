#!/bin/bash

case $1 in
    start)
        docker run --name=%NAME% \
        --net=domoticanet \
        --restart=unless-stopped \
        -d \
        -e TZ=Europe/Amsterdam \
        -e WEATHERAPI_API="" \
        -e WEATHERAPI_LAT="52.6596" \
        -e WEATHERAPI_LONG="4.8283" \
        -e MQTT_URL="mqtt://mosquitto:1883/"
        %NAME%
        ;;
    stop)
        docker stop %NAME% | xargs docker rm
        ;;
    *)
        echo "unknown or missing parameter $1"
esac
