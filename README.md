# PikoMirror

![](_docs/mirror.png)

High-performance storage for traffic for [http_mirror](https://nginx.org/en/docs/http/ngx_http_mirror_module.html).

Easy, hackable, and ships with web UI.

The project is under heavy development:

- not all features yet implemented,
- no documentation.

![Screenshot from 2024-06-03 22-59-23](https://github.com/pikocloud/pikomirror/assets/6597086/f43f70ec-dd90-4c61-b71a-1cc87877db01)


## Installation

- Binary in [releases](releases)
- Docker image

```
ghcr.io/pikocloud/pikomirror:latest
```

See example in [directory](dev/docker-compose.yaml)

## Roadmap

- [ ] S3 storage for blobs
- [ ] Internal authorization (currently use something like oauth-proxy)
- [ ] Enhanced TLS
- [ ] Pass-through mode