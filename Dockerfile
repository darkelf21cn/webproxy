FROM ubuntu:20.04

ENV SS_SERVER_HOST=
ENV SS_SERVER_PORT=
ENV SS_SERVER_PASSWORD=
ENV SS_SERVER_ENCRYPT_METHOD=aes-128-ctr
ENV SS_SERVER_TIMEOUT=300
ENV CHINA_IP_LIST_UPDATE_INTERVAL_HOUR=24

RUN export DEBIAN_FRONTEND=noninteractive \
    && apt update \
    && apt install squid shadowsocks-libev privoxy wget -y
COPY ./privoxy.conf /etc/privoxy/config
COPY ./gen_squid_conf.sh /usr/local/sbin/gen_squid_conf.sh
COPY ./start.sh /usr/local/sbin/start.sh
RUN chmod +x /usr/local/sbin/start.sh
RUN chmod +x /usr/local/sbin/gen_squid_conf.sh
RUN /usr/local/sbin/gen_squid_conf.sh

EXPOSE 3128
ENTRYPOINT [ "/bin/bash", "-c", "/usr/local/sbin/start.sh" ]
