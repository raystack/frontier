FROM node:12.13.0-alpine AS builder
WORKDIR /app
COPY package*.json ./
COPY yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build

FROM node:12.13.0-alpine AS server
ENV NODE_ENV production
ENV NEW_RELIC_HOME ./build
WORKDIR /app
COPY package*.json ./
COPY yarn.lock ./
RUN yarn install --frozen-lockfile --production
COPY --chown=node:node --from=builder ./app/build ./build
USER node
CMD ["yarn", "start"]
