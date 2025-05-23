.PHONY: postgres air dev

postgres:
	@if service postgresql status | grep -q online; then \
		echo "PostgreSQL is already running."; \
	else \
		echo "Starting PostgreSQL..."; \
		sudo service postgresql start; \
	fi

air:
	air

dev: postgres
	air