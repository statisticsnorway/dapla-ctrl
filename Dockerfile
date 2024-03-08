FROM node:20-alpine AS builder
WORKDIR /usr/local/app

COPY . .

RUN npm ci && npm run build

FROM node:20-alpine

COPY --from=builder /usr/local/app/dist /usr/local/app/dist
WORKDIR /usr/local/app

COPY --from=builder /app/dist .
COPY package*.json .
COPY server.js .

RUN npm i --save-exact express vite-express

ENV PORT 8080
EXPOSE 8080

ENTRYPOINT sh -c "./vite-envs.sh && npm run prod"