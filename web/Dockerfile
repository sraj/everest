# Frontend Dockerfile for Everest
# Development image with Vite hot-reload

FROM node:22-alpine

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm install

# Copy source code
COPY . .

# Expose Vite dev server port
EXPOSE 5173

# Run Vite dev server with host flag for Docker
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
