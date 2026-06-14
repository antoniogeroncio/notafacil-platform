# Quickstart — 001 Autenticação, Isolamento e Gestão de Equipe

Guia de execução e validação ponta a ponta. Detalhes de modelo/contrato:
ver [data-model.md](data-model.md) e [contracts/auth-api.md](contracts/auth-api.md).

## Pré-requisitos

- Go 1.23+, Node 20+, Docker (para MongoDB via Compose).

## Subir o ambiente (Princípio X)

```bash
# MongoDB efêmero
cd backend && docker compose up -d mongo

# Backend (API em :8080)
cd backend && go run ./cmd/api

# Frontend (Next.js em :3000)
cd frontend && npm install && npm run dev
```

## Rodar os testes

```bash
# Backend: unidade + integração (sobe Mongo via testcontainers)
cd backend && go test ./...

# Frontend: hooks + golden-path
cd frontend && npx vitest run
```

## Cenários de validação (mapeados à spec)

1. **Convite (US1 / SC-001)** — autenticado como `Admin`, `POST /api/v1/invites`
   com um e-mail novo e role `Editor` → 201, usuário `Pendente` criado, e-mail de
   ativação capturado pelo `FakeSender`. Repetir o mesmo e-mail → **409**.
2. **Permissão (FR-008)** — como `Editor`, `POST /api/v1/invites` → **403**.
3. **Ativação (US2 / SC-003)** — abrir o link `/accept-invite/{token}`, definir
   nome + senha válida → conta `Ativo` e autenticado. Token > 48h ou reusado →
   **410**.
4. **Isolamento (US3 / SC-002)** — criar dados nos Tenants A e B; como usuário de
   A, `GET /api/v1/users` retorna só de A; acessar `_id` de B → **404**. Enviar
   `tenantId` de B no corpo → ignorado.
5. **Segredos (SC-005)** — inspecionar respostas e logs dos fluxos acima: nenhum
   `senhaHash`, token de convite ou senha aparece.

## Definition of Done (gates)

- `go test ./...` verde, incluindo o **teste de violação** cross-tenant.
- Golden-path da tela de ativação verde (sessão limpa, happy path + erro visível).
- `next lint` / `tsc` sem erros. Cobertura de Services/Hooks conforme gates do `qa`.
