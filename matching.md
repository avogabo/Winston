# Matching strategy

## Objetivo

Resolver el mejor match posible para cada NZB antes de publicarlo en AltMount.

## Orden de prioridad

1. IDs explícitos (`tmdb_id`, `tvdb_id`, `imdb_id`)
2. `relative_path_override`
3. metadata manual (`title`, `year`, `kind`, `season`, `episode`)
4. FileBot
5. parsing básico del nombre del NZB como último recurso

## Salidas

El matcher debe producir:

- metadata normalizada
- confianza (`high|medium|low`)
- candidato principal
- candidatos alternativos si hay duda
- `proposed_path`

## Regla de autoimport

- `high` => autoimport
- `medium` => configurable, por defecto autoimport si no hay conflicto
- `low` => `needs_review`

## Correcciones

El usuario puede corregir usando:

- id exacto
- elegir candidato
- override de path
- cambio de kind
