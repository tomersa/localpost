.PHONY: build test install

BUILD_DIR=build
LPOST_PATH=$(BUILD_DIR)/lpost

build:
	mkdir -p $(BUILD_DIR)/ && go build -o $(LPOST_PATH)

test:
	go test ./...

install: build
	chmod +x $(LPOST_PATH)
	sudo mv $(LPOST_PATH) /usr/local/bin/lpost
	@echo "Checking for shell binaries and adding autocompletion..."
	@if [ -f /usr/bin/zsh ]; then \
		if ! grep -q '$(LPOST_PATH) completion --shell zsh' $$HOME/.zshrc 2>/dev/null; then \
			echo '# Completions for lpost' >> $$HOME/.zshrc; \
			echo 'source <($(LPOST_PATH) completion --shell zsh)' >> $$HOME/.zshrc; \
			echo "Autocompletion added to ~/.zshrc"; \
		else \
			echo "Autocompletion already present in ~/.zshrc"; \
		fi; \
	elif [ -f /usr/bin/bash ]; then \
		if ! grep -q 'lpost completion --shell bash' $$HOME/.bashrc 2>/dev/null; then \
			echo '# Completions for lpost' >> $$HOME/.bashrc; \
			echo 'source <(lpost completion --shell bash)' >> $$HOME/.bashrc; \
			echo "Autocompletion added to ~/.bashrc"; \
		else \
			echo "Autocompletion already present in ~/.bashrc"; \
		fi; \
	else \
		echo "Neither /usr/bin/zsh nor /usr/bin/bash found."; \
		echo "Please add autocompletion manually:"; \
		echo "source <(lpost completion --shell bash|zsh)"; \
	fi

