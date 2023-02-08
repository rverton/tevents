.PHONY: all

all: tailwind
	go build ./cmd/tevents

run:
	go run ./cmd/tevents

watch:
	nodemon --watch '*' -e html,go  --exec go run ./cmd/tevents --signal SIGTERM

tailwind:
	cd assets && npx tailwindcss -i ./styles.css -o ./output.css

tailwind-watch:
	cd assets && npx tailwindcss -i ./styles.css -o ./output.css --watch
