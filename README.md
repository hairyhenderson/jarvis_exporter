![Build](https://github.com/hairyhenderson/jarvis_exporter/workflows/Build/badge.svg)
![Docker Build](https://github.com/hairyhenderson/jarvis_exporter/workflows/Docker%20Build/badge.svg)
[![hairyhenderson/jarvis_exporter on DockerHub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/hairyhenderson/jarvis_exporter)

# jarvis_exporter

A Prometheus exporter for the Fully Jarvis standing desk. With a bit of simple
custom hardware, it can export the desk's current height, as a Prometheus
metric!

Based on the excellent reverse-engineering done by @phord at https://github.com/phord/Jarvis.

[Docker Images available](https://hub.docker.com/r/hairyhenderson/jarvis_exporter)
for Linux on x86-64, arm64, armv7 (Raspberry Pi), and armv6 (Pi Zero) platforms.

## Dashboard configuration

With `jarvis_exporter`, you can create a Grafana dashboard that looks like this:

![Grafana dashboard displaying stand/sit ratio and other metrics](dashboard.png)

This dashboard has been exported to [`dashboard/desk_metrics.json](./dashboard/desk_metrics.json).

Note that this dashboard in particular also uses some metrics exported by a
[simple script](./track_brightness.sh) that I wrote to export the display
brightness (and whether or not the display is asleep). It's scheduled to run
every minute via `cron`, and uses the [`brightness`](https://github.com/nriley/brightness)
tool to read the current brightness. The output is then scraped by
`node_exporter`'s [textfile collector](https://github.com/prometheus/node_exporter#textfile-collector).

## Hardware

_A description of the hardware I've put together is coming soon, but for now see
[this discussion](https://github.com/phord/Jarvis/discussions/8) for some hints!_

## TODO

- [ ] Support multiple UARTs for reading from both handset & desk (probably on
  a Raspberry Pi)

## License

[The MIT License](http://opensource.org/licenses/MIT)

Copyright (c) 2021-2023 Dave Henderson
