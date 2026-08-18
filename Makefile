.PHONY: frontend backend build test clean

frontend:
	cd frontend && npm ci && npm run build
	rm -rf backend/web/dist
	mkdir -p backend/web/dist
	cp -R frontend/dist/. backend/web/dist/

backend:
	cd backend && go build -o ../bin/wireguard-ui .

build: frontend backend

test:
	cd backend && go test ./...

clean:
	rm -rf bin frontend/dist backend/web/dist
	mkdir -p backend/web/dist
	@echo "restore placeholder with: git checkout -- backend/web/dist/index.html"
