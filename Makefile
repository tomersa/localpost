.PHONY: build test install

build:
	mkdir -p build/ && go build -o build/localpost

test:
	go test ./...

install: build
	chmod +x build/localpost
	sudo mv build/localpost /usr/local/bin/localpost
	@echo "Checking for shell binaries and adding autocompletion..."
	@if [ -f /usr/bin/zsh ]; then \
		if ! grep -q 'localpost completion --shell zsh' $$HOME/.zshrc 2>/dev/null; then \
			echo '# Completions for localpost' >> $$HOME/.zshrc; \
			echo 'source <(localpost completion --shell zsh)' >> $$HOME/.zshrc; \
			echo "Autocompletion added to ~/.zshrc"; \
		else \
			echo "Autocompletion already present in ~/.zshrc"; \
		fi; \
	elif [ -f /usr/bin/bash ]; then \
		if ! grep -q 'localpost completion --shell bash' $$HOME/.bashrc 2>/dev/null; then \
			echo '# Completions for localpost' >> $$HOME/.bashrc; \
			echo 'source <(localpost completion --shell bash)' >> $$HOME/.bashrc; \
			echo "Autocompletion added to ~/.bashrc"; \
		else \
			echo "Autocompletion already present in ~/.bashrc"; \
		fi; \
	else \
		echo "Neither /usr/bin/zsh nor /usr/bin/bash found."; \
		echo "Please add autocompletion manually:"; \
		echo "source <(localpost completion --shell bash|zsh)"; \
	fi

