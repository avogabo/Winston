# FileBot persistence note

Winston debe persistir la licencia de FileBot igual que aprendimos en Alfred:

- usar un volumen persistente para `/config/filebot`
- fijar `FILEBOT_HOME=/config/filebot`
- no ejecutar `filebot --license` en cada trabajo
- si se usa licencia automática al arranque, leer desde `/config/filebot/license.psm`

Objetivo: no generar altas/realtas repetidas ni volver a ensuciar la activación.
