# Feature Specification: PLG — Cadastro Self-Service, Planos e Cobrança

**Feature Branch**: `005-plg-planos-cobranca`

**Created**: 2026-06-14

**Status**: Draft

**Input**: Modelo Product-Led Growth — qualquer pessoa pode contratar o sistema pela landing page: cria sua conta, cadastra a empresa após logar e assina o plano que melhor atende, pagando via gateway PagSeguro (cartão, Pix e demais formas). Planos baseados no volume de emissão de notas por CNPJ.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Criar conta a partir da landing page (Priority: P1)

Um visitante sem cadastro acessa a landing page, escolhe começar e cria sua
conta (nome, e-mail, senha) de forma self-service, sem depender de convite.

**Why this priority**: É a porta de entrada do modelo PLG — sem auto-cadastro,
não há aquisição. Habilita todo o funil de contratação.

**Independent Test**: Pode ser testado criando uma conta nova pela landing e
autenticando com ela, sem nenhum convite prévio.

**Acceptance Scenarios**:

1. **Given** um visitante sem cadastro, **When** ele preenche nome, e-mail e
   senha válidos na landing, **Then** o sistema cria a conta e o autentica (ou
   solicita verificação de e-mail, conforme política).
2. **Given** um e-mail já cadastrado, **When** o visitante tenta criar conta com
   ele, **Then** o sistema informa que já existe conta e oferece login/recuperação.
3. **Given** uma senha que não atende à política mínima, **When** o visitante
   submete, **Then** o sistema rejeita com orientação clara.

---

### User Story 2 - Cadastrar a empresa após login (Priority: P1)

No primeiro acesso, o usuário recém-cadastrado informa os dados da sua empresa
(razão social, CNPJ, etc.) e passa a ser o `Admin` dessa empresa.

**Why this priority**: A empresa (tenant) é o que será cobrado e o que emitirá
notas. Sem ela, não há assinatura nem emissão.

**Independent Test**: Pode ser testado, com um usuário logado sem empresa,
cadastrando uma empresa e confirmando que o usuário vira `Admin` dela e passa a
operar isolado nesse tenant.

**Acceptance Scenarios**:

1. **Given** um usuário autenticado e ainda sem empresa, **When** ele cadastra
   uma empresa com CNPJ válido, **Then** o sistema cria o tenant, vincula o
   usuário como `Admin` e estabelece o isolamento (Princípio III).
2. **Given** um CNPJ já cadastrado na plataforma, **When** o usuário tenta
   cadastrá-lo, **Then** o sistema rejeita a duplicidade. [NEEDS CLARIFICATION:
   um mesmo CNPJ pode existir em mais de uma conta? esperado que não.]
3. **Given** um usuário sem empresa, **When** ele acessa áreas que exigem tenant
   (emissão, cadastros), **Then** o sistema o direciona a cadastrar a empresa
   primeiro.

---

### User Story 3 - Assinar um plano e pagar via PagSeguro (Priority: P1)

O `Admin` da empresa compara os planos, escolhe um e assina, pagando pelo
gateway PagSeguro (cartão de crédito recorrente, Pix ou demais formas do
gateway). Ao confirmar o pagamento, a assinatura fica ativa.

**Why this priority**: É a conversão (receita) e o que libera a emissão dentro do
limite contratado. Núcleo do PLG.

**Independent Test**: Pode ser testado assinando um plano em ambiente de teste do
gateway e confirmando que a assinatura fica `Ativa` após a confirmação de
pagamento, sem armazenar dados sensíveis de cartão.

**Acceptance Scenarios**:

1. **Given** um `Admin` com empresa cadastrada, **When** ele escolhe um plano e
   conclui o pagamento via PagSeguro, **Then** o sistema cria uma assinatura
   `Ativa` para a empresa, com o limite mensal e o preço do plano.
2. **Given** um pagamento recusado/não concluído, **When** o gateway retorna
   falha, **Then** a assinatura não é ativada e o usuário recebe orientação para
   tentar novamente.
3. **Given** o gateway processa o pagamento de forma assíncrona (ex.: Pix),
   **When** o sistema recebe a confirmação via webhook do PagSeguro, **Then** ele
   ativa/atualiza a assinatura correspondente.
4. **Given** qualquer fluxo de pagamento, **When** o usuário informa dados de
   cartão, **Then** esses dados trafegam direto ao gateway e **não** são
   armazenados pela plataforma (apenas referência/token do gateway).

---

### User Story 4 - Respeitar o limite de emissões do plano (Priority: P2)

Cada plano define um teto de notas por mês por CNPJ. Conforme a empresa emite, o
sistema contabiliza o uso e impede emissões acima do limite contratado,
orientando o upgrade.

**Why this priority**: É o que dá sentido econômico aos planos (cobrança por
volume). Depende da emissão (Epics 3/4) e da assinatura (US3).

**Independent Test**: Pode ser testado emitindo até o limite de um plano e
confirmando que a próxima emissão é bloqueada/orientada a upgrade, e que o
contador zera no novo mês de competência.

**Acceptance Scenarios**:

1. **Given** uma empresa com assinatura ativa e uso abaixo do limite mensal,
   **When** ela emite uma nota, **Then** a emissão é permitida e o contador de
   uso do mês é incrementado.
2. **Given** uma empresa que atingiu o limite mensal do seu plano, **When** ela
   tenta emitir outra nota, **Then** o sistema bloqueia e oferece upgrade de
   plano. [NEEDS CLARIFICATION: permitir excedente cobrado (overage) ou bloquear
   estritamente?]
3. **Given** a virada do mês de competência, **When** um novo mês começa, **Then**
   o contador de uso reinicia conforme o limite do plano.
4. **Given** uma assinatura inadimplente ou cancelada, **When** a empresa tenta
   emitir, **Then** a emissão é bloqueada com orientação de regularização.

---

### User Story 5 - Plano sob demanda com consultor (Priority: P3)

Empresas com volume acima dos planos padrão (ou necessidades específicas) podem
solicitar o plano sob demanda, registrando um contato para que um consultor
prossiga (fluxo assistido por vendas).

**Why this priority**: Cobre o topo do mercado, mas é fluxo de menor frequência e
não bloqueia o self-service. Pode ser entregue depois.

**Independent Test**: Pode ser testado solicitando o plano sob demanda e
confirmando que um lead/contato comercial é registrado para acompanhamento.

**Acceptance Scenarios**:

1. **Given** um `Admin` interessado em volume acima de 4.000 notas/mês ou
   condições específicas, **When** ele escolhe "Falar com consultor", **Then** o
   sistema registra a solicitação (lead) com os dados da empresa para contato
   comercial.

### Edge Cases

- Conta criada mas e-mail nunca verificado: [NEEDS CLARIFICATION: exigir
  verificação antes de assinar/emitir?]
- Usuário com mais de uma empresa: cada empresa/CNPJ tem assinatura e cota
  próprias. [NEEDS CLARIFICATION: confirmar se uma conta pode ter várias
  empresas.]
- Cancelamento no meio do ciclo: acesso permanece até o fim do período pago?
  [NEEDS CLARIFICATION: política de proration/cancelamento.]
- Falha de renovação recorrente (cartão expirado): período de carência antes de
  suspender emissão. [NEEDS CLARIFICATION: dias de carência.]
- Webhook do PagSeguro duplicado ou fora de ordem: processamento idempotente.
- Downgrade para um plano com limite menor que o uso atual do mês: definir
  comportamento (bloquear até o próximo ciclo?).
- Mudança de plano (upgrade) no meio do mês: a cota disponível reflete o novo
  plano imediatamente.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST permitir cadastro self-service de conta a partir de
  página pública (landing), sem convite.
- **FR-002**: O sistema MUST impedir cadastro de conta com e-mail já existente.
- **FR-003**: O sistema MUST permitir que um usuário autenticado sem empresa
  cadastre uma empresa (tenant) e seja vinculado a ela como `Admin`.
- **FR-004**: O sistema MUST impedir o cadastro de empresa com CNPJ já existente
  na plataforma.
- **FR-005**: O sistema MUST exibir o catálogo de planos com seus limites e
  preços e permitir ao `Admin` assinar um plano.
- **FR-006**: O sistema MUST processar o pagamento da assinatura via gateway
  **PagSeguro**, suportando cartão de crédito (recorrente) e Pix, além das
  demais formas oferecidas pelo gateway.
- **FR-007**: O sistema MUST NOT armazenar dados sensíveis de cartão; deve usar
  tokenização/checkout do gateway e guardar apenas referências não sensíveis
  (Princípio VI).
- **FR-008**: O sistema MUST receber e processar, de forma **idempotente**,
  notificações (webhooks) do PagSeguro para atualizar o status de pagamento e da
  assinatura.
- **FR-009**: O sistema MUST manter o estado da assinatura por empresa
  (`Ativa`, `Inadimplente`, `Cancelada`, etc.) e o ciclo/renovação.
- **FR-010**: O sistema MUST contabilizar o número de notas emitidas por CNPJ por
  mês de competência (medição de uso).
- **FR-011**: O sistema MUST bloquear a emissão quando o uso do mês atinge o
  limite do plano ou quando a assinatura não está ativa, orientando upgrade/
  regularização (integra-se às Epics 3/4).
- **FR-012**: O sistema MUST reiniciar a contagem de uso a cada novo mês de
  competência, conforme o limite vigente do plano.
- **FR-013**: O sistema MUST permitir solicitar o plano sob demanda, registrando
  um lead/contato comercial.
- **FR-014**: O sistema MUST permitir ao `Admin` mudar de plano (upgrade/
  downgrade) e refletir limite e cobrança conforme a política definida.
- **FR-015**: Operações de assinatura e cobrança MUST ser restritas ao papel
  `Admin` da empresa (Princípio III / RBAC da feature 001).

### Planos (catálogo inicial)

| Plano | Limite mensal de emissões por CNPJ | Preço mensal | Contratação |
|-------|-----------------------------------:|-------------:|-------------|
| Plano 1 | até 100 notas | R$ 100 | Self-service |
| Plano 2 | até 400 notas | R$ 300 | Self-service |
| Plano 3 | até 4.000 notas | R$ 500 | Self-service |
| Sob Demanda | acima de 4.000 / customizado | sob consulta | Consultor (vendas) |

> [NEEDS CLARIFICATION: nomes comerciais dos planos; se os valores são com
> impostos inclusos; ciclo apenas mensal ou também anual.]

### Key Entities *(include if feature involves data)*

- **Conta/Usuário**: estende a feature 001 — agora pode nascer via self-service
  (sem convite), inicialmente sem empresa.
- **Empresa (Tenant)**: criada pelo usuário após login (feature 001); passa a ter
  uma assinatura e uma cota de emissão.
- **Plano**: item do catálogo. Atributos: nome, limite mensal de emissões, preço
  mensal, tipo (self-service | sob-demanda).
- **Assinatura**: vínculo empresa↔plano. Atributos: status, data de início,
  próxima renovação, referência do gateway.
- **Pagamento/Fatura**: cobrança de um ciclo. Atributos: valor, método (cartão/
  Pix/...), status, identificador no PagSeguro. Sem dados sensíveis de cartão.
- **Uso Mensal (Medição)**: contador de notas emitidas por empresa/CNPJ por mês
  de competência, comparado ao limite do plano.
- **Lead Comercial**: solicitação de plano sob demanda para contato de vendas.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um visitante conclui conta → empresa → assinatura paga em menos de
  10 minutos, sem intervenção humana (planos self-service).
- **SC-002**: 100% das emissões acima do limite do plano ou com assinatura
  inativa são bloqueadas com orientação.
- **SC-003**: 0% de armazenamento de dados sensíveis de cartão na plataforma.
- **SC-004**: 100% dos webhooks do PagSeguro processados de forma idempotente
  (reprocessar o mesmo evento não altera o resultado).
- **SC-005**: O contador de uso reflete corretamente as emissões do mês em 100%
  dos casos testados e reinicia na virada de competência.

## Assumptions

- Gateway de pagamento: **PagSeguro**. Ambiente de testes (sandbox) usado para
  validação automatizada.
- A emissão de notas (Epics 3/4) consulta a assinatura/cota antes de faturar.
- A identidade de empresa e o RBAC vêm da feature 001; o isolamento segue o
  Princípio III.
- Cobrança recorrente mensal; Pix tratado conforme suporte do gateway
  (confirmação assíncrona via webhook).
- A landing page pública faz parte do frontend (Next.js), separada do app
  autenticado.
