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
- sökvägarna kan styras med `CONTENT_DATA_PATH` eller `-content-data`, samt `MEDIA_DATA_PATH` eller `-media-data`

Det här är tänkt som mellanlagring för migrering från `https://löranshamnförening.se` tills vi verkligen behöver databasscheman för redigering, behörigheter och communityfunktioner.

### Bygga och testköra hemsidan med docker

```bash
# bygg en docker image av hemsidan
docker build -f deployments/Dockerfile -t lorudden/hemsida:latest .
# starta upp en container
docker run --rm -p 8080:8080 lorudden/hemsida:latest
# surfa till http://localhost:8080/ för att titta på resultatet
```
