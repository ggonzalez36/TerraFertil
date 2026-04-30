# TerraFertil MVP - Documentacion completa de codigo y despliegue

Este repositorio implementa un MVP de TerraFertil con tres capas:
1. Frontend: landing page estatica.
2. Backend: API REST en Go para proyectos, simulaciones y onboarding.
3. AI Module: servicio FastAPI en Python para scoring de riesgo y persistencia de evaluaciones.

## 1) Requisitos

- macOS o Linux
- Go 1.22+
- Python 3.9+
- curl

Nota:
- Puedes correr todo sin npm.
- Si npm falla por red corporativa, usa servidor estatico con Python para la landing.

## 2) Arquitectura

```text
Landing (index.html)
   |
   |  POST /api/investment/simulate
   |  POST /api/onboarding
   v
Backend Go (:8080)
   |
   |  POST /predict-risk
   v
AI Module Python (:8001)

Persistencia:
- backend-go/terrafertil.db -> onboarding (waitlist)
- ai-module/ai_module.db    -> evaluaciones de riesgo AI
```

## 3) Estructura de codigo

### Frontend
- [index.html](index.html): landing, navegación, calculadora, formulario onboarding y llamadas a API.

### Backend Go
- [backend-go/cmd/server/main.go](backend-go/cmd/server/main.go): arranque de servidor, CORS, env vars, inicialización DB.
- [backend-go/internal/api/handlers.go](backend-go/internal/api/handlers.go): rutas HTTP y validaciones.
- [backend-go/internal/service/models.go](backend-go/internal/service/models.go): DTOs request/response.
- [backend-go/internal/service/calculator.go](backend-go/internal/service/calculator.go): fórmula de simulación.
- [backend-go/internal/service/projects.go](backend-go/internal/service/projects.go): seed de proyectos.
- [backend-go/internal/store/waitlist_sqlite.go](backend-go/internal/store/waitlist_sqlite.go): acceso a SQLite para onboarding.
- [backend-go/run-backend.sh](backend-go/run-backend.sh): script recomendado de arranque.

### AI Module Python
- [ai-module/app/main.py](ai-module/app/main.py): API FastAPI y wiring del engine.
- [ai-module/scoring/engine.py](ai-module/scoring/engine.py): orquestador del scoring.
- [ai-module/scoring/calculateTerraScore.py](ai-module/scoring/calculateTerraScore.py): cálculo numérico del TerraScore.
- [ai-module/scoring/rag.py](ai-module/scoring/rag.py): capa RAG + persistencia SQLite + lectura de contexto desde backend.
- [ai-module/requirements.txt](ai-module/requirements.txt): dependencias Python.

## 4) Variables de entorno

### Backend Go
- PORT: puerto del backend (default 8080)
- AI_SERVICE_URL: URL del AI module (default http://localhost:8001)
- DATABASE_PATH: ruta SQLite onboarding (default ./terrafertil.db)

### AI Module
- BACKEND_API_URL: URL backend para consultar proyectos (default http://localhost:8080)
- AI_DATABASE_PATH: ruta SQLite AI (default ./ai_module.db)

## 5) Base de datos

### Backend SQLite
Archivo: backend-go/terrafertil.db

Tabla principal:
- waitlist_signups
  - id INTEGER PK AUTOINCREMENT
  - email TEXT UNIQUE NOT NULL
  - source TEXT NOT NULL
  - created_at TIMESTAMP NOT NULL

### AI Module SQLite
Archivo: ai-module/ai_module.db

Tabla principal:
- ai_risk_evaluations
  - id INTEGER PK AUTOINCREMENT
  - project_stage TEXT
  - ltv REAL
  - debt_ratio REAL
  - location_score REAL
  - sponsor_track_record REAL
  - risk_score REAL
  - risk_level TEXT
  - confidence REAL
  - created_at TEXT

## 6) API contratos

### Backend Go

1. GET /health
- Response 200:
```json
{"status":"ok"}
```

2. GET /api/projects
- Response 200: lista seed de proyectos Colombia/USA.

3. POST /api/investment/simulate
- Request:
```json
{"principal":5000,"annualRate":11.5,"years":5}
```
- Response 200:
```json
{"principal":5000,"annualRate":11.5,"years":5,"finalAmount":8616.76,"profit":3616.76}
```

4. POST /api/onboarding
- Request:
```json
{"email":"usuario@correo.com","source":"landing-hero"}
```
- Response 201:
```json
{"id":1,"email":"usuario@correo.com","source":"landing-hero","createdAt":"2026-04-27T00:00:00Z","status":"created"}
```
- Response 409 (duplicado):
```json
{"error":"email already registered"}
```

5. POST /api/ai/risk-score
- Proxy hacia AI Module /predict-risk.

### AI Module

1. GET /health
- Response 200:
```json
{"status":"ok"}
```

2. POST /predict-risk
- Request:
```json
{"ltv":70,"debtRatio":58,"locationScore":74,"sponsorTrackRecord":80,"projectStage":"construction"}
```
- Response 200:
```json
{"riskScore":53.5,"riskLevel":"medium","confidence":0.74,"recommendations":["Current profile looks balanced for initial screening."]}
```

## 7) Flujo funcional del codigo

### Simulador de inversión
1. Usuario mueve sliders en [index.html](index.html).
2. Frontend hace POST a /api/investment/simulate.
3. Backend calcula con [backend-go/internal/service/calculator.go](backend-go/internal/service/calculator.go).
4. Frontend actualiza cards de resultado.
5. Si backend no responde, frontend usa fallback local.

### Onboarding
1. Usuario envía email en formulario de [index.html](index.html).
2. Frontend hace POST a /api/onboarding.
3. Backend valida email y guarda en waitlist_signups.
4. Devuelve 201 o 409 si ya existe.

### Scoring AI
1. Backend recibe /api/ai/risk-score.
2. Backend llama a AI module /predict-risk.
3. AI module usa engine y TerraScore.
4. RAG agrega contexto (incluyendo consulta de proyectos al backend).
5. AI module persiste la evaluación en ai_risk_evaluations.

## 8) Levantar todo en local

### Paso A: AI Module
```bash
cd /Users/gustavogonzalez/TerraFertil/ai-module
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
BACKEND_API_URL=http://localhost:8080 AI_DATABASE_PATH=./ai_module.db \
uvicorn app.main:app --app-dir /Users/gustavogonzalez/TerraFertil/ai-module --host 0.0.0.0 --port 8001 --reload
```

### Paso B: Backend Go
```bash
cd /Users/gustavogonzalez/TerraFertil/backend-go
./run-backend.sh
```

Arranque manual equivalente:
```bash
cd /Users/gustavogonzalez/TerraFertil/backend-go
export GOROOT="$HOME/.local/go"
export GOPATH="$HOME/go"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"
AI_SERVICE_URL=http://localhost:8001 DATABASE_PATH=./terrafertil.db PORT=8080 go run ./cmd/server
```

### Paso C: Landing
```bash
cd /Users/gustavogonzalez/TerraFertil
python3 -m http.server 8000
```

Abrir:
- http://localhost:8000

## 9) Verificaciones rapidas

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/api/projects
curl -i -X POST http://localhost:8080/api/investment/simulate -H 'Content-Type: application/json' -d '{"principal":5000,"annualRate":11.5,"years":5}'
curl -i -X POST http://localhost:8080/api/onboarding -H 'Content-Type: application/json' -d '{"email":"demo@terrafertil.co","source":"landing-hero"}'
curl -i http://localhost:8001/health
curl -i -X POST http://localhost:8001/predict-risk -H 'Content-Type: application/json' -d '{"ltv":70,"debtRatio":58,"locationScore":74,"sponsorTrackRecord":80,"projectStage":"construction"}'
curl -i -X POST http://localhost:8080/api/ai/risk-score -H 'Content-Type: application/json' -d '{"ltv":70,"debtRatio":58,"locationScore":74,"sponsorTrackRecord":80,"projectStage":"construction"}'
```

## 10) Problemas comunes

### 404 en /api/onboarding
- Causa: backend viejo corriendo en :8080 sin ruta nueva.
- Solución: mata proceso de 8080 y relanza backend actualizado.

### 502 en /api/ai/risk-score
- Causa: AI module no está arriba.
- Solución: verifica http://localhost:8001/health y logs de uvicorn.

### Error source .venv/bin/activate no such file
- Causa: se ejecutó desde directorio incorrecto.
- Solución: usa ruta absoluta de venv o entra a ai-module antes de activar.

### Puerto ocupado
- Causa: proceso previo aún activo.
- Solución: identificar con lsof -nP -iTCP:8080 -sTCP:LISTEN y matar PID.

### Docker Compose levantó pero API no responde
- Causa: contenedores arriba pero servicio no listo aún.
- Solución:
  1. Ejecuta `docker compose ps`.
  2. Revisa `docker compose logs -f backend-go` y `docker compose logs -f ai-module`.
  3. Valida endpoints:
     - `curl -i http://localhost:8080/health`
     - `curl -i http://localhost:8001/health`

## 11) URLs de trabajo

- Landing: http://localhost:8000
- Backend: http://localhost:8080
- AI module: http://localhost:8001

## 12) Docker

Se agregaron Dockerfiles para backend y AI module:
- [backend-go/Dockerfile](backend-go/Dockerfile)
- [ai-module/Dockerfile](ai-module/Dockerfile)

Orquestación con un solo comando:
- [docker-compose.yml](docker-compose.yml)

Archivos de exclusión de contexto:
- [backend-go/.dockerignore](backend-go/.dockerignore)
- [ai-module/.dockerignore](ai-module/.dockerignore)

### Build local de imágenes

```bash
cd /Users/gustavogonzalez/TerraFertil
docker build -t terrafertil/backend-go:local ./backend-go
docker build -t terrafertil/ai-module:local ./ai-module
```

### Run local con Docker Compose (recomendado)

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose up -d --build
```

Validar configuración de compose:

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose config
```

Servicios levantados:
- Backend Go: http://localhost:8080
- AI module: http://localhost:8001

Ver estado y logs:

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose ps
docker compose logs -f
```

Logs por servicio:

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose logs -f backend-go
docker compose logs -f ai-module
```

Reiniciar servicios:

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose restart
```

Rebuild de un solo servicio:

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose up -d --build backend-go
docker compose up -d --build ai-module
```

Detener servicios:

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose down
```

Detener y borrar volúmenes (reinicio limpio de DBs):

```bash
cd /Users/gustavogonzalez/TerraFertil
docker compose down -v
```

### Run local con Docker (manual)

```bash
# Red local
docker network create terrafertil-net || true

# AI module
docker run -d --name ai-module \
  --network terrafertil-net \
  -p 8001:8001 \
  -e BACKEND_API_URL=http://backend-go:8080 \
  -e AI_DATABASE_PATH=/data/ai_module.db \
  terrafertil/ai-module:local

# Backend Go
docker run -d --name backend-go \
  --network terrafertil-net \
  -p 8080:8080 \
  -e AI_SERVICE_URL=http://ai-module:8001 \
  -e DATABASE_PATH=/data/terrafertil.db \
  terrafertil/backend-go:local
```

## 13) GitHub Actions (CI/CD)

Workflows agregados:
- [CI](.github/workflows/ci.yml)
- [Docker Publish](.github/workflows/docker-publish.yml)

### CI

Se ejecuta en push/pull_request y hace:
1. Tests de Go.
2. Tests de Python.
3. Build de imágenes Docker (sin push) como chequeo de integridad.

### CD (publish de imágenes)

Se ejecuta en push a main/master, tags v* y manual (`workflow_dispatch`):
1. Login a GHCR con `GITHUB_TOKEN`.
2. Build y push de:
   - `ghcr.io/<owner>/<repo>/backend-go`
   - `ghcr.io/<owner>/<repo>/ai-module`

No requiere secretos extra para GHCR en el mismo repositorio (usa `secrets.GITHUB_TOKEN` con permisos `packages: write`).

## 14) Unit tests

### Tests backend Go

```bash
export GOROOT="$HOME/.local/go"
export GOPATH="$HOME/go"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"
cd /Users/gustavogonzalez/TerraFertil/backend-go
go test ./...
```

Solo functional tests (flujo HTTP completo):

```bash
export GOROOT="$HOME/.local/go"
export GOPATH="$HOME/go"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"
cd /Users/gustavogonzalez/TerraFertil/backend-go
go test ./tests/functional -v
```

Cobertura actual:
- [backend-go/internal/service/calculator_test.go](backend-go/internal/service/calculator_test.go)
- [backend-go/internal/service/projects_test.go](backend-go/internal/service/projects_test.go)
- [backend-go/internal/store/waitlist_sqlite_test.go](backend-go/internal/store/waitlist_sqlite_test.go)
- [backend-go/tests/functional/flow_test.go](backend-go/tests/functional/flow_test.go)

### Tests AI module Python

```bash
source /Users/gustavogonzalez/TerraFertil/ai-module/.venv/bin/activate
PYTHONPATH=/Users/gustavogonzalez/TerraFertil/ai-module \
python -m unittest discover -s /Users/gustavogonzalez/TerraFertil/ai-module/tests -p 'test_*.py'
```

Cobertura actual:
- [ai-module/tests/test_calculate_terra_score.py](ai-module/tests/test_calculate_terra_score.py)
- [ai-module/tests/test_engine.py](ai-module/tests/test_engine.py)