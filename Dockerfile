FROM node:20-alpine AS builder
WORKDIR /usr/local/app

COPY . .
RUN npm ci && npm run build

FROM node:20-alpine

WORKDIR /usr/local/app

COPY --from=builder /usr/local/app/dist ./dist
COPY package*.json server.js ./

RUN npm install --save-exact express vite-express \
    && cp -r dist/* .

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./vite-envs.sh && npm run prod"]