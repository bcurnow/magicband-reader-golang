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

FROM golang:1.16-buster as dev_image

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

FROM dev_image as bin_builder
USER root

WORKDIR /build

COPY . /build

ENV GOARCH=arm
ENV GOOS=linux
ENV GOARM=6
ENV CGO_ENABLED=1
RUN go build -tags no_d2xx -o /build/magicband-reader

FROM debian:buster-slim as prod_image
USER root

RUN apt-get update && apt-get -y install --no-install-recommends libasound2

WORKDIR /magicband-reader

COPY --from=bin_builder /build/magicband-reader /magicband-reader/
COPY --from=bin_builder /build/ca.pem /magicband-reader/

EXPOSE 9000

CMD ["/magicband-reader/magicband-reader", "--listen-address", "0.0.0.0", "--listen-port", "9000"]
