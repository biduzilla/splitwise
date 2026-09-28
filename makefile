# ─────────────────────────────────────────────────────────────────────────────
# Rateio — Makefile raiz
# ─────────────────────────────────────────────────────────────────────────────

BACKEND    := backend
COMPOSE    := docker compose -f $(BACKEND)/docker-compose.yml --env-file $(BACKEND)/.env

# Serviços
SERVICES   := ms_auth ms_group ms_expense ms_balance

# ── Default ──────────────────────────────────────────────────────────────────

.DEFAULT_GOAL := help

.PHONY: help
help: ## Mostra os comandos disponíveis
	@echo ""
	@echo "  Splitwise — comandos disponíveis"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo ""

# ── Infraestrutura (Docker) ──────────────────────────────────────────────────

.PHONY: up
up: ## Sobe Postgres, Redis, Kafka, Jaeger, Prometheus, Grafana
	$(COMPOSE) up -d
	@echo "Infra rodando. Verifique com: make ps"

.PHONY: down
down: ## Derruba a infraestrutura
	$(COMPOSE) down

.PHONY: restart
restart: down up ## Reinicia a infraestrutura

.PHONY: reset
reset: ## Derruba tudo E apaga volumes (destrutivo!)
	$(COMPOSE) down -v
	@echo "⚠ Volumes apagados. Postgres, Redis e Kafka zerados."

.PHONY: ps
ps: ## Lista containers e status
	$(COMPOSE) ps

.PHONY: logs
logs: ## Segue logs de um serviço (SVC=ms_auth)
	$(COMPOSE) logs -f $(SVC)

# ── Backend (Go) ─────────────────────────────────────────────────────────────

.PHONY: auth
auth: ## Roda ms_auth localmente (:4001)
	cd $(BACKEND)/ms_auth && go run ./cmd/api

.PHONY: group
group: ## Roda ms_group localmente (:4002)
	cd $(BACKEND)/ms_group && go run ./cmd/api

.PHONY: expense
expense: ## Roda ms_expense localmente (:4003)
	cd $(BACKEND)/ms_expense && go run ./cmd/api

.PHONY: balance
balance: ## Roda ms_balance localmente (:4004)
	cd $(BACKEND)/ms_balance && go run ./cmd/api

.PHONY: build
build: ## Compila todos os módulos do workspace
	cd $(BACKEND) && go build all

.PHONY: test
test: ## Roda todos os testes
	cd $(BACKEND) && go test all

.PHONY: tidy
tidy: ## Roda go mod tidy em todos os módulos
	@for svc in $(SERVICES); do \
		echo "→ $$svc"; \
		cd $(BACKEND)/$$svc && go mod tidy && cd ../..; \
	done

.PHONY: fmt
fmt: ## Formata todo o código Go
	cd $(BACKEND) && gofmt -w -s .

# ── Debug ────────────────────────────────────────────────────────────────────

.PHONY: redis
redis: ## Abre redis-cli no container
	$(COMPOSE) exec redis redis-cli -a $$(grep REDIS_PASSWORD $(BACKEND)/.env | cut -d= -f2)

.PHONY: redis-flush
redis-flush: ## Limpa o cache Redis (dev)
	@docker exec -it rateio_redis redis-cli -a $$(grep REDIS_PASSWORD $(BACKEND)/.env | cut -d= -f2) FLUSHALL

.PHONY: psql
psql: ## Abre psql no container Postgres
	$(COMPOSE) exec postgres psql -U rateio_user -d rateio_db

.PHONY: jaeger
jaeger: ## Abre o Jaeger UI no browser
	@explorer.exe "http://localhost:16686"

.PHONY: grafana
grafana: ## Abre o Grafana no browser
	@explorer.exe "http://localhost:3001"

.PHONY: prometheus
prometheus: ## Abre o Prometheus no browser
	@explorer.exe "http://localhost:9090"

.PHONY: kafka-ui
kafka-ui: ## Abre o Kafka UI no browser
	@explorer.exe "http://localhost:8070"