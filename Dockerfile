FROM debian:buster as lib_builder

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

FROM golang:1.16-buster

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
      less \
      libasound2-dev \
      sudo \
      vim \
    && rm -rf /var/lib/apt/lists/*

USER golang

WORKDIR /go/src/github.com/bcurnow/magicband-reader-golang
