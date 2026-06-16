---
description: "Task list for feature 001 — Autenticação, Isolamento e Gestão de Equipe"
---

# Tasks: Autenticação, Isolamento e Gestão de Equipe

**Input**: Design documents from `/specs/001-auth-isolamento-equipe/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth-api.md, quickstart.md

**Tests**: INCLUÍDOS — TDD é não-negociável nesta plataforma (Constituição, Princípios IV/V) e a spec/research exigem testes de unidade, integração (Mongo efêmero via testcontainers), **violação cross-tenant** (Princípio III) e **golden-path** da ativação (Princípio VIII).

**Organization**: Tarefas agrupadas por user story (todas P1) para implementação e teste independentes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Pode rodar em paralelo (arquivos diferentes, sem dependências pendentes)
- **[Story]**: User story a que pertence (US1, US2, US3)
- Caminhos de arquivo exatos incluídos em cada tarefa

## Path Conventions

Web app (Princípios I e XV): backend Go em `backend/`, frontend Next.js em `frontend/`,
conforme a Project Structure do [plan.md](plan.md).

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Inicialização dos projetos e infraestrutura local/CI

- [X] T001 Criar esqueleto do backend Go em `backend/` (`cmd/api/`, `internal/`, `pkg/`) e inicializar `backend/go.mod` (Go 1.23) com dependências: `chi`, `mongo-go-driver`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, `testify`, `testcontainers-go`
- [X] T002 Criar esqueleto do frontend Next.js 15 (App Router) em `frontend/` (`app/`, `hooks/`, `lib/api/`, `components/`) com `package.json`: shadcn/ui, Tailwind, `vitest`
- [X] T003 [P] Criar `backend/docker-compose.yml` subindo MongoDB para dev/CI (Princípio X)
- [X] T004 [P] Configurar lint/format: `golangci-lint`/`gofmt` no backend; `eslint`/`tsc`/`next lint` no frontend

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Infra central que TODAS as user stories dependem — inclui o núcleo de isolamento multi-tenant (Princípio III)

**⚠️ CRITICAL**: Nenhuma user story pode começar antes desta fase

- [X] T005 Conexão MongoDB + criação de índices em `backend/pkg/db/mongo.go`: `tenants{cnpj:1}` unique; `users{email:1}` unique global, `{tenantId:1,_id:1}`, `{tenantId:1,role:1}`; `invites{tokenHash:1}` unique, `{tenantId:1,email:1}`, `{expiresAt:1}` (TTL)
- [X] T006 [P] Modelos de domínio (structs) em `backend/internal/tenant/model.go`, `backend/internal/user/model.go`, `backend/internal/invite/model.go` conforme [data-model.md](data-model.md)
- [X] T007 [P] `pkg/password` bcrypt (custo 12) `Hash`/`Verify` + teste de unidade em `backend/pkg/password/password_test.go`
- [X] T008 [P] `pkg/token`: JWT de sessão HS256 (claims `sub`/`tid`/`role`/`exp`) sign/verify + geração de token de convite (256 bits) + hash SHA-256 + testes de unidade em `backend/pkg/token/token_test.go`
- [X] T009 [P] `pkg/email`: interface `EmailSender` + `FakeSender` (captura mensagens para asserção) + impl SMTP em `backend/pkg/email/`
- [X] T010 [P] Gestão de configuração (env: Mongo URI, JWT secret, SMTP, base URL do convite) em `backend/pkg/config/config.go`
- [X] T011 [P] Harness de integração com `testcontainers-go` (MongoDB efêmero) em `backend/internal/testutil/mongo.go`
- [X] T012 Middleware de autenticação em `backend/pkg/middleware/auth.go`: lê JWT do cookie httpOnly → injeta `tenantId`/`userId`/`role` no `context.Context`; **fail-closed** (sem token válido → 401)
- [X] T013 Helper `tenantScoped(ctx, filter)` em `backend/pkg/middleware/tenant.go` que adiciona `{tenantId}` do contexto a toda query (Princípio III) + teste de unidade que falha-fechado quando não há `tenantId` no contexto, em `backend/pkg/middleware/tenant_test.go`
- [X] T014 Repositórios tenant-scoped (interface + impl Mongo) usando `tenantScoped` em `backend/internal/user/repository.go` e `backend/internal/invite/repository.go`; `tenantId` jamais vem do payload
- [X] T015 Helper de erro `{code,message}` + servidor HTTP `chi` + stack de middlewares + DI em `backend/cmd/api/main.go` e `backend/pkg/httpx/response.go`

**Checkpoint**: Fundação pronta — núcleo de isolamento testado; user stories podem começar

---

## Phase 3: User Story 1 - Convidar um membro por e-mail (Priority: P1) 🎯 MVP

**Goal**: Um Admin convida um e-mail com um papel; o sistema cria um usuário `Pendente` na sua empresa, gera token de convite (48h) e dispara e-mail de ativação.

**Independent Test**: Autenticado como `Admin`, `POST /api/v1/invites` com e-mail novo → 201 + usuário `Pendente` + e-mail capturado pelo `FakeSender`; repetir o mesmo e-mail → 409; como `Editor`/`Viewer` → 403.

### Tests for User Story 1 (escrever primeiro — devem FALHAR)

- [X] T016 [P] [US1] Testes de integração de `POST /api/v1/invites` em `backend/internal/invite/handler_test.go`: 201 cria `Pendente` + e-mail capturado pelo `FakeSender`; reenvio para `Pendente` renova token sem duplicar usuário; 409 e-mail duplicado global; 403 para `Editor`/`Viewer`
- [X] T017 [P] [US1] Testes de unidade de `InviteService` (mock de UserRepo/InviteRepo/EmailSender) em `backend/internal/invite/service_test.go`

### Implementation for User Story 1

- [X] T018 [US1] `InviteService.CreateInvite` em `backend/internal/invite/service.go`: valida unicidade global de e-mail (409), cria usuário `Pendente`, gera+hash token (expiresAt +48h), persiste convite, dispara e-mail de ativação
- [X] T019 [US1] Guard RBAC `RequireRole(Admin)` em `backend/pkg/middleware/rbac.go` (Editor/Viewer → 403)
- [X] T020 [US1] Handler `POST /api/v1/invites` + mapeamento `InviteView` (sem token em texto puro) + rota em `backend/internal/invite/handler.go` e `backend/cmd/api/main.go`

**Checkpoint**: US1 funcional e testável isoladamente (Backend-First — sem UI de convite nesta feature)

---

## Phase 4: User Story 2 - Ativar conta e definir senha (Priority: P1)

**Goal**: O convidado acessa o link, define nome+senha (política mínima), passa a `Ativo` e é autenticado (cookie de sessão). Inclui login/logout/me.

**Independent Test**: Consumir link válido em `/accept-invite/{token}`, definir senha válida → conta `Ativo` + autenticado; token > 48h ou reusado → 410; token inválido → 404; senha fraca → 422.

### Tests for User Story 2 (escrever primeiro — devem FALHAR)

- [ ] T021 [P] [US2] Testes de integração de `POST /api/v1/invites/{token}/accept` em `backend/internal/auth/accept_test.go`: 200 ativa (`status=Ativo`, `usedAt`, cookie de sessão); 410 expirado/usado; 404 token inválido/adulterado; 422 senha fraca
- [ ] T022 [P] [US2] Testes de integração de auth em `backend/internal/auth/handler_test.go`: `POST /login` 200/401 (não-`Ativo` → 401); `GET /me` 200/401; `POST /logout` 204
- [ ] T023 [P] [US2] Testes de unidade de `AuthService` (ativação, política de senha, login) em `backend/internal/auth/service_test.go`
- [ ] T024 [P] [US2] Golden-path da tela de ativação (happy path + erro visível, API mockada) em `frontend/app/accept-invite/[token]/page.test.tsx`
- [ ] T025 [P] [US2] Testes de hooks `useAuth`/`useInvite` em `frontend/hooks/useAuth.test.ts` e `frontend/hooks/useInvite.test.ts`

### Implementation for User Story 2

- [ ] T026 [US2] `AuthService.AcceptInvite` em `backend/internal/auth/service.go`: valida convite `válido`, aplica política mínima de senha (≥8, letras+números), faz hash, ativa usuário (nome/`ativadoEm`), marca `usedAt`, emite sessão
- [ ] T027 [US2] `AuthService.Login`/`Logout`/`Me` em `backend/internal/auth/service.go` (verifica `Ativo` + bcrypt, emite JWT em cookie httpOnly SameSite=Strict)
- [ ] T028 [US2] Handlers `POST /accept`, `POST /auth/login`, `POST /auth/logout`, `GET /me` + rotas em `backend/internal/auth/handler.go` e `backend/cmd/api/main.go`
- [ ] T029 [P] [US2] Repositórios de API `authApi` e `inviteApi` em `frontend/lib/api/authApi.ts` e `frontend/lib/api/inviteApi.ts`
- [ ] T030 [P] [US2] Hooks (ViewModel) `useAuth` e `useInvite` em `frontend/hooks/useAuth.ts` e `frontend/hooks/useInvite.ts`
- [ ] T031 [US2] Tela de ativação `frontend/app/accept-invite/[token]/page.tsx` (form nome+senha, sucesso e erro visível)
- [ ] T032 [US2] Tela de login `frontend/app/(auth)/login/page.tsx`

**Checkpoint**: Fluxo de onboarding completo (US1+US2) funcional de ponta a ponta

---

## Phase 5: User Story 3 - Garantia de isolamento entre empresas (Priority: P1)

**Goal**: Comprovar e expor o isolamento: cada usuário só acessa dados da sua empresa; acesso cross-tenant retorna 404; `tenantId` do cliente é ignorado.

**Independent Test**: Criar dados nos Tenants A e B; como usuário de A, `GET /api/v1/users` retorna só de A; acessar `_id` de B → 404; enviar `tenantId` de B no corpo → ignorado.

### Tests for User Story 3 (escrever primeiro — devem FALHAR)

- [ ] T033 [P] [US3] Teste de **VIOLAÇÃO** cross-tenant em `backend/internal/user/isolation_test.go`: usuário do Tenant A acessa `_id`/recurso do Tenant B → 404 (sem vazar existência); `GET /api/v1/users` retorna apenas do tenant do contexto (SC-002)
- [ ] T034 [P] [US3] Teste de integração em `backend/pkg/middleware/tenant_isolation_test.go`: `tenantId` enviado no corpo/query é ignorado; empresa derivada exclusivamente do token (FR-010)
- [ ] T035 [P] [US3] Testes de integração de `GET /api/v1/users` em `backend/internal/user/handler_test.go`: 200 tenant-scoped (Admin); 403 para `Editor`/`Viewer`

### Implementation for User Story 3

- [ ] T036 [US3] Handler `GET /api/v1/users` (tenant-scoped, role `Admin`) + mapeamento `UserView` (sem `senhaHash`) + rota em `backend/internal/user/handler.go` e `backend/cmd/api/main.go`
- [ ] T037 [US3] Auditar todos os repositórios (`user`, `invite`) para garantir uso de `tenantScoped` e retorno 404 fail-closed em acesso cross-tenant

**Checkpoint**: Todas as user stories independentemente funcionais; isolamento comprovado por teste

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Validação final e endurecimento transversal

- [ ] T038 [P] Executar os cenários de validação ponta a ponta do [quickstart.md](quickstart.md)
- [ ] T039 [P] Auditar respostas de API e logs: nenhum `senhaHash`, token de convite ou senha exposto em qualquer cenário (SC-005)
- [ ] T040 [P] Verificar cobertura conforme os gates do agente `qa` (Services no backend; Hooks no frontend) — Princípios IV/V
- [ ] T041 [P] Documentação de dev (`backend/README.md`, `.env.example`) e atualização de `quickstart.md` se necessário

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: sem dependências
- **Foundational (Phase 2)**: depende do Setup — **BLOQUEIA** todas as user stories
- **User Stories (Phase 3–5)**: dependem da Fase 2; podem então seguir em paralelo ou na ordem P1
- **Polish (Phase 6)**: depende das user stories desejadas

### User Story Dependencies

- **US1 (P1)**: após Fundação. Testável isoladamente mintando token de sessão de Admin via `pkg/token` (não depende do endpoint de login do US2)
- **US2 (P1)**: após Fundação. Entrega login/logout/me + ativação; integra com o convite criado em US1, mas testável isoladamente semeando um convite válido
- **US3 (P1)**: após Fundação. Reusa o núcleo `tenantScoped`; entrega `GET /users` e os testes de violação

### Within Each User Story

- Testes escritos e **falhando** antes da implementação (TDD)
- Models → Services → Endpoints → Frontend
- Backend e contrato antes do frontend (Princípio XIV)

### Parallel Opportunities

- Setup: T003, T004 em paralelo
- Foundational: T006–T011 em paralelo; depois T012/T013 (isolamento), T014 (repos, depende de T013), T015 (wiring, depende de T012)
- US1: T016, T017 em paralelo (testes); US2: T021–T025 em paralelo; US3: T033–T035 em paralelo
- Frontend US2: T029, T030 em paralelo
- Com equipe: após a Fase 2, US1/US2/US3 podem ser tocadas por devs diferentes

---

## Parallel Example: User Story 2

```bash
# Testes do US2 juntos (devem falhar primeiro):
Task: "Integração accept em backend/internal/auth/accept_test.go"
Task: "Integração auth (login/logout/me) em backend/internal/auth/handler_test.go"
Task: "Unidade AuthService em backend/internal/auth/service_test.go"
Task: "Golden-path accept-invite em frontend/app/accept-invite/[token]/page.test.tsx"
Task: "Hooks useAuth/useInvite"

# Repositórios e hooks de frontend em paralelo:
Task: "authApi + inviteApi em frontend/lib/api/"
Task: "useAuth + useInvite em frontend/hooks/"
```

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1 (Setup) → 2. Phase 2 (Foundational, CRÍTICA — isolamento) → 3. Phase 3 (US1) → validar convite isoladamente → demo.

### Incremental Delivery

1. Setup + Foundational → fundação pronta (isolamento testado)
2. US1 → convite (MVP backend)
3. US2 → ativação + login (onboarding completo, com frontend)
4. US3 → `GET /users` + comprovação de isolamento
5. Polish → quickstart + auditoria de segredos + cobertura

---

## Notes

- [P] = arquivos diferentes, sem dependências pendentes
- TDD é não-negociável: verificar que cada teste falha antes de implementar
- Commit após cada tarefa ou grupo lógico
- Parar em qualquer checkpoint para validar a story isoladamente
- Não-negociáveis sob teste explícito: isolamento cross-tenant (III, T033/T034) e segurança de credenciais (VI, T039)
