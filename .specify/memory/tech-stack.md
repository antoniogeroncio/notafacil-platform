# Tech Stack — NotaFácil Platform

> Escolhas tecnológicas concretas que implementam os princípios agnósticos da
> [`constitution.md`](constitution.md). Decisões específicas por feature (schemas,
> endpoints, bibliotecas pontuais) ficam em cada `specs/NNN-*/plan.md`, gerado
> por `/speckit-plan`. Este documento é a referência global da stack.

## Visão geral

| Camada | Tecnologia |
|--------|------------|
| Frontend | Next.js (React, App Router) — MVVM, Princípio XV |
| Backend | Go (Camadas — Princípio I) |
| Persistência | MongoDB — Multi-Tenant Single-Database |
| Infraestrutura local | Containers via Docker Compose (Princípio X) |

## Backend (Go)

- **Camadas (Princípio I):** `Controller → Service → Repository`. Em Go, a
  camada **Controller é o HTTP handler**. Services dependem de **interfaces** de
  repositório; implementações concretas são injetadas (DI manual via
  construtores em `main`/wiring, ou ferramenta de wiring).
- **Organização:** `backend/internal/<domínio>/` com `*_handler.go`,
  `*_service.go`, `*_repository.go`. Services testáveis vivem como
  `*_service.go` (auditados pelo agente `qa`).
- **Isolamento multi-tenant (Princípio III):** o `tenantId` é extraído do token
  de sessão por middleware e injetado no **contexto da requisição**
  (`context.Context`). Os repositórios obtêm o tenant do contexto e o aplicam a
  toda query por padrão — nenhum handler/service passa `tenantId` manualmente a
  partir de input do cliente. Toda coleção com dados de tenant tem índice
  composto que inclui `tenantId`.
- **Driver MongoDB:** driver oficial `mongo-go-driver`.
- **Testes:** `go test ./...`; unidade com `testify/mock` (Princípios IV/V);
  integração com `testcontainers-go` subindo MongoDB efêmero.
- **Segredos fiscais (Princípio VI):** certificados A1, senhas e credenciais de
  API criptografados em repouso; descriptografados apenas em memória no uso.
  Hash de senha de usuário com algoritmo forte (ex.: bcrypt/argon2).

## Integração Fiscal (Strategy)

Padrão **Strategy** selecionado por tenant conforme `authType`:

- `CERTIFICATE` → `CertificateProvider`: descriptografa o A1 em memória, aplica
  assinatura ICP-Brasil no XML e transmite.
- `API_CREDENTIALS` → `ApiCredentialsProvider`: injeta credenciais (token ou
  usuário/senha) nos headers da requisição REST e transmite sem assinatura A1.

O motor de emissão depende da **interface** do provedor; adicionar um novo
provedor não altera o motor (extensibilidade — Epic 4 / Princípio I).

## Frontend (Next.js / React)

- **Framework:** Next.js (App Router) + TypeScript, sem `any`. UI com
  shadcn/ui. Páginas interativas marcadas com `'use client'`.
- **Camadas (Princípio XV):** `Component/Page → Hook/Store → Repository`. A Page
  nunca chama `lib/api/` diretamente.
- **Organização:** `frontend/app/(dashboard)/<feature>/page.tsx` (View),
  `frontend/hooks/use<Feature>.ts` (ViewModel), `frontend/lib/api/`
  (Repository). Hooks são auditados pelo agente `qa`
  (`hooks/__tests__/use<Feature>.test.ts`).
- **Chamadas de API:** apenas no Repository, com `credentials: 'include'`
  (cookie HttpOnly de sessão); DTOs tipados na borda.
- **Testes/lint:** `vitest` (unidade de hooks com Repository mockado),
  `next lint`, `tsc --noEmit`, e golden-path no runtime real do browser
  (Princípio VIII).

## Persistência (MongoDB)

- **Multi-Tenant Single-Database:** isolamento lógico por `tenantId`.
- **Índices:** toda coleção de dados de tenant tem índice composto incluindo
  `tenantId` (isolamento + performance).
- Schemas detalhados por feature vivem nos respectivos `plan.md`.

## API

- REST versionada sob `/api/v1/` (Princípio XII).
- Erros estruturados `{ code, message }`; códigos HTTP semânticos (409
  duplicidade, 401/403 auth/autorização).
- Agregações (totais de nota, financeiro) calculadas no servidor.
- Autenticação: backend emite token de sessão carregando `tenantId` e papel
  (`Admin`/`Editor`/`Viewer`). Convites com token de validade 48h. Sem SSO
  externo na v1.

## Infraestrutura (Princípio X)

- `docker-compose` sobe MongoDB e serviços com um único comando a partir de um
  checkout limpo.
- CI executa sobre a mesma definição de infraestrutura.

## Estrutura de diretórios (alvo)

```
backend/        # Go — handlers, services, repositories, internal/
frontend/       # Next.js — app, components, hooks, lib/api
specs/          # Spec Kit — specs, plans, tasks, contracts
.specify/       # Constituição, tech-stack, templates e scripts do Spec Kit
```

> As pastas `backend/` e `frontend/` surgem a partir de `/speckit-implement`;
> até lá o repositório contém apenas as specs.
