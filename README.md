# Winston

Winston vigila una carpeta de NZBs, decide la ruta lógica final, importa en AltMount de forma secuencial y expone una UI para revisar/corregir antes de publicar.

## Qué hace

- no borra, mueve ni renombra los NZBs origen
- importa en AltMount de uno en uno
- espera a que AltMount termine antes de seguir
- puede usar `preserve`, `template` o `filebot`
- permite revisar y corregir desde web
- guarda ajustes persistentes en `.winston-settings.json`

## Instalación simple desde cero

### 1. Necesitas

- AltMount funcionando
- una carpeta con NZBs
- licencia de FileBot si quieres modo `filebot`

### 2. Prepara carpetas persistentes

Ejemplo:

```bash
mkdir -p /opt/winston/config/filebot
mkdir -p /opt/winston/nzb
```

Copia tu licencia de FileBot aquí:

```bash
/opt/winston/config/filebot/license.psm
```

Importante:
- Winston activa esa licencia al arrancar
- el estado activado queda persistido en `/config/filebot/data`
- así no parece un equipo nuevo en cada recreate

### 3. Arranca con Docker

```bash
docker run -d \
  --name winston \
  --restart unless-stopped \
  -p 8091:8091 \
  -e WINSTON_SOURCE_ROOT=/data/nzb \
  -e WINSTON_ALTMOUNT_BASE_URL=http://TU_ALTMOUNT:8989 \
  -e WINSTON_ALTMOUNT_API_KEY=TU_API_KEY \
  -e WINSTON_DEFAULT_MODE=filebot \
  -e FILEBOT_HOME=/config/filebot \
  -v /opt/winston/config:/config \
  -v /opt/winston/nzb:/data/nzb \
  ghcr.io/avogabo/winston:latest
```

### 4. Abre la UI

- `http://TU_HOST:8091`

### 5. Ajustes mínimos recomendados

En la pestaña **Ajustes**:

- `AltMount Base URL`
- `AltMount API Key`
- `Ruta Winston NZB`
- `Ruta visible en AltMount`
- `Modo de import = filebot`

Si usas Plex también:

- `Plex Base URL`
- `Plex Token`
- `Plex Path From`
- `Plex Path To`

## Docker Compose de ejemplo

```yaml
services:
  winston:
    image: ghcr.io/avogabo/winston:latest
    container_name: winston
    restart: unless-stopped
    ports:
      - "8091:8091"
    environment:
      WINSTON_HTTP_LISTEN_ADDR: ":8091"
      WINSTON_SOURCE_ROOT: /data/nzb
      WINSTON_ALTMOUNT_BASE_URL: http://192.168.1.100:8989
      WINSTON_ALTMOUNT_API_KEY: TU_API_KEY
      WINSTON_DEFAULT_MODE: filebot
      WINSTON_SLEEP_BETWEEN_IMPORTS: 3s
      WINSTON_AUTOIMPORT_MEDIUM: "true"
      WINSTON_FILEBOT_FORMAT_MOVIE: Peliculas/{plex}
      WINSTON_FILEBOT_FORMAT_SERIES: Series/{plex}
      WINSTON_FILEBOT_DB: TheMovieDB
      FILEBOT_HOME: /config/filebot
    volumes:
      - /opt/winston/config:/config
      - /opt/winston/nzb:/data/nzb
```

## Plantilla rápida para Unraid

### Puertos
- `8091:8091`

### Path mappings
- `/config` → appdata persistente de Winston
- `/data/nzb` → carpeta real de NZBs

### Variables importantes
- `WINSTON_SOURCE_ROOT=/data/nzb`
- `WINSTON_ALTMOUNT_BASE_URL=http://IP_UNRAID:8989`
- `WINSTON_ALTMOUNT_API_KEY=...`
- `WINSTON_DEFAULT_MODE=filebot`
- `FILEBOT_HOME=/config/filebot`

### Archivo que debes copiar una vez
- `/config/filebot/license.psm`

## Cómo funciona FileBot aquí

Winston ya hace esto automáticamente al arrancar:

- usa FileBot `5.1.6`
- fija `FILEBOT_HOME=/config/filebot`
- enlaza `/opt/filebot/data` a `/config/filebot/data`
- activa `license.psm` si aún no existe `/config/filebot/data/.license`
- reutiliza la activación en siguientes recreates

## Defaults recomendados

### Películas
- DB: `TheMovieDB`
- formato: `Peliculas/{plex}`

### Series
- formato: `Series/{plex}`

Nota: la elección de DB para series puede depender del comportamiento real de FileBot y sus proveedores en cada versión. La integración actual prioriza estabilidad operativa.

## API disponible

- `GET /api/review/items`
- `GET /api/review/item?source=...`
- `POST /api/review/correct?source=...`
- `POST /api/review/approve?source=...`
- `POST /api/review/import?source=...`
- `GET /api/settings`
- `POST /api/settings`
- `GET /api/filebot/status`

## Variables de entorno

```env
WINSTON_HTTP_LISTEN_ADDR=:8091
WINSTON_SOURCE_ROOT=/data/nzb
WINSTON_ALTMOUNT_BASE_URL=http://altmount:8989
WINSTON_ALTMOUNT_API_KEY=
WINSTON_PLEX_BASE_URL=http://plex:32400
WINSTON_PLEX_TOKEN=
WINSTON_PLEX_PATH_FROM=/home
WINSTON_PLEX_PATH_TO=/media/biblioteca
WINSTON_DEFAULT_MODE=filebot
WINSTON_SLEEP_BETWEEN_IMPORTS=3s
WINSTON_AUTOIMPORT_MEDIUM=true
WINSTON_MOVIES_TEMPLATE=Peliculas/{quality}/{alpha}/{title} ({year})
WINSTON_SERIES_TEMPLATE=Series/{alpha}/{series}/Temporada {season}/{series} - {episode}
WINSTON_FILEBOT_FORMAT_MOVIE=Peliculas/{plex}
WINSTON_FILEBOT_FORMAT_SERIES=Series/{plex}
WINSTON_FILEBOT_DB=TheMovieDB
WINSTON_FILEBOT_BINARY=/usr/local/bin/filebot
FILEBOT_HOME=/config/filebot
```

## Desarrollo local

```bash
go build ./...
cp .env.example .env
set -a; source .env; set +a
./winston
```

Frontend:

```bash
cd web
npm install
npm run dev
```

## Estado actual

Winston ya está listo para:

- instalación con Docker simple
- persistencia correcta de licencia FileBot
- UI de ajustes
- revisión y corrección
- importación secuencial hacia AltMount

Lo que más ayuda a un usuario nuevo es montar bien `/config`, `/data/nzb` y copiar `license.psm` una sola vez.
