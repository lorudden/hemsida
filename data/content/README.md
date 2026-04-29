# Content Repository

Det här katalogträdet är den filbaserade lagringsytan för synligt sid- och nyhetsinnehåll som ska migreras från `https://löranshamnförening.se`.

## Format

- varje `*.json`-fil representerar en samling innehåll, till exempel sidor eller nyheter
- varje samling innehåller ett stabilt `id`, en `title` och en lista `items`
- varje item beskriver källsida, planerad målväg och migrationsstatus för synligt innehåll
- bilder kan ligga under `image_collections` per sida eller nyhet, så flera bilder kan återanvändas i till exempel karuseller
- importerade bildreferenser behåller alltid `source_url` för spårbarhet
- vanliga importkörningar kan samtidigt ladda ner WordPress-bilder till `data/media/legacy/` och fylla på `storage_path`, `mime_type` och `image_import_mode: "downloaded"`
- `-dry-run` lämnar manifesten orörda och laddar inte ner några bildfiler

## Syfte

- ge oss versionshanterad inventering av publika sidor och nyheter
- göra det lätt att lägga till testdata innan vi bygger editor- eller databasstöd
- hålla innehållsmigrering och media-migrering separata men strukturellt lika
- låta oss bygga UI-komponenter för flera bilder samtidigt som vi successivt tar över ägarskapet av äldre bildfiler

API-exponering finns på `/api/content` och läser katalogen från `CONTENT_DATA_PATH` eller flaggan `-content-data`.
