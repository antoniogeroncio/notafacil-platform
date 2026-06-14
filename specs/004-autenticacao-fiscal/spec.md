# Feature Specification: Estratégia de Autenticação Fiscal Modular

**Feature Branch**: `004-autenticacao-fiscal`

**Created**: 2026-06-14

**Status**: Draft

**Input**: Epic 4 — O motor de faturamento deve decidir como se autenticar e transmitir a nota conforme a configuração fiscal de cada empresa (certificado A1 vs credenciais de API), de forma modular e extensível.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Configurar o método de autenticação fiscal da empresa (Priority: P1)

Um administrador configura como a sua empresa se autentica junto ao órgão
emissor: ou enviando um certificado digital A1 (com senha), ou informando
credenciais de API (ex.: chave, cliente/segredo ou usuário/senha). Essas
credenciais são segredos e ficam protegidas.

**Why this priority**: Sem uma configuração fiscal válida nenhuma nota pode ser
transmitida. É pré-requisito de todo faturamento e define qual estratégia será
usada.

**Independent Test**: Pode ser testado configurando cada método para uma empresa
e confirmando que o segredo é armazenado de forma protegida e nunca retornado em
texto puro.

**Acceptance Scenarios**:

1. **Given** um administrador na Empresa A, **When** ele configura autenticação
   por certificado A1 enviando o arquivo e a senha, **Then** o certificado e a
   senha são armazenados protegidos (criptografados/hash) e associados à Empresa
   A, **And** nunca são retornados em texto puro em respostas ou logs.
2. **Given** um administrador na Empresa A, **When** ele configura autenticação
   por credenciais de API, **Then** as credenciais são armazenadas protegidas e
   associadas à Empresa A.
3. **Given** uma configuração fiscal existente, **When** qualquer usuário
   consulta a configuração, **Then** o sistema indica o método ativo sem expor
   os valores secretos.

---

### User Story 2 - Transmitir a nota pela estratégia correta (Priority: P1)

Ao faturar uma nota, o motor seleciona automaticamente a estratégia de
autenticação/transmissão conforme o método configurado na empresa: assinatura
ICP-Brasil com o certificado A1, ou transmissão via credenciais de API sem
assinatura A1. O usuário não precisa saber qual caminho foi usado.

**Why this priority**: É a entrega central da Epic — efetivar a emissão fiscal.
Depende da configuração (US1) e da nota montada (Epic 3).

**Independent Test**: Pode ser testado faturando uma nota em uma empresa
configurada com A1 e outra configurada com API, verificando que cada uma usa o
caminho de transmissão correto e produz um resultado de emissão.

**Acceptance Scenarios**:

1. **Given** uma empresa configurada com certificado A1, **When** o sistema
   processa o faturamento de uma nota válida, **Then** ele usa a estratégia de
   certificado: descriptografa o A1 em memória, aplica a assinatura ICP-Brasil
   ao documento e o transmite.
2. **Given** uma empresa configurada com credenciais de API, **When** o sistema
   processa o faturamento de uma nota válida, **Then** ele usa a estratégia de
   credenciais: autentica via as credenciais e transmite o documento sem
   assinatura A1.
3. **Given** uma empresa sem configuração fiscal válida, **When** o usuário tenta
   faturar, **Then** o sistema impede a transmissão e orienta a configurar a
   autenticação fiscal.
4. **Given** uma falha/recusa do órgão emissor na transmissão, **When** o
   sistema recebe o erro, **Then** ele registra o resultado e expõe ao usuário um
   status compreensível, sem vazar segredos.

### Edge Cases

- Certificado A1 expirado ou senha incorreta: transmissão bloqueada com
  mensagem clara, sem expor a senha.
- Credenciais de API inválidas/revogadas: falha tratada e reportada.
- Troca do método de autenticação de uma empresa com notas em andamento: notas
  já transmitidas não são afetadas; novas usam o método atual.
- Necessidade de um novo provedor/órgão emissor no futuro: a arquitetura deve
  permitir adicionar uma estratégia sem alterar o motor de emissão. [NEEDS
  CLARIFICATION: lista de órgãos/municípios a suportar na v1.]
- Indisponibilidade temporária do órgão emissor: [NEEDS CLARIFICATION: política
  de retentativa/fila para transmissão.]

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Cada empresa MUST poder configurar exatamente um método de
  autenticação fiscal ativo: certificado A1 ou credenciais de API.
- **FR-002**: O sistema MUST armazenar certificados, senhas de certificado e
  credenciais de API de forma protegida (criptografados/hash) e por empresa.
- **FR-003**: O sistema MUST NOT expor segredos fiscais em respostas de API,
  logs, mensagens de erro ou telemetria.
- **FR-004**: O motor de faturamento MUST selecionar a estratégia de
  autenticação/transmissão a partir do método configurado na empresa.
- **FR-005**: Na estratégia de certificado, o sistema MUST descriptografar o A1
  apenas em memória, aplicar a assinatura ICP-Brasil ao documento e transmiti-lo.
- **FR-006**: Na estratégia de credenciais de API, o sistema MUST autenticar com
  as credenciais e transmitir o documento sem assinatura A1.
- **FR-007**: O sistema MUST impedir o faturamento quando não há configuração
  fiscal válida.
- **FR-008**: O sistema MUST registrar o resultado da transmissão (sucesso/erro)
  e expor um status compreensível ao usuário.
- **FR-009**: A arquitetura MUST permitir adicionar novas estratégias de
  provedor sem modificar o motor de emissão (extensibilidade).

### Key Entities *(include if feature involves data)*

- **Configuração Fiscal da Empresa**: define como a empresa se autentica.
  Atributos: tipo de autenticação (`CERTIFICATE` | `API_CREDENTIALS`),
  certificado A1 criptografado, hash da senha do certificado, e/ou credenciais
  de API (ex.: chave, cliente, segredo). Pertence a uma empresa.
- **Resultado de Emissão**: registro do desfecho da transmissão de uma nota.
  Atributos: status, identificador de protocolo/retorno do órgão (quando
  houver), mensagem de erro (sem segredos).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% das notas faturadas usam a estratégia correspondente ao método
  configurado na sua empresa.
- **SC-002**: 0% de exposição de segredos fiscais em respostas, logs ou
  mensagens, em qualquer cenário testado.
- **SC-003**: Tentativas de faturar sem configuração válida são bloqueadas em
  100% dos casos com orientação ao usuário.
- **SC-004**: Adicionar um novo provedor de transmissão não exige alteração no
  motor de emissão (verificável por inspeção de design e testes).

## Assumptions

- A montagem/validação da nota provém da feature 003.
- A configuração fiscal é gerida por administradores (papel da feature 001).
- Os formatos de documento fiscal e os endpoints dos órgãos emissores serão
  detalhados no plano, conforme os municípios-alvo.
