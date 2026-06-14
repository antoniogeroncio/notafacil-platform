# Contract — Auth & Team API (`/api/v1`)

Contrato REST da feature 001. Erros seguem `{ "code": string, "message": string }`
(pt-BR na `message`), com HTTP semântico (Princípio XII). Sessão via cookie
httpOnly. `tenantId` e `role` derivam do token — nunca do corpo (Princípio III).

> Backend-First (Princípio XIV): este contrato é derivado do backend real; o
> frontend só consome o que estiver aqui implementado.

## Autenticação

### POST /api/v1/auth/login
Autentica um usuário `Ativo`.
- **Body**: `{ "email": string, "senha": string }`
- **200**: define cookie de sessão; retorna `{ "user": UserView }`
- **401** `invalid_credentials`: e-mail/senha inválidos ou usuário não `Ativo`.

### POST /api/v1/auth/logout
- **204**: limpa o cookie de sessão.

### GET /api/v1/me
- **200**: `{ "user": UserView }` (usuário do contexto autenticado)
- **401** `unauthenticated`.

## Convites (equipe)

### POST /api/v1/invites  *(role: Admin)*
Convida um membro para a empresa do Admin.
- **Body**: `{ "email": string, "role": "Admin"|"Editor"|"Viewer" }`
- **201**: `{ "invite": InviteView }` — cria `user` `Pendente`, gera token (48h),
  dispara e-mail de ativação.
- **403** `forbidden`: papel sem permissão (Editor/Viewer).
- **409** `email_conflict`: e-mail já existe na plataforma.
- **422** `validation_error`: e-mail/role inválidos.

### POST /api/v1/invites/{token}/accept
Ativa a conta a partir do link de convite. **Público** (sem sessão).
- **Body**: `{ "nome": string, "senha": string }`
- **200**: ativa o usuário (`status=Ativo`), marca convite como usado, autentica
  (define cookie); retorna `{ "user": UserView }`.
- **410** `invite_expired`: token expirado ou já usado.
- **404** `invite_not_found`: token inválido/adulterado.
- **422** `weak_password`: senha não atende à política mínima.

### GET /api/v1/users  *(role: Admin)*
Lista os membros da empresa do contexto (tenant-scoped).
- **200**: `{ "users": UserView[] }` — apenas usuários do mesmo tenant.
- **403** `forbidden`.

## Schemas

```jsonc
// UserView (nunca inclui senhaHash nem tokens)
{ "id": string, "nome": string, "email": string,
  "role": "Admin"|"Editor"|"Viewer", "status": "Pendente"|"Ativo" }

// InviteView (nunca inclui o token em texto puro)
{ "id": string, "email": string, "role": string, "expiresAt": string }
```

## Regras transversais (testáveis)

- Toda rota autenticada injeta `tenantId`/`role` do token; valor de empresa no
  corpo é ignorado.
- Acesso a recurso de outro tenant → **404** (não vaza existência) — SC-002.
- `senhaHash`, tokens de convite e qualquer segredo nunca aparecem em respostas
  ou logs — SC-005.
- Convite expira em 48h (SC-003); e-mail é único global (SC-004).
