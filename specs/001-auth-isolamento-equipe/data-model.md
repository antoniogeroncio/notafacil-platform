# Data Model — 001 Autenticação, Isolamento e Gestão de Equipe

MongoDB (single-database, multi-tenant). Toda coleção com dados de tenant inclui
`tenantId` e um índice composto que começa por `tenantId` (Princípio III).

## Coleção `tenants`

Raiz do isolamento. Não carrega `tenantId` (é o próprio tenant).

| Campo | Tipo | Notas |
|-------|------|-------|
| `_id` | ObjectId | identidade do tenant |
| `razaoSocial` | string | obrigatório |
| `cnpj` | string | obrigatório, **único na plataforma** |
| `criadoEm` | datetime | |

**Índices**: `{cnpj: 1}` unique.

## Coleção `users`

| Campo | Tipo | Notas |
|-------|------|-------|
| `_id` | ObjectId | |
| `tenantId` | ObjectId | obrigatório; FK → tenants |
| `nome` | string | preenchido na ativação |
| `email` | string | **único global** (identidade de login) |
| `senhaHash` | string | bcrypt; ausente enquanto `Pendente` |
| `role` | enum | `Admin` \| `Editor` \| `Viewer` |
| `status` | enum | `Pendente` \| `Ativo` |
| `criadoEm` / `ativadoEm` | datetime | |

**Índices**: `{email: 1}` unique (global, FR-004); `{tenantId: 1, _id: 1}`
(isolamento/performance); `{tenantId: 1, role: 1}`.

**Regras de validação**:
- `email` válido e único globalmente (409 em conflito).
- `senhaHash` nunca retornado pela API (Princípio VI).
- `role` ∈ enum; `status` inicia `Pendente` (convite) ou `Ativo` (signup 005).

**Transições de estado** (`status`):
```
Pendente --(ativação via convite válido + senha)--> Ativo
```
Ativação só a partir de convite válido (não expirado, não usado).

## Coleção `invites`

| Campo | Tipo | Notas |
|-------|------|-------|
| `_id` | ObjectId | |
| `tenantId` | ObjectId | empresa do convite |
| `email` | string | convidado |
| `role` | enum | papel atribuído |
| `tokenHash` | string | SHA-256 do token (segredo nunca em texto puro) |
| `expiresAt` | datetime | criação + 48h |
| `usedAt` | datetime\|null | preenchido na ativação |
| `invitedByUserId` | ObjectId | autor (Admin) |

**Índices**: `{tenantId: 1, email: 1}`; `{tokenHash: 1}` unique;
`{expiresAt: 1}` (TTL opcional para limpeza).

**Estados (derivados)**:
```
válido      = usedAt == null AND now < expiresAt
expirado    = now >= expiresAt
usado       = usedAt != null
```

**Regras**:
- Reenvio para `Pendente`: reemite token (novo `tokenHash`/`expiresAt`) sem criar
  segundo usuário.
- Ativação: exige convite `válido`; marca `usedAt` e ativa o `user`.

## Relacionamentos

```
Tenant 1───* User        (User.tenantId)
Tenant 1───* Invite      (Invite.tenantId)
Invite *───1 User(email) (mesmo email/tenant; vínculo lógico)
```

## Invariantes de isolamento (testáveis — Princípio III)

- Toda leitura/escrita em `users`/`invites` é filtrada por `tenantId` do contexto.
- Buscar um `_id` de outro tenant retorna **not-found** (sem vazar existência).
- `tenantId` jamais provém do payload do cliente.
