# Feature Specification: Cadastro de Entidades e Templates

**Feature Branch**: `002-cadastros-templates`

**Created**: 2026-06-14

**Status**: Draft

**Input**: Epic 2 — O sistema deve armazenar os registros base (clientes, serviços, modelos) para acelerar o preenchimento do faturamento.

## Clarifications

### Session 2026-06-14

Itens de menor impacto resolvidos por **padrão** (sem pergunta dedicada);
ajustáveis no `/speckit-plan` se necessário:

- D: Campos obrigatórios mínimos → cliente: razão social/nome + documento (CNPJ/CPF); serviço: descrição + valor unitário + classificação fiscal. (Conjunto completo detalhado no plano.)
- D: Dois clientes com o mesmo documento na mesma empresa → **avisar duplicidade, mas permitir** (não bloquear).
- D: Salvar template com nome já existente na empresa → **bloquear** nome duplicado (unicidade de nome por empresa).

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cadastrar clientes e serviços da empresa (Priority: P1)

Um usuário com permissão de edição cadastra e mantém os clientes e o catálogo de
produtos/serviços da sua empresa, para reutilizá-los na emissão de notas. Todos
os registros pertencem exclusivamente à empresa do usuário.

**Why this priority**: Os cadastros são a base de dados que alimenta a emissão.
Sem clientes e serviços cadastrados não há aceleração de preenchimento — é o
primeiro valor tangível da Epic.

**Independent Test**: Pode ser testado cadastrando, listando, editando e
removendo clientes e serviços em uma empresa, e confirmando que outra empresa
não vê esses registros.

**Acceptance Scenarios**:

1. **Given** um usuário com permissão de edição na Empresa A, **When** ele
   cadastra um cliente ou um produto/serviço, **Then** o registro é persistido
   associado obrigatoriamente à Empresa A.
2. **Given** registros cadastrados na Empresa A, **When** um usuário da Empresa
   B lista clientes ou serviços, **Then** nenhum registro da Empresa A é
   retornado.
3. **Given** um usuário com permissão somente leitura, **When** ele tenta criar
   ou editar um cadastro, **Then** a ação é negada por falta de permissão.

---

### User Story 2 - Salvar um modelo (template) de nota recorrente (Priority: P2)

A partir do formulário de emissão, o usuário salva a estrutura atual como um
modelo reutilizável para serviços recorrentes — incluindo itens padrão,
retenções e tributação — omitindo os campos variáveis (cliente e valor).

**Why this priority**: Acelera significativamente a emissão recorrente, mas
depende dos cadastros e do formulário de emissão (Epic 3) para entregar valor
pleno; por isso vem após o cadastro base.

**Independent Test**: Pode ser testado preenchendo uma estrutura de nota,
salvando-a como template e confirmando que o template guarda os campos fixos e
não guarda cliente nem valor.

**Acceptance Scenarios**:

1. **Given** um usuário preenchendo uma nota, **When** ele escolhe "Salvar como
   Template" e nomeia o modelo, **Then** o sistema persiste a estrutura atual
   (itens padrão, retenções padrão, tributação padrão) na empresa do usuário,
   **And** não armazena os campos variáveis cliente e valor.
2. **Given** um template salvo, **When** o usuário inicia uma nova emissão a
   partir dele, **Then** os campos fixos são pré-preenchidos e os campos
   variáveis ficam em branco para preenchimento.

### Edge Cases

- Cadastro de cliente/serviço com dados obrigatórios faltando: rejeitado com
  indicação dos campos pendentes. Obrigatórios mínimos — cliente: razão social/
  nome + documento (CNPJ/CPF); serviço: descrição + valor unitário +
  classificação fiscal.
- Dois clientes com o mesmo documento (CNPJ/CPF) na mesma empresa: o sistema
  **avisa a duplicidade, mas permite** o cadastro.
- Salvar template com nome já existente na empresa: o sistema **bloqueia**
  (nome de template é único por empresa).
- Remoção de um cliente/serviço referenciado por um template existente: definir
  comportamento (impedir, ou manter o template com snapshot dos dados).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Usuários com permissão de edição MUST poder criar, listar, editar
  e remover clientes da sua empresa.
- **FR-002**: Usuários com permissão de edição MUST poder criar, listar, editar
  e remover produtos/serviços (catálogo) da sua empresa.
- **FR-003**: O sistema MUST associar obrigatoriamente todo cliente, serviço e
  template à empresa do usuário autenticado.
- **FR-004**: O sistema MUST impedir que registros de uma empresa sejam vistos
  ou manipulados por usuários de outra empresa.
- **FR-005**: O sistema MUST restringir criação/edição/remoção a papéis com
  permissão de edição; papéis somente leitura têm acesso apenas de consulta.
- **FR-006**: O sistema MUST permitir salvar a estrutura de uma nota como
  template nomeado, preservando itens padrão, retenções padrão e tributação
  padrão.
- **FR-007**: Ao salvar um template, o sistema MUST omitir (não persistir) os
  campos variáveis cliente e valor.
- **FR-008**: O sistema MUST permitir iniciar uma emissão a partir de um
  template, pré-preenchendo os campos fixos.
- **FR-009**: O catálogo de serviços MUST armazenar os dados necessários para
  acelerar a emissão (ex.: descrição, valor unitário, unidade de medida,
  classificação fiscal do serviço).

### Key Entities *(include if feature involves data)*

- **Cliente**: tomador do serviço, pertencente a uma empresa. Atributos: razão
  social/nome, documento (CNPJ/CPF), inscrição municipal, endereço, e-mail.
- **Produto/Serviço**: item do catálogo da empresa. Atributos: descrição, valor
  unitário, unidade de medida, código de classificação fiscal (ex.: NCM/CNAE).
- **Template de Nota**: modelo reutilizável de nota da empresa. Atributos: nome,
  itens padrão, retenções padrão, código de tributação padrão. Não contém
  cliente nem valor.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Um usuário consegue cadastrar um novo cliente ou serviço em menos
  de 1 minuto.
- **SC-002**: 100% dos registros criados ficam corretamente associados à empresa
  do autor; 0% visíveis por outra empresa.
- **SC-003**: Iniciar uma nota a partir de um template reduz o número de campos
  preenchidos manualmente em pelo menos 50% frente à emissão do zero.
- **SC-004**: 0% dos templates salvos contêm dados de cliente ou valor.

## Assumptions

- A feature de emissão (Epic 3) consome estes cadastros; aqui só tratamos
  criação/manutenção dos registros base e do template.
- A classificação fiscal (NCM/CNAE) é informada/validada conforme regras a
  definir no plano.
- Papéis e permissões seguem o modelo definido na feature 001.
