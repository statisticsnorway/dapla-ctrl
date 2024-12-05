FROM node:20-alpine AS builder
WORKDIR /usr/local/app

COPY . .
RUN npm ci && npm run build

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

USER ${USERNAME}

RUN npm install --ignore-scripts --save-exact express vite-express

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["sh", "-c", "./dist/vite-envs.sh && npm run prod"]
