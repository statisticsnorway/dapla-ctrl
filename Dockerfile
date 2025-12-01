FROM node:24-alpine AS builder
WORKDIR /usr/local/app
RUN apk add pnpm

COPY . .

RUN pnpm install --ignore-scripts
RUN pnpm prepare
RUN pnpm run build
# Delete node_modules that contain dev deps and only install runtime deps
# in the version that is copied to the final output
RUN rm -rf node_modules/
RUN pnpm install -P --ignore-scripts

FROM gcr.io/distroless/nodejs20-debian12:debug-nonroot

COPY --from=builder --chown=nonroot:nonroot --chmod=777 /usr/local/app/dist /app/dist/
COPY --from=builder --chown=nonroot:nonroot --chmod=555 /usr/local/app/node_modules /app/node_modules
COPY --from=builder --chown=nonroot:nonroot --chmod=444 /usr/local/app/package*.json /app/
COPY --from=builder --chown=nonroot:nonroot --chmod=555 /usr/local/app/server.js /app/

WORKDIR /app
ENV NODE_ENV=production
ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "/busybox/sh ./dist/vite-envs.sh && /nodejs/bin/node server.js"]
