ARG NODE_IMAGE=node:12.16.1-alpine

FROM ${NODE_IMAGE} AS BUILDER
WORKDIR /app
COPY web/package-lock.json .
COPY web/package.json .
RUN npm ci
COPY web/. .
ARG REACT_APP_BACKEND_URL
ENV REACT_APP_BACKEND_URL=$REACT_APP_BACKEND_URL
RUN npm run build

FROM ${NODE_IMAGE}
WORKDIR /app
COPY --from=BUILDER /app/build .
EXPOSE 8080
# use http-server for simplicity, should be enough for production
# can consider using Nginx if we hit performance issues
ENTRYPOINT ["npx", "http-server", "."]