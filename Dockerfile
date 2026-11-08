FROM golang:1.27.3-alpine AS go-source
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./

FROM go-source AS engine
RUN --mount=type=cache,target=/root/.cache/go-build GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -trimpath -o /game.wasm ./cmd/ebitengine
RUN cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" /wasm_exec.js

FROM go-source AS api
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api ./cmd/api

FROM node:24-alpine AS browser
WORKDIR /src/client
COPY client/package.json client/package-lock.json ./
RUN npm ci
COPY client/ ./
COPY --from=engine /game.wasm public/engine/game.wasm
COPY --from=engine /wasm_exec.js public/engine/wasm_exec.js
RUN NEXT_EXPORT=1 npm run build

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=api /api /app/api
COPY --from=browser /src/client/out /app/public
ENV PORT=8080 STATIC_DIR=/app/public
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
