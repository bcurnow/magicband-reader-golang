# magicband-reader-golang

A re-implementation of the [magicband-reader](https://github.com/bcurnow/magicband-reader) in Go.

Runs on a Raspberry Pi (built and tested for the Pi Zero, armv6) with an MFRC522 RFID reader, a
ring of WS281x addressable LEDs, and a speaker. It reads RFID tag UIDs, checks them against
[rfid-security-svc](https://github.com/bcurnow/rfid-security-svc) for authorization, and gives
audio/visual feedback (sound + LED effects) for read/authorized/unauthorized events.

## Hardware

- Raspberry Pi (armv6, e.g. Pi Zero/Zero W)
- MFRC522 RFID reader, connected via SPI
- WS281x addressable LED ring(s), connected via PWM/GPIO
- A speaker/audio output (ALSA)

## How it works

1. `reader.go` polls the MFRC522 over SPI for a tag UID.
2. Each UID is dispatched through a chain of handlers (`handler/`), registered by priority:

   | Priority | Handler       | Does                                              |
   |----------|---------------|----------------------------------------------------|
   | 10       | `readSound`   | Plays the "read" sound                              |
   | 11       | `spin`        | Starts the LED spin effect                          |
   | 12       | `authorize`   | Calls rfid-security-svc to authorize the UID        |
   | 13       | `stopSpin`    | Stops the LED spin effect                           |
   | 20       | `showStatus`  | Fades the LEDs to a color based on the result       |
   | 21       | `authSound`   | Plays the authorized/unauthorized sound             |
   | 22       | `stopStatus`  | Fades the LEDs back off                             |
   | last     | `logging`     | Logs the final result                               |

3. A small HTTP server (`/get_uid`, see `router.go`) lets an external caller (e.g.
   rfid-security-svc itself) long-poll for the next UID read instead of relying on the handler
   chain, with an optional `?timeout=<seconds>` query parameter (default 60s).

## Configuration

Configuration is handled by [peterbourgon/ff](https://github.com/peterbourgon/ff): every flag can
also be set via an environment variable (prefixed `MR_`, dashes become underscores - e.g.
`--api-key` is `MR_API_KEY`) or via a YAML config file (`--config-file`, default
`/etc/magicband-reader/magicband-reader.yml`).

| Flag                  | Default                                    | Description                                                                                    |
|------------------------|---------------------------------------------|--------------------------------------------------------------------------------------------------|
| `--api-key`            | *(none)*                                    | API key to authenticate to rfid-security-svc                                                     |
| `--api-ssl-verify`     | `ca.pem`                                    | A CA cert file path to validate the rfid-security-svc connection against, or `false` to skip validation entirely (insecure). Cannot be set to `true`. |
| `--api-url`            | `https://localhost:5000/api/v1.0`           | rfid-security-svc base URL                                                                       |
| `--authorized-sound`   | `authorized.wav`                            | Sound played when a band is authorized (relative to `--sound-dir`)                               |
| `--brightness`         | `100`                                       | LED brightness, 0-255                                                                            |
| `--config-file`        | `/etc/magicband-reader/magicband-reader.yml`| YAML config file to load (optional)                                                              |
| `--inner-ring-size`    | `20`                                        | Number of LEDs in the inner ring                                                                 |
| `--listen-address`     | `localhost`                                 | Address the `/get_uid` HTTP server listens on                                                     |
| `--listen-port`        | `8080`                                      | Port the `/get_uid` HTTP server listens on                                                        |
| `--log-level`          | `info`                                      | `debug`, `info`, `warning`, `error`, `fatal`                                                      |
| `--log-report-caller`  | `false`                                     | Include calling function/file/line in log output (only at `trace` level)                          |
| `--outer-ring-size`    | `40`                                        | Number of LEDs in the outer ring                                                                  |
| `--permission`         | `MagicBand Reader`                          | Permission name to validate against rfid-security-svc                                             |
| `--read-sound`         | `read.wav`                                  | Sound played when a band is read (relative to `--sound-dir`)                                      |
| `--sound-dir`          | `/sounds`                                   | Directory containing sound files                                                                 |
| `--unauthorized-sound` | `unauthorized.wav`                          | Sound played when a band is unauthorized (relative to `--sound-dir`)                              |
| `--volume-level`       | `0`                                         | Positive/negative adjustment applied to the base volume                                          |

## Building

All builds cross-compile for `linux/arm/v6` (CGO is required for the LED and audio drivers).

- `make build-docker` - builds the production Docker image
- `make build-docker-dev` - builds the dev Docker image (build toolchain, no app binary)
- `make build` - builds the binary directly to `bin/magicband-reader` (must be run inside an
  environment with the arm/v6 C toolchain and `libws2811`/`libasound2-dev` available - the dev
  Docker image provides this, see `make dev` below)

## Running

- `make dev` - starts the dev Docker image with the repo bind-mounted at `/workspace`, `/dev` and
  a local `sounds/` directory mounted in, for building/testing against real hardware
- `make prod` - runs the production image in the background, same device/sounds mounts
- `make run` - runs `bin/magicband-reader` directly (pass extra flags via `READER_ARGS`)

## Development

- `make local` - clears the screen, `go mod tidy`, formats, vets, tests, and builds
- `make lr` - `local` followed by `run`
- `make format` - `gofmt -l -w -s .`
- `make vet` - `go vet`
- `make lint` - `golangci-lint run`
- `make test` - `go test`
- `make tidy` - `go mod tidy`
- `make clean` - removes `bin/`

CI (GitHub Actions) runs the same vet/build/test/lint checks plus a full `linux/arm/v6` Docker
build on every push and PR; CodeQL scans run separately. See `.github/workflows/`.

## License

[Apache License 2.0](LICENSE)
