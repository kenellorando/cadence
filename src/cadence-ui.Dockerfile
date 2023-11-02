# syntax=docker/dockerfile:1
FROM node:20-alpine3.17 as ui-builder
WORKDIR /ui-builder
COPY ./ui/src ./src
COPY ./ui/static ./static
COPY ./ui/package*.json ./
COPY ./ui/*.config.js ./
RUN npm ci
## Produces a production "build" directory, see svelte.config.js
RUN npm run build
RUN npm prune --production


CMD [ "node", "build"]

# ARG ARCH=
# FROM node:20-alpine3.17 
# LABEL maintainer="Ken Ellorando (kenellorando.com)"
# LABEL source="github.com/kenellorando/cadence"

# WORKDIR /ui
# COPY --from=ui-builder /ui-builder/package*.json ./
# COPY --from=ui-builder /ui-builder/build ./build/
# RUN npm ci --omit dev
# ENV NODE_ENV production
# CMD [ "node", "build"]
