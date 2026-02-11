# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
ENV VITE_API_URL=https://demo.pews.app
RUN npm run build

# Runtime stage - Node adapter
FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package*.json ./
RUN npm install --omit=dev
ENV PORT=3000
ENV ORIGIN=https://demo.pews.app
EXPOSE 3000
CMD ["node", "build"]
