# --- BUILD STAGE ---
FROM node:22-alpine AS builder

WORKDIR /app

# Copy package descriptors
COPY package.json package-lock.json ./

# Install dependencies (cached layer)
RUN npm ci

# Copy source code
COPY . .

# Build for production
RUN npm run build -- --configuration=production

# --- RUN STAGE ---
FROM nginx:1.27-alpine

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy build artifacts to Nginx html directory
COPY --from=builder /app/dist/frontend/browser /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
