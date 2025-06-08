FROM oven/bun:1.0.13-slim

WORKDIR /app

# 只复制 .json（不复制不存在的 bun.lockb）
COPY package.json ./

RUN bun install

COPY . .

EXPOSE 3000

CMD ["bun", "index.js"]
