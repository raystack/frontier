FROM node:20-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

COPY package.json .
RUN npm install
COPY . .
RUN npm run build
EXPOSE 3000
ENV PORT 3000

COPY ./.next/standalone build/.
COPY ./.next/static build/.next/static/.
CMD ["node", "build/server.js"]
