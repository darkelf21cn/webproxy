FROM golang:1.17 AS gfwpass
WORKDIR /gfwpass
COPY ./gfwpass/go.mod go.mod
COPY ./gfwpass/go.sum go.sum
ENV GOPROXY="https://goproxy.cn"
RUN go mod download
COPY ./gfwpass .
RUN go build .

FROM ubuntu:22.04
RUN export DEBIAN_FRONTEND=noninteractive \
    && apt update \
    && apt install squid shadowsocks-libev privoxy wget -y
COPY ./privoxy.conf /etc/privoxy/config
COPY ./gen_squid_conf.sh /usr/local/sbin/gen_squid_conf.sh
COPY ./start.sh /usr/local/sbin/start.sh
RUN chmod +x /usr/local/sbin/start.sh
RUN chmod +x /usr/local/sbin/gen_squid_conf.sh
RUN /usr/local/sbin/gen_squid_conf.sh
COPY --from=gfwpass /gfwpass/gfwpass /usr/local/bin/gfwpass
RUN chmod +x /usr/local/bin/gfwpass
EXPOSE 3128
ENV GFWPASS_LOGLEVEL=
ENV GFWPASS_PORT=
ENV GFWPASS_SUBS_URL=
ENV GFWPASS_SUBS_INTERVAL_HOUR=
ENV GFWPASS_HC_URLS=
ENV GFWPASS_HC_INTERVAL_SEC=
ENV GFWPASS_HC_TIMEOUT_SEC=
ENV GFWPASS_HC_APPEMPTS=
ENV CHINA_IP_LIST_UPDATE_INTERVAL_HOUR=24
ENTRYPOINT [ "/bin/bash", "-c", "/usr/local/sbin/start.sh" ]
