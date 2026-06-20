# messaging-gateway (SVC-01)

Receiver de webhooks de WhatsApp (vía Kapso) del proyecto **whatsapp-agent**.
Verifica la firma del webhook, normaliza el mensaje y lo **encola en River** para
que el `agent-runtime` (SVC-02) lo procese async. Responde 200 al BSP recién
después de encolar (no perder mensajes).

- Épica: **E03** (Hito 1 — esqueleto end-to-end). Gobernanza: `../../management`.
- Arquitectura: hexagonal (P-04). Logs: canonical / envelope ADR-001 (P-20) vía `go-shared`.
- Lenguaje: Go ([[ADR-0001]]). Cola: River ([[ADR-0002]] no; STACK-04).

## Estructura

```
src/
├── main.go                 # wiring + http server + shutdown
├── config/                 # env (12-factor)
├── logging/                # wrapper de go-shared canonical logger
├── contract/               # InboundMessageArgs — CONTRATO cross-repo con agent-runtime (D-03)
└── inbound/                # contexto "ingreso de mensajes"
    ├── domain.go           # InboundMessage (VO) + puerto MessageQueue
    ├── service.go          # caso de uso: verificar → encolar
    ├── kapso.go            # infra: verificación de firma + parseo del webhook
    ├── riverqueue.go       # infra: adaptador River (productor)
    ├── logqueue.go         # infra: adaptador dev (solo loguea)
    └── http.go             # infra: handlers /health y /api/v1/webhooks/kapso
```

## Endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/health` | liveness |
| POST | `/api/v1/webhooks/kapso` | webhook entrante (verifica firma, encola, 200) |

## Correr local

Requiere el lab y la base `whatsapp_agent` (`cd ../.. && make -C ~/Projects infra`; `cd services && make db`).

```bash
# Compilar y correr contra lab-postgres (desde el host: DB_HOST=localhost)
go build -o /tmp/wa-gateway ./src
DB_HOST=localhost DB_USER=whatsapp_agent DB_PASSWORD=whatsapp_agent DB_NAME=whatsapp_agent \
  QUEUE_DRIVER=river PORT=8101 /tmp/wa-gateway

# Probar
curl -X POST localhost:8101/api/v1/webhooks/kapso -H 'Content-Type: application/json' \
  -d '{"tenant_slug":"demo","message_id":"wamid.TEST","from":"5491155550000","to":"549110000","text":"Hola"}'
```

`QUEUE_DRIVER=log` corre sin tocar la cola (solo loguea el encolado).

> **Módulo privado:** `go-shared` requiere `GOPRIVATE=github.com/hornosg/*` y acceso por
> SSH/token. En Docker el token va como BuildKit secret:
> `GITHUB_TOKEN=$(gh auth token) docker compose build messaging-gateway`.

## TODOs (siguiente iteración de E03)

- [ ] Mapear la forma real del webhook de **Kapso** (`kapso.go`) y confirmar header/esquema de firma.
- [ ] **Outbox** para outbound (envío a Kapso API) + manejo de **ventana 24h** y templates Meta.
- [ ] Extraer `contract.InboundMessageArgs` a un paquete compartido con `agent-runtime` (D-03 / ADR-0003).
- [ ] Tenancy: resolver `tenant_slug` real (hoy viene del payload) contra `core.tenants`.
