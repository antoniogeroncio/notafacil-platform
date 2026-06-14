# Feature Specification: Integração de Emissão Fiscal via Focus NFe

**Feature Branch**: `004-autenticacao-fiscal`

**Created**: 2026-06-14

**Status**: Draft

**Input**: Epic 4 — A emissão/transmissão fiscal é **delegada à API do Focus NFe** (https://focusnfe.com.br), que abstrai a complexidade de assinatura ICP-Brasil, formatos e integração com os órgãos emissores por município. O sistema foca nas funcionalidades de produto; o provedor cuida da emissão. A integração é encapsulada atrás de uma abstração para permitir trocar/adicionar provedores no futuro.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configurar a empresa para emissão (Priority: P1)

Um administrador habilita sua empresa para emitir, fornecendo os dados fiscais e
o certificado digital A1 (com senha). O sistema registra a empresa no provedor de
emissão (Focus NFe). O certificado e a senha são segredos e ficam protegidos.

**Why this priority**: Sem a empresa habilitada no provedor, nenhuma nota pode
ser emitida. É pré-requisito de todo faturamento.

**Independent Test**: Pode ser testado configurando uma empresa (com certificado
no ambiente de teste do provedor) e confirmando que ela fica apta a emitir e que
nenhum segredo é retornado em texto puro.

**Acceptance Scenarios**:

1. **Given** um administrador na Empresa A, **When** ele envia o certificado A1 e
   a senha e os dados fiscais, **Then** o sistema registra/atualiza a empresa no
   provedor de emissão e marca a Empresa A como apta a emitir, **And** os
   segredos são tratados de forma protegida e nunca retornados em texto puro.
2. **Given** um certificado inválido/expirado ou senha incorreta, **When** o
   provedor recusa o cadastro, **Then** o sistema reporta o erro de forma
   compreensível, sem expor a senha, e a empresa permanece inapta.
3. **Given** uma configuração existente, **When** qualquer usuário consulta a
   configuração fiscal, **Then** o sistema indica o estado (apta/inapta, validade
   do certificado) sem expor os valores secretos.

---

### User Story 2 - Emitir a nota via provedor (Priority: P1)

Ao faturar uma nota válida, o sistema a envia ao provedor (Focus NFe), que
processa a emissão de forma assíncrona. O sistema acompanha o resultado e
apresenta ao usuário o status (autorizada, processando, erro), guardando os
artefatos retornados (protocolo, XML/PDF/DANFSe).

**Why this priority**: É a entrega central da Epic — efetivar a emissão fiscal.
Depende da configuração (US1), da nota montada (Epic 3) e da cota/assinatura
ativa (Epic 5).

**Independent Test**: Pode ser testado faturando uma nota no ambiente de teste do
provedor e verificando que o sistema reflete o status retornado e armazena a
referência da emissão.

**Acceptance Scenarios**:

1. **Given** uma empresa apta, com assinatura ativa e cota disponível (Epic 5),
   **When** o usuário fatura uma nota válida, **Then** o sistema a envia ao
   provedor com uma referência única e marca a nota como "Processando".
2. **Given** uma nota enviada ao provedor, **When** o provedor retorna
   autorização, **Then** o sistema marca a nota como "Emitida" e guarda
   protocolo e artefatos (XML/PDF) para consulta/download.
3. **Given** uma empresa inapta (sem configuração válida no provedor), **When** o
   usuário tenta faturar, **Then** o sistema bloqueia e orienta a configurar a
   emissão.
4. **Given** uma empresa sem assinatura ativa, ou que atingiu o teto mensal de
   emissões (franquia + 1.000 sem liberação manual) (Epic 5), **When** o usuário
   tenta faturar, **Then** o sistema bloqueia e orienta regularização/liberação —
   sem enviar ao provedor. (Acima da franquia, mas abaixo do teto, a emissão é
   permitida em modo pagamento por uso.)
5. **Given** uma recusa/erro do provedor (ex.: dados fiscais inválidos), **When**
   o sistema recebe o erro, **Then** ele marca a nota como "Erro", registra a
   mensagem acionável e não consome cota indevidamente.

---

### User Story 3 - Acompanhar status assíncrono via webhook (Priority: P2)

Como a emissão é assíncrona, o sistema recebe atualizações de status do provedor
(via webhook/callback) e mantém o estado da nota em sincronia, inclusive para
cancelamentos.

**Why this priority**: Garante que o status exibido reflita a realidade do órgão
emissor sem o usuário precisar reabrir/atualizar manualmente; importante mas
posterior ao caminho feliz de emissão.

**Independent Test**: Pode ser testado simulando um callback do provedor para uma
nota "Processando" e confirmando que o status é atualizado de forma idempotente.

**Acceptance Scenarios**:

1. **Given** uma nota "Processando", **When** o sistema recebe o callback de
   autorização do provedor, **Then** atualiza a nota para "Emitida" e anexa os
   artefatos.
2. **Given** um callback recebido mais de uma vez (duplicado/fora de ordem),
   **When** o sistema o processa, **Then** o resultado é **idempotente** (não
   duplica emissão nem corrompe o estado).
3. **Given** uma solicitação de cancelamento aceita pelo provedor, **When** o
   callback chega, **Then** a nota passa a "Cancelada".

### Edge Cases

- Certificado A1 expirado/senha incorreta no cadastro: empresa permanece inapta,
  mensagem clara, senha nunca exposta.
- Indisponibilidade temporária do provedor: emissão fica "Processando" com
  política de retentativa/consulta; nunca trava a UI. [NEEDS CLARIFICATION:
  retentativa/fila e timeout.]
- Município/serviço não suportado pelo provedor: erro acionável ao usuário.
  [NEEDS CLARIFICATION: municípios-alvo na v1.]
- Necessidade futura de outro provedor: a abstração permite adicionar/trocar sem
  alterar o motor de emissão.
- Onde fica o certificado: [NEEDS CLARIFICATION: o certificado é apenas repassado
  ao provedor e não retido pela plataforma, ou também armazenado criptografado
  por nós?]
- Ambientes: separação entre homologação e produção do provedor. [NEEDS
  CLARIFICATION: por tenant ou global.]

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST habilitar uma empresa para emissão registrando-a no
  provedor de emissão (Focus NFe) a partir dos dados fiscais e do certificado A1.
- **FR-002**: O sistema MUST tratar certificado, senha de certificado e
  credenciais do provedor como segredos (criptografados em repouso, em memória só
  no uso), por empresa, e nunca expô-los (Princípio VI).
- **FR-003**: O sistema MUST enviar a nota faturada ao provedor com uma
  referência única e registrar o estado da emissão.
- **FR-004**: O sistema MUST refletir o resultado da emissão (Processando/
  Emitida/Erro/Cancelada) e armazenar os artefatos retornados (protocolo, XML,
  PDF/DANFSe) para consulta e download.
- **FR-005**: O sistema MUST impedir o faturamento quando a empresa está inapta
  (sem configuração válida no provedor).
- **FR-006**: O sistema MUST impedir o faturamento quando a assinatura não está
  ativa ou quando a empresa atingiu o teto mensal de emissões (franquia + 1.000
  sem liberação manual) (integração com Epic 5), sem enviar a nota ao provedor.
  Acima da franquia e abaixo do teto, a emissão segue em modo pagamento por uso.
- **FR-007**: O sistema MUST receber e processar de forma **idempotente** os
  callbacks/webhooks de status do provedor.
- **FR-008**: O sistema MUST encapsular a integração atrás de uma abstração de
  provedor de emissão, permitindo adicionar/trocar o provedor sem modificar o
  motor de emissão (extensibilidade — Princípio I).
- **FR-009**: O sistema MUST expor mensagens de erro acionáveis ao usuário, sem
  vazar segredos nem detalhes técnicos internos.

### Key Entities *(include if feature involves data)*

- **Configuração Fiscal da Empresa**: habilita a emissão. Atributos: estado
  (apta/inapta), validade do certificado, referência da empresa no provedor,
  ambiente (homologação/produção). Segredos (certificado/senha/credenciais do
  provedor) tratados conforme Princípio VI. Pertence a uma empresa.
- **Emissão (Resultado)**: desfecho da emissão de uma nota. Atributos:
  referência, status (Processando/Emitida/Erro/Cancelada), protocolo do órgão,
  artefatos (XML/PDF), mensagem de erro (sem segredos).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% das notas faturadas por empresas aptas e dentro da cota são
  enviadas ao provedor e têm seu status refletido ao usuário.
- **SC-002**: 0% de exposição de segredos fiscais em respostas, logs ou
  mensagens, em qualquer cenário testado.
- **SC-003**: Tentativas de faturar com empresa inapta, assinatura inativa ou
  cota esgotada são bloqueadas em 100% dos casos, com orientação.
- **SC-004**: 100% dos callbacks do provedor processados de forma idempotente.
- **SC-005**: Trocar o provedor de emissão não exige alterar o motor de emissão
  (verificável por inspeção de design e testes).

## Assumptions

- Provedor de emissão na v1: **Focus NFe** (abstrai assinatura ICP-Brasil,
  formatos e integração com órgãos por município). Validação automatizada usa o
  ambiente de homologação do provedor.
- A montagem/validação da nota provém da feature 003; a cota/assinatura, da
  feature 005.
- A configuração fiscal é gerida por administradores (papel da feature 001).
- Endpoints, formatos de payload e eventos de webhook do Focus NFe serão
  detalhados no plano (`plan.md`).
