# build environment
FROM node:20-alpine AS build
WORKDIR /app
COPY . .
RUN npm ci && npm run build

# production environment
FROM node:20-alpine
WORKDIR /app
COPY --from=build /app/dist .

COPY package*.json .
COPY server.js .

RUN npm i --save-exact express vite-express

ENV PORT 8080
EXPOSE 8080

ENTRYPOINT sh -c "./vite-envs.sh && npm run prod"