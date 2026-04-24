# Corrections

## Tipos de corrección admitidos

- `set_tmdb_id`
- `set_tvdb_id`
- `set_imdb_id`
- `set_kind`
- `set_relative_path_override`
- `choose_candidate`

## Efecto esperado

Una corrección debe:

1. actualizar metadata del item,
2. recalcular preview y proposed path,
3. pasar el item a `corrected` o `approved`,
4. permitir reimportación/republicación si ya estaba importado y quedó mal.
