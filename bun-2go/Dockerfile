FROM oven/bun:1.0.13-slim

WORKDIR /app

COPY . .

RUN bun install --production

COPY . .

EXPOSE 3000

RUN chmod +x index.js

CMD ["bun", "index.js"]
