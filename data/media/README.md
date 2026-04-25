# Media Repository

Det här katalogträdet är den första filbaserade lagringsytan för publikt material som ska migreras från `https://löranshamnförening.se`.

## Format

- varje `*.json`-fil representerar en samling media
- varje samling innehåller ett stabilt `id`, en `title` och en lista `items`
- varje item kan beskriva både källdata från nuvarande webbplats och planerad lagringsplats i den nya sajten

## Syfte

- ge oss testdata utan att införa databasschema för tidigt
- göra innehållsinventeringen versionshanterad och lätt att granska
- skapa en tydlig mellanlandning inför senare import till databaser eller editorstöd

API-exponering finns på `/api/media` och läser katalogen från `MEDIA_DATA_PATH` eller flaggan `-media-data`.
