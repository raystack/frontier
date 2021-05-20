FROM node:12.13.0-alpine AS builder
RUN apk add --no-cache git
RUN mkdir -p /opt/library
WORKDIR /opt/library
COPY package*.json ./
RUN npm install
COPY . .

FROM node:12.13.0-alpine
WORKDIR /opt/library
COPY --from=builder /opt/library .
RUN apk update && apk add bash
RUN chmod +x ./wait-for-it.sh
CMD ["./wait-for-it.sh", "shield:5000/ping", "--", "npm", "run", "start"]