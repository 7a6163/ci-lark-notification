local-build:
	docker build -t mobydeck/ci-lark-notification .
	docker image prune -f
