FROM node:20-alpine AS builder
WORKDIR /usr/local/app
RUN apk add pnpm

COPY . .

RUN pnpm install && pnpm run build
RUN pnpm install --ignore-scripts vite-express

FROM gcr.io/distroless/nodejs20-debian12:debug-nonroot

COPY --from=builder --chown=nonroot:nonroot --chmod=777 /usr/local/app/dist /app/dist/
COPY --from=builder --chown=nonroot:nonroot --chmod=555 /usr/local/app/node_modules /app/node_modules
COPY --from=builder --chown=nonroot:nonroot --chmod=555 /usr/local/app/package*.json /usr/local/app/server.js /app/

WORKDIR /app
ENV NODE_ENV=production
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "/busybox/sh ./dist/vite-envs.sh && /nodejs/bin/node server.js"]
