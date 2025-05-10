FROM golang:alpine AS builder

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# 移动到工作目录：/build
WORKDIR /build

# 将代码复制到容器中
COPY go.mod .
COPY go.sum .
# 下载依赖信息
RUN go mod download

# 将代码复制到容器中
COPY . .

# 将我们的代码编译成二进制可执行文件 bubble
RUN go build -o nil_app .

###################
# 接下来创建一个小镜像
###################
FROM debian:bullseye-slim

# 设置时区为北京时间（Debian 系统）
RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    dnsutils \
    curl \
    netcat-openbsd; \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime; \
    echo "Asia/Shanghai" > /etc/timezone; \
    dpkg-reconfigure -f noninteractive tzdata; \
    update-ca-certificates; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

# 从builder镜像中把静态文件拷贝到当前目录
COPY ./wait-for.sh /
# 关键修复：直接复制本地 config 目录到容器内的 /config 目录
COPY ./config /config

# 从builder镜像中拷贝二进制文件
COPY --from=builder /build/nil_app /

# 设置可执行权限
RUN chmod 755 /wait-for.sh

# 声明服务端口
EXPOSE 8888

# 需要运行的命令
ENTRYPOINT ["/nil_app"]