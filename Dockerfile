FROM node:12.13.0-alpine AS builder
RUN apk add --no-cache git
RUN mkdir -p /opt/shield
WORKDIR /opt/shield
COPY package*.json ./
COPY yarn.lock ./
RUN yarn install
COPY . .
RUN yarn build

FROM node:12.13.0-alpine
WORKDIR /opt/shield
COPY --from=builder /opt/shield .
CMD ["yarn", "start"]