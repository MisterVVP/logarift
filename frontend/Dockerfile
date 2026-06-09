FROM node:22-bookworm-slim
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --ignore-scripts --audit=false --fund=false
COPY frontend/index.html frontend/vite.config.js ./
COPY frontend/src ./src
EXPOSE 5173
CMD ["npm", "run", "dev"]
