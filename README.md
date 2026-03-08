# Balón de Oro API

API REST pura en Go (stdlib únicamente) para gestionar el historial de ganadores del **Balón de Oro**.

**Tema:** Ganadores del Balón de Oro (1956–presente)  
**Puerto:** `24725`  
**Tecnología:** Go 1.22 — sin frameworks externos

---

## Estructura del proyecto

```
.
├── main.go              # Entry point y routing
├── go.mod
├── Dockerfile
├── docker-compose.yml
├── data/
│   └── winners.json     # Persistencia en disco
├── handlers/
│   └── winners.go       # Lógica de cada endpoint
├── models/
│   └── winner.go        # Struct Winner
└── store/
    └── store.go         # CRUD thread-safe con persistencia
```

---

## Modelo de datos

```json
{
  "id":                1,
  "player":            "Lionel Messi",
  "nationality":       "Argentina",
  "club":              "FC Barcelona",
  "year":              2009,
  "votes":             473,
  "position":          "Forward",
  "goals_that_season": 38
}
```

Valores válidos para `position`: `Forward` | `Midfielder` | `Defender` | `Goalkeeper`

---

## Endpoints

| Método   | Ruta                    | Descripción                           |
|----------|-------------------------|---------------------------------------|
| `GET`    | `/api/ping`             | Health check                          |
| `GET`    | `/api/winners`          | Listar todos / filtrar con queries    |
| `GET`    | `/api/winners/{id}`     | Obtener ganador por path parameter    |
| `POST`   | `/api/winners`          | Registrar nuevo ganador               |
| `PUT`    | `/api/winners/{id}`     | Reemplazar registro completo          |
| `PATCH`  | `/api/winners/{id}`     | Actualizar campos parcialmente        |
| `DELETE` | `/api/winners/{id}`     | Eliminar registro                     |

---

## Query Parameters en `GET /api/winners`

| Parámetro     | Tipo   | Descripción                                    |
|---------------|--------|------------------------------------------------|
| `id`          | int    | Buscar por ID exacto                           |
| `player`      | string | Filtrar por nombre (parcial, case-insensitive) |
| `nationality` | string | Filtrar por nacionalidad                       |
| `club`        | string | Filtrar por club (parcial)                     |
| `year`        | int    | Filtrar por año de entrega                     |
| `position`    | string | Filtrar por posición                           |
| `min_goals`   | int    | Filtrar por mínimo de goles en esa temporada   |

Todos los filtros son combinables entre sí.

---

## Ejemplos de uso

### Health check
```bash
curl http://localhost:24725/api/ping
```

### GET — Todos los ganadores
```bash
curl http://localhost:24725/api/winners
```

### GET — Por query parameter
```bash
curl "http://localhost:24725/api/winners?id=5"
```

### GET — Por path parameter
```bash
curl http://localhost:24725/api/winners/5
```

### GET — Filtros combinados
```bash
curl "http://localhost:24725/api/winners?nationality=Argentina&position=Forward&min_goals=50"
curl "http://localhost:24725/api/winners?club=Real+Madrid"
```

### POST — Registrar nuevo ganador
```bash
curl -X POST http://localhost:24725/api/winners \
  -H "Content-Type: application/json" \
  -d '{
    "player": "Vinicius Jr.",
    "nationality": "Brazil",
    "club": "Real Madrid",
    "year": 2025,
    "votes": 650,
    "position": "Forward",
    "goals_that_season": 26
  }'
```

### PUT — Reemplazar registro completo
```bash
curl -X PUT http://localhost:24725/api/winners/1 \
  -H "Content-Type: application/json" \
  -d '{
    "player": "Lionel Messi",
    "nationality": "Argentina",
    "club": "FC Barcelona",
    "year": 2009,
    "votes": 500,
    "position": "Forward",
    "goals_that_season": 38
  }'
```

### PATCH — Actualizar campo parcial
```bash
curl -X PATCH http://localhost:24725/api/winners/1 \
  -H "Content-Type: application/json" \
  -d '{"votes": 500}'
```

### DELETE — Eliminar registro
```bash
curl -X DELETE http://localhost:24725/api/winners/15
```

---

## Casos de error

### 404 — No encontrado
```bash
curl http://localhost:24725/api/winners/999
# {"error":"Not Found","code":404,"message":"winner not found"}
```

### 422 — Validación fallida
```bash
curl -X POST http://localhost:24725/api/winners \
  -H "Content-Type: application/json" \
  -d '{"player":"","nationality":"Argentina","club":"Barcelona","year":2009,"votes":400,"position":"Forward","goals_that_season":38}'
# {"error":"Unprocessable Entity","code":422,"message":"player is required"}
```

### 400 — Campo desconocido en PATCH
```bash
curl -X PATCH http://localhost:24725/api/winners/1 \
  -H "Content-Type: application/json" \
  -d '{"unknown_field": "value"}'
# {"error":"Bad Request","code":400,"message":"unknown field: unknown_field"}
```

---

## Ejecución

### Con Docker Compose
```bash
docker-compose up --build
```

### Sin Docker
```bash
go run .
```

Servidor en `http://localhost:24725`.

---

## Formato de respuesta

**Éxito:** `{ "data": { ... } }`  
**Error:** `{ "error": "Not Found", "code": 404, "message": "winner not found" }`

---

## Validaciones

| Campo               | Regla                                             |
|---------------------|---------------------------------------------------|
| `player`            | Requerido, no vacío                               |
| `nationality`       | Requerido, no vacío                               |
| `club`              | Requerido, no vacío                               |
| `year`              | Entre 1956 y 2100                                 |
| `votes`             | No negativo                                       |
| `position`          | `Forward`, `Midfielder`, `Defender`, `Goalkeeper` |
| `goals_that_season` | No negativo                                       |

---

## Persistencia

Cada escritura (POST, PUT, PATCH, DELETE) guarda cambios en `data/winners.json` inmediatamente, con mutex para thread-safety.
