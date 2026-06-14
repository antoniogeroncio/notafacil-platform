# Implementation Plan: Autenticação, Isolamento e Gestão de Equipe

**Branch**: `001-auth-isolamento-equipe` | **Date**: 2026-06-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-auth-isolamento-equipe/spec.md`

## Summary

Entregar o núcleo de identidade da plataforma: autenticação por e-mail/senha,
contexto de sessão com `tenantId` + papel, convite de membros por e-mail com
ativação em 48h, RBAC (`Admin`/`Editor`/`Viewer`) e — o mais crítico — o
**isolamento multi-tenant central e por padrão** (Princípio III). Backend em Go
(Handler → Service → Repository) sobre MongoDB single-database com filtro por
tenant injetado via `context.Context`; frontend Next.js (login + ativação de
convite). Abordagem dirigida por testes (Princípios IV/V), incluindo testes de
**violação** (acesso cross-tenant negado) e **golden-path** da ativação.

## Technical Context

**Language/Version**: Go 1.23 (backend); TypeScript 5 / Next.js 15 App Router (frontend)

**Primary Dependencies**: backend — `chi` (router HTTP), `mongo-go-driver`, `golang-jwt/jwt v5`, `x/crypto/bcrypt`, `testify`; frontend — Next.js, shadcn/ui, Tailwind, `vitest`

**Storage**: MongoDB (multi-tenant single-database; índices compostos com `tenantId`)

**Testing**: backend — `go test` (unidade com `testify/mock`; integração com `testcontainers-go` + MongoDB efêmero); frontend — `vitest` (hooks) + golden-path no browser

**Target Platform**: Linux server em containers (Docker Compose para dev/CI)

**Project Type**: Web application (backend Go + frontend Next.js)

**Performance Goals**: convite→e-mail < 1 min (SC-001); login/p95 < 300ms; verificação de isolamento sem custo perceptível (filtro indexado)

**Constraints**: isolamento cross-tenant 100% (SC-002); zero exposição de segredos (SC-005); `tenantId` derivado só do token (FR-010)

**Scale/Scope**: SaaS B2B inicial (milhares de empresas, dezenas de usuários por empresa); esta feature: ~6 endpoints, 3 coleções, 2 telas

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Princípio | Como o plano atende | Status |
|-----------|---------------------|--------|
| I — Camadas | Handler → Service → Repository; integrações (e-mail) atrás de interface injetada | ✅ |
| II — Clean Code | Código em inglês; erros com contexto; sem segredos em log | ✅ |
| III — Isolamento (NN) | Middleware extrai `tenantId` do JWT → `context`; repositórios filtram por tenant por padrão; `tenantId` do cliente ignorado | ✅ |
| IV — Estratégia de Testes | Unidade em Services; integração (Mongo efêmero) em Handlers/Repos; testes de violação cross-tenant | ✅ |
| V — TDD (NN) | Testes escritos antes; Red→Green→Refactor | ✅ |
| VI — Segredos (NN) | Senha com bcrypt; tokens de convite com hash; nada em texto puro | ✅ |
| VIII — Golden-Path | Tela de ativação de convite (form por URL) coberta por golden-path | ✅ |
| X — IaC | `docker-compose` sobe MongoDB; CI usa a mesma definição | ✅ |
| XI/XII — API | REST `/api/v1`, erros `{code,message}`, 401/403/409 semânticos | ✅ |
| XIV — Backend-First (NN) | Backend e contrato antes do frontend; contratos em `contracts/` | ✅ |
| XV — Frontend MVVM | `Page → Hook → Repository`; cookie httpOnly de sessão | ✅ |

**Resultado**: PASS — nenhuma violação. Complexity Tracking vazio.

## Project Structure

### Documentation (this feature)

```text
specs/001-auth-isolamento-equipe/
├── plan.md              # Este arquivo
├── research.md          # Fase 0 — decisões técnicas
├── data-model.md        # Fase 1 — coleções, índices, transições
├── quickstart.md        # Fase 1 — guia de execução/validação
├── contracts/
│   └── auth-api.md      # Fase 1 — contrato REST (/api/v1)
└── tasks.md             # Fase 2 — (/speckit-tasks)
```

### Source Code (repository root)

```text
backend/
├── cmd/api/main.go                      # wiring/DI + HTTP server
├── internal/
│   ├── tenant/                          # Empresa: handler/service/repository
│   ├── user/                            # Usuário: handler/service/repository
│   ├── invite/                          # Convite: handler/service/repository
│   └── auth/                            # login, sessão, ativação
├── pkg/
│   ├── middleware/                      # auth + tenant context (Princípio III)
│   ├── token/                           # JWT de sessão + token de convite
│   ├── password/                        # hash/verify (bcrypt)
│   └── email/                           # EmailSender (interface) + impl
├── go.mod
└── docker-compose.yml                   # MongoDB para dev/CI

frontend/
├── app/
│   ├── (auth)/login/page.tsx            # login
│   └── accept-invite/[token]/page.tsx   # ativação de convite (golden-path)
├── hooks/                               # useAuth, useInvite (ViewModel)
├── lib/api/                             # authApi, inviteApi (Repository)
└── components/
```

**Structure Decision**: Web application (Princípios I e XV). Backend Go em
`backend/` com camadas por domínio; frontend Next.js em `frontend/`. O
isolamento multi-tenant vive em `pkg/middleware` + repositórios tenant-scoped.

## Complexity Tracking

> Sem violações da constituição — nada a justificar.
