# Managed-model upstream API keys (Sigmaze fork)

This fork adds `models.<model>.upstreamApiKeyEnv`. It names an environment
variable in the LLMSnap process; **it is not the secret value**.

```yaml
apiKeys:
  - "${env.API_KEY_1}"   # inbound LiteLLM -> LLMSnap
models:
  bielik-lite:
    cmd: your-backend-lifecycle-command
    proxy: http://127.0.0.1:8101
    upstreamApiKeyEnv: BIELIK_API_KEY
```

Set `BIELIK_API_KEY` in the LLMSnap container environment, outside version
control. When this property is present, LLMSnap rejects startup if the
environment variable is empty or missing. It injects
`Authorization: Bearer <backend key>` for requests routed to that managed
model and never forwards incoming `Authorization` or `x-api-key` headers.
Models without `upstreamApiKeyEnv` remain supported for backends with no
HTTP API-key requirement.

Health checks follow the existing LLMSnap behavior and do not inject API
keys; use an unauthenticated readiness endpoint such as vLLM's `/health`.
Do not use this feature to forward the inbound `API_KEY_1` credential.

For reproducible tests from a fresh checkout, first build the embedded UI and
test helper (including the Linux ARM64 helper used on GX10), then run focused
and full Go tests:

```sh
npm --prefix ui-svelte ci --ignore-scripts
npm --prefix ui-svelte run build
make simple-responder
go test ./proxy/... -run 'ManagedModelUpstream|UpstreamAPIKeyEnv|APIKeyAuth' -count=1
go test ./...
```
