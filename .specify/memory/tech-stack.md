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

## Emissão Fiscal (Focus NFe)

A emissão/transmissão é **delegada à API do Focus NFe** (https://focusnfe.com.br),
que abstrai a assinatura ICP-Brasil, os formatos e a integração com os órgãos
emissores por município. O produto foca em funcionalidades; o provedor cuida da
emissão.

- **Abstração:** o motor de emissão depende de uma **interface**
  `FiscalEmissionProvider`; a v1 implementa `FocusNFeProvider` (`pkg/focusnfe`).
  Trocar/adicionar provedor não altera o motor (extensibilidade — Princípio I).
- **Fluxo:** registrar a empresa no provedor (certificado A1 + dados fiscais) →
  enviar a nota com referência única → emissão **assíncrona** → status via
  consulta e/ou **webhook** (idempotente) → guardar protocolo, XML e PDF/DANFSe.
- **Segredos (Princípio VI):** certificado A1, senha e token do provedor são
  criptografados em repouso e nunca expostos. [Definir no plano: o certificado é
  apenas repassado ao Focus NFe ou também retido por nós.]
- **Ambientes:** homologação (testes automatizados) e produção do provedor.

## Cobrança / Pagamentos (PagSeguro) — Epic 5

- **Gateway:** **PagSeguro**, encapsulado como integração de infra (`pkg/pagseguro`)
  injetada no Service (Princípio I). Cartão de crédito recorrente, Pix e demais
  formas do gateway.
- **PCI / segredos (Princípio VI):** **não** armazenar dados de cartão; usar
  tokenização/checkout do gateway; guardar apenas referências não sensíveis.
- **Webhooks idempotentes:** notificações do PagSeguro atualizam status de
  pagamento/assinatura; reprocessar o mesmo evento não muda o resultado.
- **Planos & medição:** assinatura recorrente por empresa/CNPJ; franquia mensal
  por plano (100 / 400 / 4.000 / sob demanda). Política de cota: **só emite com
  plano ativo**; acima da franquia → **pagamento por uso** (aviso a cada 100
  notas extras) até o teto de **franquia + 1.000**; além do teto, **bloqueio**
  que só é liberado por **administrador do sistema** (papel de plataforma,
  distinto do `Admin` do tenant). Medição por mês de competência, checada antes
  de faturar (gating em Epics 3/4).

## Frontend (Next.js / React)

- **Framework:** Next.js (App Router) + TypeScript, sem `any`. Páginas
  interativas marcadas com `'use client'`.
- **Design/UX:** UI com **shadcn/ui + Tailwind**, **mobile-first responsivo**,
  **acessível (WCAG AA)** e com **dark mode** desde o início. Layout moderno e
  sóbrio (SaaS fiscal B2B). As diretrizes operacionais de design (tokens,
  estados de tela, UX de formulários, padrões fiscais) vivem na skill
  `ux-design`.
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
