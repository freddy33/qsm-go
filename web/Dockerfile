FROM node:12 AS BUILDER
WORKDIR /app
COPY package.json .
COPY package-lock.json .
RUN npm ci
COPY . .
RUN npm run build

FROM node:12-alpine
RUN npm -g install http-server
WORKDIR /app
COPY --from=BUILDER /app/build .
EXPOSE 3000
ENTRYPOINT [ "http-server", ".", "-p", "3000" ]
