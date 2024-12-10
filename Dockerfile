FROM node:20-alpine AS builder
WORKDIR /usr/local/app
RUN apk add pnpm

COPY . .

RUN rm -rf node_modules

RUN pnpm i && pnpm run build

FROM node:20-alpine

WORKDIR /usr/local/app
RUN apk add pnpm

COPY --from=builder /usr/local/app/dist ./dist
COPY package*.json server.js ./

RUN pnpm install --ignore-scripts vite-express

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./dist/vite-envs.sh && npm run prod"]
