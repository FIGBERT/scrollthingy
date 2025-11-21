dev-client:
    @cd client && npm install && gleam run -m lustre/dev start

dev-server:
    @cd server && go run .

_build-client:
    @cd client && npm install && gleam run -m lustre/dev build

build-prod: _build-client
    @cp client/priv/static/client.mjs server/internal/frontend/client.mjs
    @cd server && go build .

compile:
    @docker build --output=. .

clean:
    rm -rf client/build client/priv client/node_modules
    rm -f server/server
