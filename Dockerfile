# syntax=docker/dockerfile:1

# ---- Étape 1 : build du frontend Vue/Vite ----
FROM node:22-alpine AS frontend
WORKDIR /src/app/frontend
COPY app/frontend/package.json app/frontend/package-lock.json ./
RUN npm ci
COPY app/frontend/ ./
RUN npm run build

# ---- Étape 2 : compilation du backend Go avec le frontend intégré ----
FROM golang:1.26-alpine AS backend
WORKDIR /src
COPY app/backend/ ./
COPY --from=frontend /src/app/backend/static/ ./static/
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/ojozone-api ./cmd/server

# ---- Étape 3 : image d'exécution minimale, non privilégiée ----
FROM alpine:3.20
RUN adduser -D -u 10001 ojozone
COPY --from=backend /out/ojozone-api /usr/local/bin/ojozone-api
USER ojozone
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["ojozone-api"]