# Feature Specification: Autenticação, Isolamento e Gestão de Equipe

**Feature Branch**: `001-auth-isolamento-equipe`

**Created**: 2026-06-14

**Status**: Draft

**Input**: Epic 1 — O sistema deve garantir o isolamento total de dados entre empresas e permitir que o administrador gerencie os acessos da sua empresa.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Convidar um membro da equipe por e-mail (Priority: P1)

Um administrador de uma empresa precisa dar acesso a um funcionário. Ele informa
o e-mail profissional e o papel do funcionário, e o sistema envia um convite por
e-mail com um link para que o funcionário defina sua própria senha e ative o
acesso.

**Why this priority**: Sem a capacidade de convidar pessoas, apenas o
administrador inicial consegue usar o sistema. É o ponto de entrada de qualquer
equipe e habilita todas as demais features.

**Independent Test**: Pode ser testado de ponta a ponta cadastrando um e-mail
com um papel, confirmando que um convite pendente é criado e que um e-mail de
ativação é disparado — entregando valor (onboarding de equipe) por si só.

**Acceptance Scenarios**:

1. **Given** um administrador autenticado na Empresa A, **When** ele convida o
   e-mail "funcionario@empresa.com" com o papel "Editor", **Then** o sistema
   registra um usuário com status "Pendente" associado à Empresa A **And**
   dispara um e-mail de ativação contendo um link de convite válido por 48 horas.
2. **Given** que o e-mail "funcionario@empresa.com" já está cadastrado na
   Empresa A, **When** o administrador tenta convidá-lo novamente, **Then** o
   sistema rejeita a operação informando conflito de duplicidade e não cria um
   segundo registro.
3. **Given** um usuário com papel "Editor" ou "Viewer", **When** ele tenta
   convidar outra pessoa, **Then** o sistema nega a ação por falta de permissão.

---

### User Story 2 - Ativar conta e definir senha (Priority: P1)

Um funcionário convidado recebe o e-mail, acessa o link de convite e define sua
senha, passando a ter acesso ao sistema com o papel atribuído.

**Why this priority**: O convite (US1) só entrega valor quando o convidado
consegue efetivamente entrar. As duas histórias juntas formam o fluxo mínimo de
onboarding.

**Independent Test**: Pode ser testado consumindo um link de convite válido,
definindo uma senha e autenticando com sucesso como o papel atribuído.

**Acceptance Scenarios**:

1. **Given** um convite válido e não expirado, **When** o convidado define uma
   senha que atende à política mínima, **Then** sua conta passa para status
   "Ativo" e ele consegue autenticar.
2. **Given** um convite expirado (mais de 48 horas) ou já utilizado, **When** o
   convidado tenta ativá-lo, **Then** o sistema recusa a ativação e orienta a
   solicitar novo convite.

---

### User Story 3 - Garantia de isolamento entre empresas (Priority: P1)

Qualquer usuário autenticado só enxerga e manipula os dados da sua própria
empresa. Em nenhuma circunstância dados de uma empresa aparecem para usuários de
outra.

**Why this priority**: É a promessa central de um sistema multi-tenant e um
requisito legal/de confiança. Uma falha aqui compromete todo o produto.

**Independent Test**: Pode ser testado criando dados em duas empresas distintas
e confirmando que cada usuário só acessa os dados da sua empresa, e que tentar
acessar um recurso de outra empresa é negado.

**Acceptance Scenarios**:

1. **Given** usuários autenticados nas Empresas A e B com dados em cada uma,
   **When** o usuário da Empresa A lista qualquer recurso, **Then** apenas dados
   da Empresa A são retornados.
2. **Given** um usuário autenticado na Empresa A, **When** ele tenta acessar
   diretamente um recurso identificado como pertencente à Empresa B, **Then** o
   sistema responde como se o recurso não existisse para ele (acesso negado),
   sem vazar a existência ou o conteúdo do recurso.
3. **Given** qualquer requisição autenticada, **When** o cliente tenta informar
   manualmente um identificador de empresa diferente do seu, **Then** o sistema
   ignora esse valor e usa exclusivamente a empresa derivada da identidade
   autenticada.

### Edge Cases

- Convite para um e-mail que já existe em **outra** empresa: a unicidade é por
  empresa, não global. [NEEDS CLARIFICATION: confirmar se um mesmo e-mail pode
  ter contas em múltiplas empresas ou se a identidade de login é global.]
- Reenvio de convite para um usuário ainda "Pendente": deve renovar/reemitir o
  link sem criar duplicidade.
- Token de convite adulterado ou inválido: ativação negada sem revelar detalhes.
- Administrador é desativado enquanto há convites pendentes que ele emitiu:
  convites permanecem válidos.
- Tentativa de um usuário elevar o próprio papel: negada.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST permitir que um administrador convide um novo
  usuário informando e-mail e papel (`Admin`, `Editor` ou `Viewer`).
- **FR-002**: O sistema MUST criar o usuário convidado com status "Pendente"
  associado à empresa do administrador.
- **FR-003**: O sistema MUST gerar um link/token de convite com validade de 48
  horas e disparar um e-mail de ativação com template profissional.
- **FR-004**: O sistema MUST impedir o cadastro de um e-mail já existente na
  mesma empresa, retornando um erro de conflito.
- **FR-005**: O sistema MUST permitir que o convidado defina sua senha via link
  válido, transicionando o status para "Ativo".
- **FR-006**: O sistema MUST rejeitar ativação por link expirado, já usado ou
  inválido.
- **FR-007**: O sistema MUST autenticar usuários ativos e estabelecer um
  contexto de sessão que carregue a empresa e o papel do usuário.
- **FR-008**: O sistema MUST restringir ações por papel (ex.: apenas `Admin`
  convida/gerencia equipe; `Viewer` tem acesso somente leitura). [NEEDS
  CLARIFICATION: matriz completa de permissões por papel.]
- **FR-009**: O sistema MUST filtrar todo acesso a dados pela empresa do
  contexto autenticado, de forma central e por padrão.
- **FR-010**: O sistema MUST derivar o identificador de empresa exclusivamente
  da identidade autenticada, ignorando qualquer valor de empresa fornecido pelo
  cliente.
- **FR-011**: O sistema MUST tratar senhas de usuário de forma segura (sem
  armazenamento em texto puro) e nunca expor segredos em respostas ou logs.

### Key Entities *(include if feature involves data)*

- **Empresa (Tenant)**: a organização cliente da plataforma. Atributos:
  identidade, razão social, CNPJ, data de criação. Raiz do isolamento de dados.
- **Usuário**: pessoa com acesso ao sistema, pertencente a uma empresa.
  Atributos: nome, e-mail, papel (`Admin`/`Editor`/`Viewer`), status
  (`Pendente`/`Ativo`). Relaciona-se a exatamente uma empresa.
- **Convite**: vínculo temporário que permite a um usuário pendente ativar sua
  conta. Atributos: validade (48h), estado (válido/expirado/usado).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um administrador consegue convidar um novo membro em menos de 1
  minuto, do início ao disparo do e-mail.
- **SC-002**: 100% das tentativas de acesso cross-tenant (a dados de outra
  empresa) são negadas, sem vazamento de existência ou conteúdo.
- **SC-003**: Convites expiram corretamente: 0% dos links com mais de 48 horas
  permitem ativação.
- **SC-004**: 0% de e-mails duplicados criados dentro da mesma empresa.
- **SC-005**: Nenhum segredo (senha, token) aparece em respostas de API ou logs
  em qualquer cenário testado.

## Assumptions

- Existe um serviço de envio de e-mail disponível para disparar os convites.
- A política mínima de senha será definida no plano (comprimento/complexidade).
- A identidade de empresa do usuário é estabelecida no login e não muda durante
  a sessão.
- O **primeiro Admin** e a **Empresa** nascem pelo cadastro self-service (PLG,
  feature 005): o usuário cria a conta pela landing, cadastra a empresa após
  logar e vira `Admin`. Os convites desta feature cobrem os **membros adicionais**
  da equipe; não são o único caminho de entrada.
- SSO/OAuth externo está fora de escopo na v1.
