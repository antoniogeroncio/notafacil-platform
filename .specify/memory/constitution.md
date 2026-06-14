# Constituição de Engenharia — NotaFácil Platform

**Versão**: 2.0.1 | **Ratificada**: 2026-06-14 | **Última Emenda**: 2026-06-14

> Este documento define os **princípios inegociáveis de engenharia de software** que regem todo o desenvolvimento da plataforma multi-tenant de emissão de NFS-e. Os princípios são **agnósticos de tecnologia**: aplicam-se independente de linguagem, framework ou plataforma. As escolhas tecnológicas concretas que implementam estes princípios estão em [`tech-stack.md`](tech-stack.md).
>
> Estrutura adaptada da Constituição de Engenharia Care Platform (v3.2.0). As lacunas de numeração (VII, IX, XIII) correspondem a princípios da matriz original ainda não adotados neste projeto.

---

## Procedimento de Emenda

- **MAJOR**: remoção ou redefinição incompatível de princípio — requer aprovação da equipe.
- **MINOR**: novo princípio ou expansão material de um existente — requer ao menos um review de aprovação.
- **PATCH**: esclarecimentos, correções de redação — requer ao menos um review.

Toda emenda deve documentar: o princípio alterado, a justificativa e o plano de migração para código existente em violação.

---

## Princípio I — Arquitetura em Camadas (NÃO NEGOCIÁVEL)

Todo código de produto do backend deve ser organizado em **três camadas isoladas** com dependência unidirecional:

```
Controller  →  Service  →  Repository
```

- **Controller**: ponto de entrada HTTP. Responsável por receber requisições, validar inputs, aplicar guards de autenticação/autorização e serializar respostas. Não contém lógica de negócio. Mapeia dados de entrada em **DTOs** que fluem para dentro.
- **Service**: contém todas as regras de negócio. Orquestra operações, aplica validações de domínio e coordena repositórios. Não tem conhecimento de HTTP, frameworks web ou detalhes de persistência.
- **Repository**: camada de acesso a dados. Abstrai operações de persistência (queries de banco, chamadas a APIs externas). Services dependem de **interfaces** de repositório — nunca de implementações concretas.

**DTOs** são definidos na camada Controller e fluem para dentro (Controller → Service → Repository). Repositories traduzem DTOs para/de modelos de persistência.

**Injeção de Dependência** é obrigatória para desacoplar todas as camadas. Importações que violem a hierarquia de camadas são falha bloqueante de code review.

**Integrações externas** (provedor de emissão fiscal, gateway de pagamento, mensageria, e-mail) são encapsuladas em classes de infraestrutura injetadas como dependência do Service. Nenhum SDK externo deve ser instanciado diretamente dentro de um Service ou Controller. O motor de emissão depende de uma **abstração de provedor de emissão** (v1: Focus NFe) e o de cobrança, de uma abstração de gateway (v1: PagSeguro), permitindo trocar o provedor sem alterar o motor (ver [`tech-stack.md`](tech-stack.md)).

---

## Princípio II — Clean Code

- **Responsabilidade Única (SRP)**: toda função, método ou módulo tem uma responsabilidade claramente definida. Funções com mais de ~40 linhas devem ser decompostas, salvo justificativa algorítmica documentada.
- **DRY**: duplicação de lógica de negócio é proibida. Abstrações compartilhadas vivem em camadas ou pacotes comuns.
- **Nomenclatura expressiva**: nomes de variáveis, funções, classes e módulos devem revelar intenção, não implementação.
- **Código em inglês**: todo código-fonte, comentários e nomes de identificadores devem ser escritos em inglês. Strings voltadas ao usuário são em pt-BR.
- **Caminhos de erro explícitos**: todo caminho de erro deve ser tratado. Exceções silenciosas são proibidas. Mensagens de erro ao usuário não expõem stack traces, identificadores técnicos nem segredos.

---

## Princípio III — Isolamento Multi-Tenant (NÃO NEGOCIÁVEL)

Nenhuma requisição, query, job ou log pode vazar dados de um tenant (empresa) para outro. Este é o defeito mais grave possível e bloqueia qualquer release.

- Toda entidade de domínio pertence a exatamente um tenant e carrega seu identificador.
- Todo acesso a dados deve ser filtrado pelo tenant do contexto autenticado, **de forma central e por padrão** — nunca dependendo de o desenvolvedor lembrar de adicionar o filtro em cada query (ver mecanismo em [`tech-stack.md`](tech-stack.md)).
- O identificador de tenant é derivado **exclusivamente da identidade autenticada** (token de sessão), nunca de input do cliente (body, query string, header arbitrário). Qualquer tenant informado pelo cliente é ignorado.
- Tentativa de acesso a recurso de outro tenant responde como se o recurso não existisse — sem vazar existência ou conteúdo.
- O isolamento é uma regra crítica sob TDD (Princípio V): exige testes que cobrem o cenário de **violação** (acesso cross-tenant deve falhar).

---

## Princípio IV — Estratégia de Testes

A plataforma adota dois níveis de teste: **unidade** e **integração**. Nenhum outro nível é obrigatório (exceto golden-path, Princípio VIII).

### Testes de Unidade
- Cobrem a camada **Service** (regras de negócio).
- São **isolados**: sem dependência de rede, banco de dados, filesystem ou serviços externos.
- Dependências externas (repositórios, integrações) são substituídas por **mocks/fakes** controlados pelo teste.
- Todo novo Service deve ter cobertura de unidade — ausência bloqueia merge.

### Testes de Integração
- Cobrem fluxos com I/O real (banco de dados, endpoints HTTP via Controller).
- Usam infraestrutura efêmera (containers) criada e destruída pelo próprio teste.
- Não dependem de ambiente externo fixo.
- Cobrem os caminhos críticos: **isolamento multi-tenant**, autenticação/autorização por papel, emissão e transmissão fiscal.

### Gates de CI
- PR com novo Service sem teste de unidade → **bloqueado**.
- PR que toca no **isolamento multi-tenant** sem teste de violação cross-tenant → **bloqueado**.
- PR que toca em **credenciais ou dados fiscais** sem demonstrar os controles de segurança testados (Princípio VI) → **bloqueado**.
- PR que toca no **motor de emissão fiscal** sem teste dos caminhos de assinatura/transmissão e de falha → **bloqueado**.
- Qualquer falha de teste bloqueia merge para a branch principal.

---

## Princípio V — TDD (Test-Driven Development — NÃO NEGOCIÁVEL)

Nenhuma linha de código de produção é escrita antes de existir um teste que falha por essa ausência.

O ciclo obrigatório é:

```
🔴 Red    → escreve o teste que descreve o comportamento esperado (falha)
🟢 Green  → escreve o mínimo de código para o teste passar
🔵 Refactor → limpa o código sem quebrar o teste
```

**Aplicação por camada**:

- **Services**: teste de unidade escrito antes do service. Dependências externas (repositórios, integrações) mockadas.
- **Controllers / Repositories**: teste de integração escrito antes da implementação. Usa container efêmero.

**Regras inegociáveis**:

- Código de produção sem teste precedente é **não rastreável** — proibido em merge.
- O teste deve falhar pela razão correta antes de ser feito passar — teste que passa sem implementação é inválido.
- Testes não testam implementação interna; testam **comportamento observável**.
- TDD não substitui a cobertura — substitui a ordem: teste vem primeiro, sempre.

---

## Princípio VI — Segurança de Credenciais e Dados Fiscais (NÃO NEGOCIÁVEL)

Certificados digitais (A1), senhas de certificado, credenciais de API de terceiros e senhas de usuário são segredos.

- Devem ser armazenados **criptografados em repouso**, descriptografados apenas em memória no momento do uso.
- **Nunca** são expostos em respostas de API, logs, mensagens de erro ou telemetria.
- Senhas de usuário e de certificado são sempre tratadas via hash/criptografia apropriada — nunca em texto puro.
- O acesso a esses segredos é restrito ao tenant proprietário (Princípio III).
- Manuseio de segredos é regra crítica sob TDD (Princípio V): exige teste cobrindo o cenário de violação (ex.: segredo não retornado em texto puro; senha de certificado inválida/expirada bloqueia a transmissão).

---

## Princípio VIII — Testes de Golden-Path

Além dos testes de unidade e integração (Princípio IV), toda feature que introduza um formulário de ação acessível por URL direta (incluindo deep links e redirects pós-autenticação) deve ser coberta por ao menos um **teste de golden-path** antes do merge. Aplica-se a fluxos críticos do cliente web — ex.: emissão de nota, configuração fiscal do tenant, ativação de convite.

### O que um teste de golden-path cobre

1. Inicia com **sessão limpa** — sem estado em memória carregado por navegação anterior.
2. Simula a **interação completa do usuário** desde a abertura da tela até a conclusão da ação (preenchimento de formulário, submissão, verificação do redirecionamento resultante).
3. Executa no **runtime real** do cliente (browser).
4. Isola chamadas a serviços externos com **respostas controladas** — não requer backend em execução.
5. Injeta o estado de sessão (token, cookie) diretamente no contexto do cliente, independentemente do fluxo de login.

### O que teste de golden-path NÃO é

- Não substitui testes de unidade de regras de negócio.
- Não é teste de regressão visual.
- Não é teste de carga ou performance.

### Gate

- PR que introduz formulário de ação sem teste de golden-path correspondente → **bloqueado**.
- Testes de golden-path devem cobrir: happy path, estado de sessão fresh (sem memória), e ao menos um caminho de erro visível ao usuário.

---

## Princípio X — Infraestrutura como Código

Todo ambiente de desenvolvimento deve ser reprodutível com um único comando, a partir de um checkout limpo, sem instalação manual de dependências no host.

- Configurações de infraestrutura vivem no repositório.
- Containers são o padrão para serviços com estado (banco de dados, mensageria).
- O pipeline de CI deve executar sobre a mesma definição de infraestrutura usada em desenvolvimento.

---

## Princípio XI — Contratos de API como Fonte de Verdade

Contratos de API (endpoints, schemas de request/response, códigos de status) são documentados explicitamente e mantidos em sincronia com a implementação.

- Toda discrepância entre contrato documentado e implementação real é um bug de documentação.
- O contrato reflete o comportamento **atual** do backend, não o comportamento desejado.
- Mudanças em handlers de API devem ser acompanhadas, no mesmo PR, da atualização do contrato correspondente.

---

## Princípio XII — Design de API

- **Versionamento**: endpoints REST são prefixados com versão (ex.: `/api/v1/`). Mudanças incompatíveis requerem novo prefixo.
- **Erros estruturados**: respostas de erro retornam JSON com ao menos `code` e `message`. Códigos HTTP são semanticamente corretos (4xx para erro do cliente, 5xx para erro do servidor; ex.: 409 para duplicidade, 401/403 para auth/autorização).
- **Agregação no servidor**: cálculos de agregação (totais de nota, financeiro, estatísticas) são computados pela API. Clientes recebem valores pré-agregados.
- **Autenticação centralizada**: a autenticação é centralizada no backend, que emite o próprio token de sessão (carregando tenant e papel). Clientes não usam tokens de provedores de identidade externos diretamente nas chamadas de API.

---

## Princípio XIV — Backend-First (NÃO NEGOCIÁVEL)

O backend é **sempre implementado antes** do cliente web. Nenhum cliente pode ser desenvolvido com base em suposições sobre o contrato da API.

O fluxo obrigatório por história é:

```
1. Spec aprovada (specs/NNN-*/spec.md)
2. Backend implementa o endpoint
3. Contrato extraído do backend real → specs/NNN-*/contracts/
4. Cliente (frontend) implementa consumindo o contrato documentado — sem desvios
```

**Regras inegociáveis**:

- Clientes **não antecipam** endpoints não implementados. Nenhuma chamada de API é escrita antes do handler existir no backend.
- O contrato documentado em `specs/` é **derivado do código-fonte real** do backend — nunca de suposição, memória ou documentação desatualizada.
- Qualquer divergência entre o comportamento real do backend e o contrato documentado é um **bug de documentação** — deve ser corrigido no mesmo PR.
- Mocks e stubs de desenvolvimento devem ser **fiéis ao contrato documentado**. Mock que inventa campos ou comportamentos não existentes no backend é proibido.

---

## Princípio XV — Arquitetura Frontend (MVVM)

Todo código de frontend web deve ser organizado em **três camadas isoladas** com dependência unidirecional:

```
Component  →  Hook / Store  →  Repository
```

- **Component** (View): responsável exclusivamente por renderizar a interface e capturar eventos do usuário. Não contém lógica de negócio nem chama APIs diretamente. Recebe dados e callbacks como props ou via hook.
- **Hook / Store** (ViewModel): gerencia estado, dados derivados e orquestração de operações. Conecta o Component ao Repository. Contém as regras de apresentação — formatação, validação de UI, lógica condicional de exibição (ex.: autocomplete de catálogo no grid de itens da nota).
- **Repository** (API Service): responsável por todas as chamadas HTTP, desserialização de respostas e mapeamento para DTOs tipados. Não tem conhecimento de estado de UI.

**DTOs** são definidos na borda do Repository (onde a resposta da API chega) e fluem para dentro. Components e Hooks/Stores trabalham exclusivamente com DTOs tipados — nunca com JSON bruto.

**Regras inegociáveis**:

- Components **não importam** serviços de API diretamente. Toda chamada de rede passa pelo Hook/Store.
- Hooks/Stores **não constroem** requests HTTP diretamente. Toda lógica de rede fica no Repository.
- Estado global de aplicação (autenticação, dados compartilhados entre rotas) vive em Stores. Estado local de UI vive no próprio Component ou em hooks locais.
- Testes de unidade cobrem Hooks/Stores com o Repository mockado. Testes de integração cobrem o Repository com a API real (ou um servidor mock fiel ao contrato).
