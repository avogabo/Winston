# Winston workflow

## Objetivo

Winston decide cómo publicar NZBs en AltMount usando metadata, reglas configurables y una capa de revisión cuando el match no es suficientemente fiable.

## Principios

1. Si el match es claro, Winston importa automáticamente.
2. Si el match es dudoso, Winston no importa todavía y deja el item en revisión.
3. Si el item ya fue importado pero el match/ruta es incorrecto, Winston debe permitir corrección posterior.
4. Bajo ninguna circunstancia Winston borra el NZB origen.

## Estados por item

- `detected`: Winston ha detectado el item y ha generado una propuesta inicial.
- `needs_review`: el match no tiene suficiente confianza o existe conflicto.
- `approved`: el item tiene match final aceptado y puede importarse.
- `importing`: Winston lo ha enviado a AltMount y espera finalización.
- `imported`: AltMount completó el import.
- `failed`: el import o la resolución falló.
- `corrected`: el item fue corregido manualmente tras una detección o import erróneo.

## Regla de autoimport

Winston puede autoimportar sin confirmación si se cumple al menos una de estas condiciones fuertes:

- viene `tmdb_id`, `tvdb_id` o `imdb_id` explícito,
- el match es único y consistente con título+año,
- la confianza supera el umbral configurado,
- no existe conflicto relevante entre candidatos.

## Regla de revisión

Winston debe pasar a `needs_review` cuando:

- hay múltiples candidatos plausibles,
- hay ambigüedad fuerte entre obras parecidas,
- falta el año y el título no es fiable,
- el tipo (movie/series) no es sólido,
- la metadata inferida contradice la estructura esperada.

## Preview

Cada item debe poder mostrar preview con:

- `source_nzb_path`
- `kind`
- metadata detectada
- candidato principal
- lista opcional de candidatos alternativos
- `relative_path` propuesto
- motivo de duda si aplica
- nivel de confianza

## Corrección manual

El usuario debe poder corregir un item por cualquiera de estas vías:

- indicar `tmdb_id`
- indicar `tvdb_id`
- indicar `imdb_id`
- elegir un candidato de una lista de coincidencias
- forzar `relative_path_override`
- cambiar `kind`

## Corrección post-import

Si un item ya fue importado y se detecta un match incorrecto:

1. Winston registra la corrección.
2. Winston recalcula metadata y `relative_path`.
3. Winston debe republicar/reimportar correctamente en AltMount.
4. El NZB origen permanece intacto.

La operación exacta sobre AltMount dependerá de si el backend soporta update directo o requiere reimport controlado.

## Backpressure

Winston siempre procesa secuencialmente:

1. un item,
2. espera a `completed` en AltMount,
3. sleep configurado,
4. siguiente item.

Nunca se hace un volcado masivo de imports a la cola.
