dev:
	cd dev && docker compose stop
	cd dev && docker compose rm -vf
	cd dev && docker compose up --build -d --remove-orphans

dev-logs:
	cd dev && docker compose logs -f

dev-env:
	cd dev && docker compose stop
	cd dev && docker compose up -d postgres

reset-dev-env:
	cd dev && docker compose stop
	cd dev && docker compose rm -vf


gen:
	go generate ./...

.PHONY: dev