FROM oven/bun:1.4.2 AS base
WORKDIR /app

FROM base AS build
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM base AS dev
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

EXPOSE 3000
ENTRYPOINT ["bun", "run", "dev"]

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
