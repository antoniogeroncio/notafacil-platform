# NotaFácil Platform — Visão Geral (Spec-Driven Development)

**Sistema:** Plataforma Multi-Tenant de Emissão de Notas Fiscais (NFS-e)
**Método:** Spec-Driven Development com [GitHub Spec Kit](https://github.com/github/spec-kit)
**Arquitetura:** Multi-Tenant Single-Database (isolamento lógico por tenant)

> Este documento deixou de ser a especificação. As especificações agora vivem
> nos artefatos do Spec Kit (abaixo). Mantenha-o apenas como índice de navegação.

## Onde estão as specs

| Artefato | Caminho | O que contém |
| --- | --- | --- |
| Constituição | [`.specify/memory/constitution.md`](../.specify/memory/constitution.md) | Princípios de engenharia não-negociáveis (agnósticos de tecnologia) |
| Tech Stack | [`.specify/memory/tech-stack.md`](../.specify/memory/tech-stack.md) | Escolhas concretas (Go, Next.js, MongoDB, Strategy fiscal) que implementam os princípios |
| Epic 1 | [`specs/001-auth-isolamento-equipe/spec.md`](../specs/001-auth-isolamento-equipe/spec.md) | Autenticação, isolamento multi-tenant e gestão de equipe |
| Epic 2 | [`specs/002-cadastros-templates/spec.md`](../specs/002-cadastros-templates/spec.md) | Clientes, catálogo de serviços e templates de nota |
| Epic 3 | [`specs/003-emissao-hibrida/spec.md`](../specs/003-emissao-hibrida/spec.md) | Motor de emissão híbrida (catálogo + digitação livre) |
| Epic 4 | [`specs/004-autenticacao-fiscal/spec.md`](../specs/004-autenticacao-fiscal/spec.md) | Estratégia modular de autenticação/transmissão fiscal |

## Como o conteúdo foi reorganizado (SDD)

No Spec-Driven Development separamos o **quê/porquê** do **como**:

- **Specs** (`specs/NNN-*/spec.md`): comportamento observável e regras de
  negócio (BDD: Given/When/Then), **tech-agnostic**. Não citam MongoDB,
  Go, Next.js ou Strategy.
- **Constituição** (`.specify/memory/constitution.md`): princípios de engenharia
  **agnósticos de tecnologia** (camadas, TDD, isolamento multi-tenant, segurança
  fiscal, backend-first, MVVM no frontend).
- **Tech Stack** (`.specify/memory/tech-stack.md`): as escolhas concretas que
  implementam os princípios (Go, Next.js, MongoDB, Strategy fiscal, papéis).
- **Plano** (`plan.md`, gerado por feature via `/speckit-plan`): decisões
  técnicas por feature — schemas, índices por tenant, contratos de API etc.

Os detalhes técnicos do `init.md` original (schemas Mongo, middleware de
injeção de `tenantId`, classes `CertificateProvider`/`ApiCredentialsProvider`)
foram preservados como **restrições** na constituição e como insumo para os
planos — eles entram em cada `plan.md`, não nas specs.

## Fluxo de trabalho (por feature)

Execute, dentro do Claude Code, na ordem:

1. `/speckit-constitution` — manter os princípios do projeto.
2. `/speckit-specify` — criar/editar a spec (tech-agnostic).
3. `/speckit-clarify` — resolver os `[NEEDS CLARIFICATION]` antes de planejar.
4. `/speckit-plan` — definir stack/estrutura conforme a constituição.
5. `/speckit-tasks` — gerar tasks executáveis.
6. `/speckit-analyze` — checar consistência entre spec, plan e tasks.
7. `/speckit-implement` — implementar com testes.

## Pontos em aberto

As specs contêm marcações `[NEEDS CLARIFICATION]` (ex.: identidade de e-mail
global vs por empresa, matriz de permissões por papel, municípios-alvo na v1,
política de retentativa de transmissão). Resolva-as com `/speckit-clarify`
antes de avançar para o plano de cada feature.

## Nota de segurança

O `specify init` instalou integrações em `.claude/`. Como agentes podem
armazenar credenciais nessa pasta, considere adicionar `.claude/` (ou partes
dela) ao `.gitignore` antes de commitar.
