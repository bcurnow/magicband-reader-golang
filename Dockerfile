ARG GO_VERSION=1.25.12
ARG DEBIAN_VERSION=bookworm

FROM debian:${DEBIAN_VERSION} AS lib_builder

WORKDIR /foundry

RUN apt-get update && apt-get install --no-install-recommends -y \
      build-essential \
      ca-certificates \
      cmake \
      git \
    && git clone https://github.com/jgarff/rpi_ws281x.git \
    && cd rpi_ws281x \
    && mkdir build \
    && cd build \
    && cmake -D BUILD_SHARED=OFF -D BUILD_TEST=OFF .. \
    && cmake --build . \
    && make install

FROM debian:${DEBIAN_VERSION} AS dev_image

# The official golang image no longer publishes an arm/v5-or-lower manifest entry on
# Debian bookworm tags (only linux/arm/v7 remains), which breaks
# `docker buildx build --platform linux/arm/v6` (the Raspberry Pi Zero's architecture:
# QEMU needs an actual arm/v6-compatible image to emulate, since CGO_ENABLED=1 requires
# a native C toolchain, not just Go's own cross-compiler). Go still publishes a
# linux-armv6l release tarball directly, so install that by hand instead of depending on
# the golang image's shrinking platform matrix. If GO_VERSION is bumped, update
# GO_ARMV6_SHA256 to match (from https://go.dev/dl/?mode=json).
ARG GO_VERSION
ARG GO_ARMV6_SHA256=6cd7311c02c73ba0b482a1cf8c885268edf23519261bf4b5cef3353ad934d1f1

RUN apt-get update \
    && apt-get install --no-install-recommends -y ca-certificates curl \
    && curl -fsSLo /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-armv6l.tar.gz" \
    && echo "${GO_ARMV6_SHA256}  /tmp/go.tar.gz" | sha256sum -c - \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/local/go/bin:${PATH}"

COPY --from=lib_builder /usr/local/lib/libws2811.a /usr/local/lib/
COPY --from=lib_builder /usr/local/include/ws2811 /usr/local/include/ws2811

ARG USER_ID
ARG GROUP_ID

# Don't attempt to set the user to the root user (uid=0) or group (gid=0)
RUN if [ ${USER_ID:-0} -eq 0 ] || [ ${GROUP_ID:-0} -eq 0 ]; then \
        groupadd golang \
        && useradd -g golang golang \
        ;\
    else \
        groupadd -g ${GROUP_ID} golang \
        && useradd -l -u ${USER_ID} -g golang golang \
        ;\
    fi \
    && install -d -m 0755 -o golang -g golang /home/golang \
    && mkdir -p /etc/sudoers.d  \
    && echo "golang ALL=(ALL) NOPASSWD: ALL" > /etc/sudoers.d/golang-all-nopasswd

RUN apt-get update \
    && apt-get -y install --no-install-recommends \
      build-essential \
      less \
      libasound2-dev \
      pkg-config \
      sudo \
      vim \
    && rm -rf /var/lib/apt/lists/*

USER golang

WORKDIR /workspace

FROM dev_image AS bin_builder
USER root

WORKDIR /build

COPY audio ./audio/
COPY config ./config/
COPY context ./context/
COPY event ./event/
COPY handler ./handler/
COPY led ./led/
COPY rfidsecuritysvc ./rfidsecuritysvc/
COPY ca.pem ./
COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY reader.go ./
COPY router.go ./

ENV GOARCH=arm
ENV GOOS=linux
ENV GOARM=6
ENV CGO_ENABLED=1
RUN go build -tags no_d2xx -o /build/magicband-reader

FROM debian:${DEBIAN_VERSION}-slim AS prod_image
USER root

RUN apt-get update && apt-get -y install --no-install-recommends libasound2

WORKDIR /magicband-reader

COPY --from=bin_builder /build/magicband-reader /magicband-reader/
COPY --from=bin_builder /build/ca.pem /magicband-reader/

# Create the database volume
RUN mkdir /sounds && chmod 750 /sounds
VOLUME /sounds

EXPOSE 9000

ENTRYPOINT ["/magicband-reader/magicband-reader"]

CMD ["--listen-address", "0.0.0.0", "--listen-port", "9000", "--sound-dir", "/sounds"]
