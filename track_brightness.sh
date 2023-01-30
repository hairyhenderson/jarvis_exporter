#!/bin/sh
#
# This script is used to track the brightness of the display on a Mac.
# It is intended to be run by cron every minute.
#
# The script uses the brightness command from https://github.com/nriley/brightness
# to get the current brightness of the display. It then writes the value to a
# Prometheus-format file which is scraped by the node_exporter's textfile
# collector.
#
# See https://github.com/prometheus/node_exporter#textfile-collector

awake="$(/usr/local/bin/brightness -l | head -n1 | grep awake)"
brightness="$(/usr/local/bin/brightness -l | awk '/display 0: brightness/{print $4}')"
state="awake"
if [ -z "${awake}" ]; then
    state="asleep"
fi
echo "macos_display_brightness_percent{state=\"$state\"} $brightness" > /usr/local/etc/textcollector/brightness.prom
