FROM node:20-alpine AS builder
WORKDIR /usr/local/app
RUN apk add pnpm

COPY . .

RUN rm -rf node_modules && \
    pnpm install && pnpm run build

FROM node:20-alpine

# Create a user with a specific UID, GID, and home directory
ARG USERNAME=appuser
ARG UID=1001
ARG GID=1001
ARG HOME_DIR=/home/appuser

RUN addgroup -g ${GID} ${USERNAME}
RUN adduser -D -u ${UID} -G ${USERNAME} -h ${HOME_DIR} ${USERNAME}
RUN mkdir -p ${HOME_DIR} && chown -R ${UID}:${GID} ${HOME_DIR}
WORKDIR ${HOME_DIR}/app

COPY --from=builder /usr/local/app/dist ${HOME_DIR}/app/dist
COPY package*.json server.js ./

# Ensure appuser owns all files in /home/appuser/app
RUN chown -R ${UID}:${GID} ${HOME_DIR}/app

RUN apk add pnpm
USER ${USERNAME}

RUN pnpm install --ignore-scripts vite-express

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./dist/vite-envs.sh && npm run prod"]
