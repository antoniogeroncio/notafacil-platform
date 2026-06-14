# Research — 001 Autenticação, Isolamento e Gestão de Equipe

Decisões técnicas que resolvem o Technical Context do plano. Cada item:
**Decisão / Justificativa / Alternativas**.

## 1. Mecanismo de isolamento multi-tenant (Princípio III)

- **Decisão**: Middleware de autenticação extrai `tenantId` (e `userId`/`role`)
  do JWT de sessão e os injeta no `context.Context`. Os repositórios obtêm o
  `tenantId` **do contexto** e o aplicam a toda query/insert por meio de um
  helper `tenantScoped(ctx, filter)` que adiciona `{tenantId: ...}`. Handlers e
  services nunca recebem `tenantId` do corpo/query.
- **Justificativa**: filtro central e por padrão — não depende de o dev lembrar;
  cumpre FR-009/FR-010. Falha fechada: sem `tenantId` no contexto, o repositório
  retorna erro.
- **Alternativas**: filtro manual por query (rejeitado: fácil de esquecer →
  vazamento); banco-por-tenant (rejeitado: constituição define single-database).

## 2. Token de sessão

- **Decisão**: JWT assinado (HS256) com claims `sub` (userId), `tid` (tenantId),
  `role`, `exp` curto (ex.: 30 min) + refresh httpOnly. Entregue ao frontend em
  **cookie httpOnly + SameSite=Strict** (não em localStorage).
- **Justificativa**: cookie httpOnly protege contra XSS; claims carregam o
  contexto de tenant/role exigido por FR-007. HS256 simples para v1.
- **Alternativas**: sessão server-side em Mongo (mais estado; adiável); tokens
  em localStorage (rejeitado: risco XSS).

## 3. Token de convite (48h)

- **Decisão**: token aleatório de alta entropia (256 bits) enviado no link; no
  banco guarda-se apenas o **hash** (SHA-256) do token, com `expiresAt` (+48h) e
  `usedAt`. Validação compara hash; expira por tempo ou uso.
- **Justificativa**: cumpre FR-003/FR-006 e Princípio VI (não armazenar o
  segredo em texto puro). Hash permite invalidar sem guardar o token.
- **Alternativas**: JWT de convite (rejeitado: não revogável facilmente antes do
  exp); guardar token puro (rejeitado: vaza se o banco vazar).

## 4. Hash de senha

- **Decisão**: `bcrypt` (custo 12) via `golang.org/x/crypto/bcrypt`.
- **Justificativa**: padrão consolidado, salt embutido; cumpre FR-011/Princípio VI.
- **Alternativas**: argon2id (ótimo, porém mais conf.); deixado como evolução.

## 5. Política mínima de senha

- **Decisão**: mínimo 8 caracteres, com ao menos letras e números; rejeitar
  senhas obviamente fracas. Validada no service de ativação (US2).
- **Justificativa**: equilíbrio usabilidade/segurança para v1.
- **Alternativas**: regras mais rígidas/zxcvbn (evolução futura).

## 6. Roteador HTTP e estrutura

- **Decisão**: `chi` para roteamento + middlewares; `mongo-go-driver` oficial.
- **Justificativa**: leve, idiomático, fácil de compor middlewares (auth/tenant).
- **Alternativas**: gin/echo (mais peso); net/http puro (mais boilerplate).

## 7. Envio de e-mail (convite)

- **Decisão**: interface `EmailSender` injetada no service (Princípio I). Impl
  v1 via provedor transacional SMTP/API; em dev/CI, um `FakeSender` que captura
  mensagens para asserção nos testes.
- **Justificativa**: desacopla o domínio do provedor; testes não dependem de
  rede. Provedor concreto é detalhe de configuração.
- **Alternativas**: chamar SDK direto no service (rejeitado: viola Princípio I).

## 8. Estratégia de testes

- **Decisão**: unidade nos Services com repositórios/EmailSender mockados
  (`testify/mock`); integração de Handlers/Repos com `testcontainers-go` subindo
  MongoDB efêmero; **teste de violação** dedicado (usuário do Tenant A não acessa
  recurso do Tenant B → not-found); **golden-path** da tela de ativação no
  browser com API mockada.
- **Justificativa**: cumpre Princípios IV/V/VIII e os gates do agente `qa`.
- **Alternativas**: só unidade (rejeitado: não cobre isolamento real no Mongo).

## 9. Relação com o cadastro self-service (feature 005)

- **Decisão**: esta feature expõe os serviços de Usuário/Empresa/sessão; a
  criação do **primeiro Admin + Empresa** pela landing é orquestrada na 005,
  reutilizando os mesmos services. Convite cobre membros adicionais.
- **Justificativa**: evita duplicar identidade; mantém 1 conta = 1 empresa (v1).
- **Alternativas**: signup próprio na 001 (rejeitado: pertence ao escopo PLG).
