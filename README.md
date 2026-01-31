embed-store
---
Сервис для хранения, векторного поиска и кластеризации текстовых данных на основе эмбеддингов.

Проект реализует:

- хранение чанков текста с эмбеддингами фиксированной размерности (384),
- REST API для добавления и получения данных,
- векторный поиск с использованием pgvector,
- оффлайн-кластеризацию данных (k-means),
- импорт большого датасета из CSV батчами и параллельно.

Проект предназначен для исследовательских и учебных целей.

---

## Использовано 

- **Go**
- **PostgreSQL 16**
- **pgvector**
- **Docker / Docker Compose**
- **golang-migrate**

---

## Архитектура

Сервис состоит из следующих логических частей:

- **Database layer**
  Работа с PostgreSQL, pgvector, SQL-запросы, batch-вставки.

- **Importer**
  Импорт CSV-файла с чанками и эмбеддингами с поддержкой батчей и многопоточности.

- **Clusterization**
  Реализация k-means кластеризации с евклидовой метрикой (L2).

- **HTTP API**
  REST API для добавления, получения и поиска данных.

Все компоненты запускаются в одном сервисе и управляются флагами конфигурации.

---

## Быстрый старт (Docker)

### 1. Подготовка окружения

Скопировать файл с переменными окружения:

```bash
cp .env.example .env
```

Отредактировать `.env` при необходимости.

---

### 2. Запуск проекта

```bash
docker compose up -d --build
```

Поднимаются контейнеры:

- `database` – PostgreSQL + pgvector
- `migrate` – применение миграций
- `service` – основной Go-сервис

---

## Переменные окружения
Основные параметры:
```env
# Database
DB_HOST=[host]
DB_PORT=[port]
DB_USER=[user]
DB_PASSWORD=[password]
DB_NAME=[name]
DB_SSLMODE=[disable/enable]

# HTTP
HTTP_ADDR=[ardress:port]

# Logging
LOG_LEVEL=[info/warn/error/debug]

# Import
RUN_IMPORT=[true/false]
IMPORT_FILE=/data/sample.csv
IMPORT_WORKERS=[recommend: 6]
IMPORT_BATCH_SIZE=[recommend: 2000]
IMPORT_LIMIT=[recommend: 0]

# Cluster
RUN_CLUSTER=[true/false]
CLUSTER_COUNT=[recommend: 64]
CLUSTER_ITERS=[recommend: 10]
CLUSTER_WORKERS=[recommend: 11]
CLUSTER_LIMIT=[recommend: 0]
CLUSTER_BATCH_SIZE=[recommend: 5000]
```

Дополнительные параметры для импорта и кластеризации задаются через конфигурацию сервиса.

---

## Миграции

Миграции находятся в директории:

```
db/migrations
```

Создание новой миграции:

```bash
migrate create -ext sql -dir db/migrations -seq migration_name
```

Применение миграций выполняется автоматически контейнером `migrate` при запуске `docker compose`.

---

## REST API

### Добавление чанка

`POST /chunks`

Пример запроса:

```json
{
  "doc_id": 67,
  "title": "Example title",
  "author": "anonymous",
  "text": "Some text content",
  "time": "2026-01-27T11:22:33Z",
  "type": "story",
  "score": 10,
  "deleted": false,
  "dead": false,
  "embedding": [0.01, 0.02, "... 384 values ..."],
  "chunk_no": 1,
  "chunk_start": 0,
  "chunk_end": 100
}
```

Ответ:

- `201 Created` – успешно
- `409 Conflict` – дубликат `(doc_id, chunk_no)`
- `400 Bad Request` – ошибка валидации

---

### Получение чанка по ID

`GET /chunks/{id}`

Без эмбеддинга:

```bash
GET /chunks/1
```

С эмбеддингом:

```bash
GET /chunks/1?embed=1
```

---

### Векторный поиск

`POST /search`

Пример запроса:

```json
{
  "embedding": [0.01, 0.02, "... 384 values ..."],
  "limit": 5,
  "include_embedding": false
}
```

Поиск реализован через запрос:

```sql
ORDER BY embedding <-> $1
LIMIT $2
```

Используется евклидово расстояние (L2).

---

### Поиск с ограничением по кластерам

`POST /search/clusters`

Пример:

```json
{
  "embedding": [0.01, 0.02, "... 384 values ..."],
  "clusters": [1, 2, 3],
  "limit": 10
}
```

Сначала данные фильтруются по `cluster_id`, затем сортируются по расстоянию до вектора запроса.

---

## Импорт данных

Импорт запускается при:

```env
RUN_IMPORT=true
```

CSV-файл монтируется в контейнер через volume.

Импорт:

- читает файл батчами,
- использует несколько воркеров,
- корректно обрабатывает дубликаты.

---

## Кластеризация

Кластеризация запускается при:

```env
RUN_CLUSTER=true
```

Используется алгоритм **k-means**:

- фиксированная размерность векторов (384),
- евклидова метрика (L2),
- результат записывается в поле `cluster_id`.

Кластеризация выполняется оффлайн и не блокирует работу HTTP API.

---

## Производительность (локальные замеры)

Датасет: ~200 000 строк

- Импорт: ~711 секунд
- Кластеризация: ~107 секунд

Параметры:
```dotenv
# Import
IMPORT_FILE=/data/sample.csv
IMPORT_WORKERS=6
IMPORT_BATCH_SIZE=2000
IMPORT_LIMIT=0

# Cluster
CLUSTER_COUNT=64
CLUSTER_ITERS=10
CLUSTER_WORKERS=11
CLUSTER_LIMIT=200000
CLUSTER_BATCH_SIZE=5000
```

---

## Структура проекта

```
cmd/service        – entrypoint
internal/runcfg    – загрузка конфигурации
internal/db        – работа с БД
internal/httpapi   – HTTP handlers
internal/importer  – CSV импорт
internal/cluster   – кластеризация (k-means)
db/migrations      – миграции
```
