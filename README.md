# Winston

Winston es un servicio puente para Unraid que:

1. toma NZBs existentes desde una carpeta fuente,
2. decide la ruta lógica final que AltMount debe publicar,
3. importa de forma secuencial en AltMount,
4. y dispara escaneos selectivos en Plex solo sobre lo recién añadido.

## Regla sagrada

**Ni Winston ni AltMount deben borrar jamás los NZBs origen.**

Consecuencias de diseño:

- Winston trata `source_nzb_path` como read-only.
- Winston no mueve, no renombra y no borra NZBs origen.
- Winston no hace cleanup destructivo sobre la carpeta fuente.
- Si algún flujo necesita duplicar algo, será por copia explícita, nunca por move/delete.

## Comportamiento principal

Winston trabaja en modo secuencial para no saturar AltMount:

1. toma un NZB,
2. calcula `relative_path`,
3. llama a `POST /api/import/file` en AltMount,
4. espera a que AltMount lo procese,
5. hace `sleep` configurable, por defecto `3s`,
6. pasa al siguiente.

## Idea de request hacia AltMount

```json
{
  "file_path": "/nzb/pelis/1080/avatar.nzb",
  "relative_path": "Peliculas/1080/A/Avatar (2000) {tvdb 1234}.mkv"
}
```


## Metadata prevista por item

Winston debe poder trabajar con metadata explícita por item, no solo deducida del nombre:

- `tmdb_id`
- `tvdb_id`
- `imdb_id`
- `kind` (`movie|series|episode|auto`)
- `title`
- `year`
- `season`
- `episode`
- `quality`
- `relative_path_override`

Si viene `tmdb_id`, Winston debe priorizarlo como pista fuerte para el naming final y para FileBot cuando aplique.

## MVP previsto

- cola interna secuencial
- cliente API para AltMount
- generador de `relative_path`
  - modo inicial `preserve`: conserva árbol relativo de la carpeta NZB
- cliente básico para Plex
- almacenamiento local simple para estado
- Docker listo para despliegue en Unraid

## Configuración prevista

- `WINSTON_ALTMOUNT_BASE_URL`
- `WINSTON_ALTMOUNT_API_KEY`
- `WINSTON_PLEX_BASE_URL`
- `WINSTON_PLEX_TOKEN`
- `WINSTON_SLEEP_BETWEEN_IMPORTS`
- `WINSTON_SOURCE_ROOT`
- `WINSTON_DEFAULT_MODE=preserve|template|filebot|manual`
- `WINSTON_MOVIES_TEMPLATE`
- `WINSTON_SERIES_TEMPLATE`

## Estados iniciales recomendados

- concurrencia fija: `1`
- sleep entre imports: `3s`
- política de borrado NZB: inexistente
- política de import: secuencial y conservadora
- naming inicial: `preserve` para conservar el árbol relativo del origen NZB

## Reglas de revisión y corrección

Winston no debe pedir confirmación para todo.

- Si el match es claro, importa automáticamente.
- Si el match es dudoso, deja preview y pasa a revisión.
- Si un item ya importado quedó mal identificado, debe poder corregirse después y volver a publicarse correctamente.

Esto convierte a Winston en un sistema:
- automático cuando hay confianza,
- prudente cuando hay ambigüedad,
- corregible tras importación.

## Plex path mapping

Si AltMount/Winston y Plex no ven la misma ruta física, Winston debe traducir la ruta antes del refresh selectivo.

Ejemplo:

- AltMount: `/home/Peliculas/1080/A/Avatar (2009).mkv`
- Plex: `/media/biblioteca/Peliculas/1080/A/Avatar (2009).mkv`

Variables:

- `WINSTON_PLEX_PATH_FROM`
- `WINSTON_PLEX_PATH_TO`
