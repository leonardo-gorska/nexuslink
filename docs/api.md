# NexusLink — Referência da API (REST)

Esta rotulagem reflete toda a interface HTTP/REST controlada pela aplicação do Gateway de Serviço Core (API). Todos os recursos seguem os padrões semânticos de verbos HTTP e códigos de erro formatados via **Problem Details RFC 7807**.

---

## 1. Tabela de Endpoints

| Método | Path | Descrição | Rate Limit | Auth |
|--------|------|-----------|:----------:|:----:|
| `GET` | `/r/{hash}` | Redirecionamento **(Hot Path)** | 1000/min/IP | ❌ |
| `POST` | `/api/v1/links` | Criar Link Encurtado | 10/min/IP | ❌ |
| `GET` | `/api/v1/links/{hash}` | Consulta de Metadata Individual | 60/min/IP | ❌ |
| `GET` | `/api/v1/links/{hash}/stats` | Análise e Estatísticas agregadas | 30/min/IP | ❌ |
| `DELETE`| `/api/v1/links/{hash}` | Soft Delete do Hash | 10/min/IP | ❌ |
| `GET` | `/healthz` | Probe de Liveness Node | — | — |
| `GET` | `/readyz` | Probe de Readiness (Dependencies) | — | — |
| `GET` | `*:9090/metrics` | Scraper Prometheus Custom | — | — |

---

## 2. Padrão de Erros (RFC 7807)

Se surgir uma ocorrência atípica ou bloqueante, a API retornará o contrato JSON em conformidade com o IETF RFC 7807 estruturado com `Content-Type: application/problem+json`:

```json
{
  "type": "https://nexuslink.dev/errors/rate-limited",
  "title": "Too Many Requests",
  "status": 429,
  "detail": "Rate limit exceeded. Try again in 48 seconds.",
  "instance": "/api/v1/links",
  "request_id": "req_8bdf887e"
}
```

**Tipos base que cobrimos:**
- `/invalid-request` (`400 Bad Request`): Erros de Parse JSON e URLs inválidas.
- `/link-not-found` (`404 Not Found`): Quando um Hash procurado foi descartado ou não é autêntico.
- `/link-expired` (`410 Gone`): Se o TTL (Time-To-Live) temporal exauriu.
- `/rate-limited` (`429 Too Many Requests`): Quando os limites atrelados ao cliente IP cruzarem os vetores estabelecidos pelo Redis.
- `/internal-error` (`500 Internal Server Error`): Queda crítica de banco/MQ no processo do uso ou Panic não contido em loop.

---

## 3. Guia Detalhado e Payload Exemplos

### 3.1. Encurtamento de URL

Gera um hash Base62 que resolverá permanentemente para a mesma origem definida na string `url`.

> `POST /api/v1/links`

**Request:**
```json
{
  "url": "https://github.com/leonardo-gorska/nexuslink",
  "expires_in": "720h"
}
```

**Response (201 Created):**
```json
{
  "hash": "aB3xK9z",
  "short_url": "http://localhost:8080/r/aB3xK9z",
  "original_url": "https://github.com/leonardo-gorska/nexuslink",
  "created_at": "2026-04-15T22:30:00Z",
  "expires_at": "2026-05-15T22:30:00Z",
  "is_active": true
}
```

### 3.2. Redirecionar URL

Executado nos requests do usuário final gerando um salto HTTP para outro ponto (Hot Path)

> `GET /r/{hash}`

**(Nenhum Payload)**

**Response (301 Moved Permanently):**
- Headers definidos obrigatoriamente:
- `Location: https://github.com/leonardo-gorska/nexuslink`
- `Cache-Control: private, max-age=90`

*Este processo também gera de forma invisível via Goroutines os Eventos AMQP de Clique.*

### 3.3. Retorno de Metadata (Informação Central)

> `GET /api/v1/links/{hash}`

**(Nenhum Payload)**

**Response (200 OK):**
```json
{
  "hash": "aB3xK9z",
  "short_url": "http://localhost:8080/r/aB3xK9z",
  "original_url": "https://github.com/leonardo-gorska/nexuslink",
  "created_at": "2026-04-15T22:30:00Z",
  "expires_at": "2026-05-15T22:30:00Z",
  "total_clicks_unaggregated": 1500,
  "is_active": true
}
```

### 3.4. Motor de Relatórios Analíticos 

Oferece os dados unificados advindos dos Analytics Workers e tabela de partições, cruzados com as definições geográficas, dispositivos e dias contínuos.

> `GET /api/v1/links/{hash}/stats?from=2026-04-10&to=2026-04-15`

**(Nenhum Payload)**

**Response (200 OK):**
```json
{
  "hash": "aB3xK9z",
  "total_clicks": 1847,
  "unique_visitors": 1203,
  "period": {
    "from": "2026-04-10",
    "to": "2026-04-15"
  },
  "by_country": [
    { "country": "BR", "clicks": 1200, "percentage": 64.9 },
    { "country": "US", "clicks": 347, "percentage": 18.8 },
    { "country": "PT", "clicks": 150, "percentage": 8.1 }
  ],
  "by_device": [
    { "device": "mobile", "clicks": 1100, "percentage": 59.6 },
    { "device": "desktop", "clicks": 700, "percentage": 37.9 },
    { "device": "tablet", "clicks": 47, "percentage": 2.5 }
  ],
  "by_browser": [
    { "browser": "Chrome", "clicks": 900, "percentage": 48.7 },
    { "browser": "Safari", "clicks": 500, "percentage": 27.1 }
  ],
  "by_day": [
    { "date": "2026-04-15", "clicks": 523 },
    { "date": "2026-04-14", "clicks": 412 },
    { "date": "2026-04-13", "clicks": 380 }
  ]
}
```

### 3.5. Exclusão (Soft Delete)

Inabilita a resolução de redirect subsequente invalidando a key dentro do Cache e configurando Flag na DB (para a regra de integridade analítica, dados não são mortos do BD).

> `DELETE /api/v1/links/{hash}`

**(Nenhum Payload)**

**Response (204 No Content)**
- (Null Body)
