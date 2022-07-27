# A region based intellegent all in one web-proxy solution

# Quick start

```bash
docker pull darkelf21cn/webproxy:latest
docker run -d --name webproxy -e GFWPASS_SUBS_URL=${SUBSCRIPTION_URL} -p 3128:3128 darkelf21cn/webproxy:latest
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
