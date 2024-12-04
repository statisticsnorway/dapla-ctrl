FROM node:20-alpine AS builder
WORKDIR /usr/local/app

COPY . .
RUN npm ci && npm run build

FROM node:20-alpine

WORKDIR /usr/local/app

COPY --from=builder /usr/local/app/dist ./dist
COPY package*.json server.js ./

# Let user write to dist directory, this is necessary for vite-envs script to work
RUN chown -R 0:0 ./dist
RUN npm install --ignore-scripts --save-exact express vite-express

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./dist/vite-envs.sh && npm run prod"]
