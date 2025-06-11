FROM oven/bun:1.0.13-slim

# 安装 curl 和 wget
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*  # 清理缓存，减小镜像体积

WORKDIR /app

# 只复制 .json（不复制不存在的 bun.lockb）
COPY package.json bun.lockb ./

RUN bun install

COPY . .

EXPOSE 3000

CMD ["bun", "index.js"]
