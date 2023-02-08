.PHONY: all

all: tailwind
	go build ./cmd/tevents

run:
	go run ./cmd/tevents

tailwind:
	cd assets && npx tailwindcss -i ./styles.css -o ./output.css

tailwind-watch:
	cd assets && npx tailwindcss -i ./styles.css -o ./output.css --watch
