# Content Repository

Det här katalogträdet är den filbaserade lagringsytan för synligt sid- och nyhetsinnehåll som ska migreras från `https://löranshamnförening.se`.

## Format

- varje `*.json`-fil representerar en samling innehåll, till exempel sidor eller nyheter
- varje samling innehåller ett stabilt `id`, en `title` och en lista `items`
- varje item beskriver källsida, planerad målväg och migrationsstatus för synligt innehåll

## Syfte

- ge oss versionshanterad inventering av publika sidor och nyheter
- göra det lätt att lägga till testdata innan vi bygger editor- eller databasstöd
- hålla innehållsmigrering och media-migrering separata men strukturellt lika

API-exponering finns på `/api/content` och läser katalogen från `CONTENT_DATA_PATH` eller flaggan `-content-data`.
