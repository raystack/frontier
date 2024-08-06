FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

COPY package.json ./
RUN pnpm install
COPY . .
RUN npm run build
EXPOSE 3000
ENV PORT 3000
CMD ["node", ".next/standalone/server.js"] 