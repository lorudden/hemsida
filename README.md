# Hemsida

Här bygger vi vårt nya community tillsammans

## Sidor som skall med

I första versionen så fokuserar vi på de sidor som inte kräver någon inloggning.

* Välkommen - https://löranshamnförening.se
* Lörudden - https://löranshamnförening.se/?page_id=69
    * Fiskemuséet
    * Hjärtstartare
    * Kapellet
* Arrendatorsföreningen
* Hamnföreningen - https://löranshamnförening.se/?page_id=91
* Nyheter - https://löranshamnförening.se/?page_id=80
* Bildgalleri
    * Tidslinje
    * Veckans bild

* Flöde
    * Nyheter
    * Pressrelease
    * Kalender
        * Möteskallelser
        * Inbjudningar

* Dokument
    * https://löranshamnförening.se/wp-content/uploads/2013/02/slutrapport_loran_bramon_2013-01-14.pdf


* Länkar
    * https://bremön.se
    * http://www.marinetraffic.com/en/ais/home/centerx:17.6/centery:62.2/zoom:10
    * https://www.kustvägen.se
    * https://löransfiskemuseum.se/
    * Lörudden Foton - https://m.facebook.com/groups/475010929359458/
    * https://sillmans.se/loran/
    * https://www.sannabremo.se/sv/
    * Skatan - https://tomtarna.nu
    * https://www.tralfisketimedelpad.se

* Väderleksrapporten
    * https://www.sjoraddning.se/api/weather/get-weather/bramon
* Passerande fartyg?
    * Skrapa data från marine traffic?
* Passerande flygplan?
    * Hämta data från https://opensky-network.org/api
* Karta?

## Tekniskt mumbo jumbo

### Filbaserad migreringsdata

Den nya sajten har nu filbaserade kataloger för både innehåll och media.

- `data/content/` innehåller sidor och nyheter och exponeras på `GET /api/content`
- `data/media/` innehåller dokument och galleriinnehåll och exponeras på `GET /api/media`
- varje `*.json`-fil motsvarar en samling som laddas vid uppstart
- importverktyget kan ladda ner äldre WordPress-bilder till `data/media/legacy/` vid importtid så att den nya sajten slipper mixed-content-problem över `https`
- sidor och nyheter kan lagra flera bilder i `image_collections`, vilket gör det lättare att bygga karuseller och liknande visningar senare
- sökvägarna kan styras med `CONTENT_DATA_PATH` eller `-content-data`, samt `MEDIA_DATA_PATH` eller `-media-data`

Det här är tänkt som mellanlagring för migrering från `https://löranshamnförening.se` tills vi verkligen behöver databasscheman för redigering, behörigheter och communityfunktioner.

### Importverktyg

Ett första importverktyg finns under `cmd/importcontent/`.

- verktyget läser `data/content/pages.json` och `data/content/news.json`
- varje entry hämtar sin `source_page_url` och försöker fylla på titel, summary, publiceringsdatum och bildreferenser
- vid vanlig körning laddas bildfiler ner till `data/media/legacy/`, medan manifesten behåller `source_url` och kompletteras med `storage_path` och `mime_type`
- `-dry-run` skriver varken manifest eller nedladdade bildfiler

Exempel:

```bash
go run ./cmd/importcontent -dry-run
go run ./cmd/importcontent
go run ./cmd/importcontent -media-dir .tmp/import-preview/media
```

### Bygga och testköra hemsidan med docker

```bash
# bygg en docker image av hemsidan
docker build -f deployments/Dockerfile -t lorudden/hemsida:latest .
# starta upp en container
docker run --rm -p 8080:8080 lorudden/hemsida:latest
# surfa till http://localhost:8080/ för att titta på resultatet
```
