# NotaFácil Platform Constitution

Plataforma Multi-Tenant de Emissão de Notas Fiscais (NFS-e).
Este documento é a fonte de verdade dos princípios não-negociáveis do produto.
Specs descrevem o **quê/porquê** (tech-agnostic); os planos (`plan.md`) descrevem
o **como**. Esta constituição governa ambos.

## Core Principles

### I. Isolamento Multi-Tenant (NÃO-NEGOCIÁVEL)

Nenhuma requisição, query, job ou log pode vazar dados de um tenant para outro.
Toda entidade de domínio pertence a exatamente um tenant e carrega seu
identificador de tenant. Todo acesso a dados DEVE ser filtrado pelo tenant do
contexto autenticado, por padrão e de forma central — nunca dependendo de o
desenvolvedor lembrar de adicionar o filtro em cada query. O identificador de
tenant é derivado da identidade autenticada (token), nunca de input do cliente
(body, query string, header arbitrário). Violar o isolamento é o defeito mais
grave possível e bloqueia qualquer release.

### II. Segurança de Credenciais Fiscais (NÃO-NEGOCIÁVEL)

Certificados digitais (A1), senhas de certificado e credenciais de API de
terceiros são segredos. Eles DEVEM ser armazenados criptografados em repouso,
descriptografados apenas em memória no momento do uso e nunca expostos em
respostas de API, logs, mensagens de erro ou telemetria. Senhas de usuário e de
certificado são sempre tratadas via hash/criptografia apropriada — nunca em
texto puro. O acesso a esses segredos é restrito ao tenant proprietário (ver
Princípio I).

### III. Especificação Antes de Implementação

Toda feature começa por uma spec (`specs/NNN-*/spec.md`) tech-agnostic,
seguida de um plano (`plan.md`) e de tasks (`tasks.md`). Código sem spec
correspondente não é aceito. As specs descrevem comportamento observável e
regras de negócio (BDD: Given/When/Then); decisões de stack e estrutura vivem no
plano. Ambiguidades são marcadas como `[NEEDS CLARIFICATION]` e resolvidas antes
do planejamento.

### IV. Test-First em Regras Críticas (NÃO-NEGOCIÁVEL)

Isolamento de tenant, autenticação, autorização por papel e o motor de emissão
fiscal DEVEM ter testes automatizados escritos antes da implementação, cobrindo
explicitamente os cenários de violação (ex.: tentativa de acesso cross-tenant
deve falhar). Nenhuma regra de negócio crítica é considerada pronta sem teste
que falharia se a regra fosse removida.

### V. Contratos de API Explícitos e Estáveis

A fronteira entre frontend e backend é um contrato de API versionado e
documentado. Erros usam códigos HTTP semânticos e payloads consistentes (ex.:
409 para conflito de duplicidade, 401/403 para auth/autorização). Mudanças de
contrato que quebram clientes exigem versionamento. O frontend nunca assume
regras fiscais não expressas no contrato.

## Restrições Técnicas

A stack é uma restrição de governança (decisões detalhadas ficam em cada
`plan.md`):

- **Frontend:** React.
- **Backend:** Go.
- **Persistência:** MongoDB, padrão Multi-Tenant Single-Database (isolamento
  lógico por identificador de tenant). Toda coleção que contém dados de tenant
  DEVE ter índice que inclua o identificador de tenant para garantir isolamento
  e performance.
- **Autorização:** modelo de papéis `Admin`, `Editor`, `Viewer`.
- **Integração fiscal:** padrão Strategy para múltiplos provedores de
  autenticação/transmissão (certificado A1 ICP-Brasil e credenciais de API),
  selecionável por tenant, sem acoplar o motor de emissão a um provedor
  específico.

Compliance: conformidade com requisitos legais de NFS-e e padrão ICP-Brasil
quando aplicável.

## Fluxo de Desenvolvimento (Spec Kit)

Ordem canônica por feature:

1. `/speckit-constitution` — manter estes princípios.
2. `/speckit-specify` — criar/editar a spec (tech-agnostic).
3. `/speckit-clarify` — resolver ambiguidades antes de planejar.
4. `/speckit-plan` — definir stack/estrutura conforme estas restrições.
5. `/speckit-tasks` — derivar tasks executáveis.
6. `/speckit-analyze` — checar consistência entre artefatos.
7. `/speckit-implement` — implementar com testes.

Cada feature é um slice independentemente testável e entregável (ver as user
stories priorizadas em cada spec).

## Governance

Esta constituição supersede outras práticas. Qualquer PR ou revisão DEVE
verificar conformidade com os princípios, em especial os NÃO-NEGOCIÁVEIS
(I, II, IV). Complexidade adicional precisa ser justificada explicitamente no
plano. Emendas exigem: registro da mudança, justificativa e atualização da
versão e da data abaixo.

Versionamento desta constituição segue SemVer: MAJOR para remoção/redefinição
incompatível de princípios, MINOR para novos princípios/seções, PATCH para
ajustes de redação.

**Version**: 1.1.0 | **Ratified**: 2026-06-14 | **Last Amended**: 2026-06-14
