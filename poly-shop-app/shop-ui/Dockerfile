# Build Stage
FROM node:20-alpine as build
WORKDIR /app

# Build-time arg for OTLP endpoint (default points to local Jaeger OTLP HTTP)
ARG VITE_OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318/v1/traces
ENV VITE_OTEL_EXPORTER_OTLP_ENDPOINT=${VITE_OTEL_EXPORTER_OTLP_ENDPOINT}

COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

# Production Stage
FROM nginxinc/nginx-unprivileged:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 8080
# Using unprivileged nginx, port 8080 is default
CMD ["nginx", "-g", "daemon off;"]
