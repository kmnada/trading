# Dockerfile for frontend

FROM node:16-alpine AS build-stage

WORKDIR /app

# Copy source code and install dependencies
COPY package*.json ./
RUN npm install
COPY . .

# Build
RUN npm run build

# Use nginx to serve the static files
FROM nginx:alpine AS production-stage
COPY --from=build-stage /app/dist/ /usr/share/nginx/html

# Expose port 80
EXPOSE 80
ENTRYPOINT ["nginx", "-g", "daemon off;"]