FROM node:20-alpine AS builder
WORKDIR /usr/local/app

COPY . .

# Required for building Vite frontend (env var must be available at build)
ARG VITE_DAPLA_TEAM_API_URL
ENV VITE_DAPLA_TEAM_API_URL=${VITE_DAPLA_TEAM_API_URL}

RUN npm ci && npm run build

FROM node:20-alpine

COPY --from=builder /usr/local/app/dist /usr/local/app/dist
WORKDIR /usr/local/app

COPY package*.json .
COPY server.js .

RUN npm i --save-exact express vite-express

ENV PORT 8080
EXPOSE 8080

CMD ["npm", "run", "prod"]