# Feature Specification: Motor de Emissão Híbrida e Flexível

**Feature Branch**: `003-emissao-hibrida`

**Created**: 2026-06-14

**Status**: Draft

**Input**: Epic 3 — A interface deve suportar tanto a emissão rápida via modelos quanto a digitação livre dos itens da nota.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Adicionar itens por catálogo ou por digitação livre (Priority: P1)

Ao montar uma nota, o usuário digita no campo de item e recebe sugestões do
catálogo da sua empresa. Se selecionar um item do catálogo, os dados (descrição,
classificação fiscal e valor) são preenchidos automaticamente. Se preferir,
digita um item totalmente livre, que é aceito apenas para aquela nota, sem ser
forçado a salvá-lo no catálogo.

**Why this priority**: É o coração da emissão e a proposta de valor "híbrida":
rapidez via catálogo sem perder a flexibilidade da entrada livre. Sem isso não
há emissão.

**Independent Test**: Pode ser testado montando uma nota com (a) um item
selecionado do catálogo, verificando o auto-preenchimento, e (b) um item digitado
livremente, verificando que é aceito sem criar registro no catálogo.

**Acceptance Scenarios**:

1. **Given** um usuário no grid de itens de uma nota, **When** ele digita no
   campo de item, **Then** o sistema sugere produtos/serviços existentes no
   catálogo da empresa que correspondem ao texto.
2. **Given** sugestões exibidas, **When** o usuário seleciona um item do
   catálogo, **Then** descrição, classificação fiscal e valor unitário são
   preenchidos automaticamente a partir do cadastro.
3. **Given** que nenhuma sugestão corresponde ao desejado, **When** o usuário
   preenche os campos do item manualmente e não seleciona nada do catálogo,
   **Then** o sistema aceita o item como entrada livre apenas para aquela nota,
   sem criar nem alterar o catálogo.

---

### User Story 2 - Montar a nota e calcular totais (Priority: P1)

O usuário compõe a nota com um ou mais itens (de catálogo e/ou livres), informa o
cliente e quantidades, e o sistema mantém os totais da nota consistentes em tempo
real para revisão antes do faturamento.

**Why this priority**: Uma nota só tem utilidade quando representa corretamente
itens, cliente e valores totais. Complementa US1 para formar uma nota completa e
pronta para transmissão (Epic 4).

**Independent Test**: Pode ser testado adicionando vários itens com quantidades e
valores e verificando que o total exibido corresponde à soma esperada, e que
selecionar um cliente associa-o à nota.

**Acceptance Scenarios**:

1. **Given** uma nota com itens e quantidades, **When** o usuário altera
   quantidade ou valor de um item, **Then** os totais da nota são recalculados e
   refletem a mudança imediatamente.
2. **Given** uma nota em edição, **When** o usuário seleciona um cliente da sua
   empresa, **Then** o cliente é vinculado à nota.
3. **Given** uma nota sem itens ou sem cliente, **When** o usuário tenta avançar
   para faturamento, **Then** o sistema impede e indica os campos obrigatórios
   faltantes.

### Edge Cases

- Item livre com classificação fiscal ausente ou inválida: [NEEDS CLARIFICATION:
  bloquear avanço ou permitir e validar só no faturamento?]
- Catálogo vazio: a digitação livre deve funcionar normalmente.
- Valor zero ou negativo em um item: rejeitado.
- Edição de uma nota iniciada a partir de um template (Epic 2): itens fixos vêm
  pré-carregados e permanecem editáveis.
- Item do catálogo é alterado/removido depois de adicionado à nota: a nota usa o
  valor capturado no momento da adição (snapshot), não reflete mudanças
  posteriores do catálogo.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: O sistema MUST oferecer, no campo de item, sugestões do catálogo da
  empresa correspondentes ao texto digitado.
- **FR-002**: Ao selecionar um item do catálogo, o sistema MUST preencher
  automaticamente descrição, classificação fiscal e valor unitário.
- **FR-003**: O sistema MUST aceitar itens de entrada livre (digitados
  manualmente) válidos apenas para a nota corrente.
- **FR-004**: O sistema MUST NOT criar ou alterar registros do catálogo a partir
  de itens de entrada livre, salvo ação explícita do usuário.
- **FR-005**: O sistema MUST permitir compor uma nota com múltiplos itens,
  mistos (catálogo e livres), com quantidade por item.
- **FR-006**: O sistema MUST recalcular os totais da nota sempre que itens,
  quantidades ou valores mudarem.
- **FR-007**: O sistema MUST permitir vincular um cliente da empresa à nota.
- **FR-008**: O sistema MUST validar a nota antes do faturamento, exigindo ao
  menos um item válido e um cliente.
- **FR-009**: O sistema MUST permitir iniciar a emissão a partir de um template
  (Epic 2), com os campos fixos pré-carregados e editáveis.
- **FR-010**: Itens adicionados de catálogo MUST registrar um snapshot dos dados
  no momento da adição, independente de alterações futuras no catálogo.

### Key Entities *(include if feature involves data)*

- **Nota (em edição/rascunho)**: documento fiscal sendo montado, pertencente a
  uma empresa. Atributos: cliente vinculado, lista de itens, totais. Precede a
  transmissão tratada na Epic 4.
- **Item da Nota**: linha da nota. Atributos: descrição, classificação fiscal,
  valor unitário, quantidade, origem (catálogo ou livre). Snapshot quando vindo
  do catálogo.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Selecionar um item do catálogo preenche todos os seus campos sem
  digitação adicional pelo usuário.
- **SC-002**: 100% dos itens de entrada livre são aceitos sem criar registros no
  catálogo.
- **SC-003**: Os totais exibidos correspondem à soma correta dos itens em 100%
  dos casos testados.
- **SC-004**: O usuário consegue emitir uma nota recorrente (via template) em
  menos da metade do tempo de uma emissão do zero.

## Assumptions

- O catálogo e os clientes provêm da feature 002.
- Esta feature cobre a montagem/validação da nota; a emissão/transmissão fiscal
  (via provedor Focus NFe) é tratada na feature 004.
- O avanço para faturamento está sujeito à assinatura ativa e à política de cota
  do plano (feature 005): sem plano ativo, bloqueia; acima da franquia, segue em
  modo pagamento por uso até o teto (+1.000), quando bloqueia até liberação
  manual do administrador do sistema. A orientação é exibida antes de enviar ao
  provedor.
- As regras de cálculo de impostos/retenções aplicáveis serão detalhadas no
  plano e podem depender da configuração fiscal da empresa.
