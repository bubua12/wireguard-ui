FROM node:20-alpine AS frontend
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci || npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend
WORKDIR /src
COPY backend/ ./backend/
COPY --from=frontend /src/frontend/dist ./backend/web/dist
WORKDIR /src/backend
RUN CGO_ENABLED=0 go build -o /wireguard-ui .

FROM alpine:3.21
RUN apk add --no-cache wireguard-tools iptables iproute2
COPY --from=backend /wireguard-ui /usr/bin/wireguard-ui
ENV WG_LISTEN=0.0.0.0:8081 \
    WG_DB_PATH=/data/wireguard.db \
    WG_CONFIG_DIR=/etc/wireguard
EXPOSE 8081
VOLUME ["/data", "/etc/wireguard"]
ENTRYPOINT ["wireguard-ui"]
