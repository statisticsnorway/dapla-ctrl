FROM node:22-alpine AS node-with-deps
WORKDIR /usr/app
RUN apk add pnpm

COPY package.json pnpm*.yaml svelte.config.js .npmrc ./

RUN pnpm install

COPY . ./

ENV VITE_GRAPHQL_ENDPOINT=http://dapla-api/graphql

RUN pnpm run build

FROM node:22-alpine
WORKDIR /usr/app

RUN apk add pnpm
ENV NODE_ENV=production

COPY --from=node-with-deps /usr/app/package*.json /usr/app/.npmrc ./
RUN pnpm install -P --ignore-scripts

COPY --from=node-with-deps /usr/app/build ./

CMD ["node", "./index.js"]
