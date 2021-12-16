# A region based intellegent all in one web-proxy solution

# Quick start

```bash
docker pull darkelf21cn/webproxy:latest
docker run -d --name webproxy -e SS_SERVER_HOST=${SERVER} -e SS_SERVER_PORT=${PORT} -e SS_SERVER_PASSWORD=${PASSWORD} -e SS_SERVER_ENCRYPT_METHOD="aes-128-ctr" -p 3128:3128 darkelf21cn/webproxy:latest
```

# How it works

```
                                                                                                       +-------------------+
                                                                                                       |                   |
                                              +-------------------------------------------------------->  Internal Sites   |
                                              |                                                        |                   |
                                              |                                                        +-------------------+
                                              |
                                              |
+---------------+        +---------------+    |    +---------------+        +-----------------+        +-------------------+
|               |        |               |    |    |               |        |                 |        |                   |
|    Client     +------->|    Squid      +----+---->    Privoxy    +-------->   Shadowsocks   +-------->  External Sites   |
|               |        |               |         |               |        |                 |        |                   |
+---------------+        +---------------+         +---------------+        +-----------------+        +-------------------+
                                 ^
                                 |
                                 |
                         +-------+-------+
                         |               |
                         | China IP list |
                         |    (APNIC)    |
                         +---------------+
```
