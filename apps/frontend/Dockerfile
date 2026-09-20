FROM oven/bun:1.4.2 AS base
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

# prod build image
FROM base AS build
COPY . .
RUN bun run build

# dev image
FROM base AS dev
EXPOSE 3000
ENTRYPOINT ["bun", "run", "dev"]

# prod image
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
