# Diccionari de recursos lexicals

*Catalan version available below / Versió en català disponible a continuació*

Source code for the online version of the Diccionari de recursos lexicals ([DIRELEX](https://direlex.softcatala.org/)).

## Dependencies

### Option 1: Docker execution

- Docker Compose

### Option 2: Native compilation with Go

- Go 1.24+

## Build and execution

### Static mode

The website can be generated as a 100% static site and served with Caddy:

```bash
docker compose -f deploy/docker-compose.static.yml up
```

### Go server mode (development)

#### Option 1: Docker

```bash
docker compose up
```

#### Option 2: Native Go

```bash
go run ./cmd/build-assets
go build -o direlex ./cmd/server
./direlex
```

Alternatively, you can use `make start` as a shortcut. Run `make` to see all available commands.

## Copyright and licenses

Copyright (c) Pere Orga Esteve <pere@orga.cat>, 2025.

The source code of this project is distributed under the [AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html.en) license or later.

### Dictionary data

Copyright (c) 2025 Carles Castellanos i Llorenç, Agustí Mayor i Lloret.

The dictionary data included in this repository is subject to a different license than the source code: [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/deed.en).

---

# Diccionari de recursos lexicals

Codi font de la versió en línia del Diccionari de recursos lexicals ([DIRELEX](https://direlex.softcatala.org/)).

## Dependències

### Opció 1: execució amb Docker

- Docker Compose

### Opció 2: compilació nativa amb Go

- Go 1.24+

## Compilació i execució

### Mode estàtic

El lloc web es pot generar com un lloc 100% estàtic i servir-lo amb Caddy:

```bash
docker compose -f deploy/docker-compose.static.yml up
```

### Mode servidor Go (desenvolupament)

#### Opció 1: Docker

```bash
docker compose up
```

#### Opció 2: Go natiu

```bash
go run ./cmd/build-assets
go build -o direlex ./cmd/server
./direlex
```

També podeu utilitzar `make start` com a drecera. Executeu `make` per veure totes les ordres.

## Copyright i llicències

Copyright (c) Pere Orga Esteve <pere@orga.cat>, 2025.

El codi font d'aquest projecte es distribueix amb la llicència [AGPL-3.0](https://www.gnu.org/licenses/agpl-3.0.html.en) o superior.

### Dades del diccionari

Copyright (c) 2025 Carles Castellanos i Llorenç, Agustí Mayor i Lloret.

Les dades del diccionari incloses en aquest repositori tenen una llicència diferent de la del codi font: [CC BY-NC 4.0](https://creativecommons.org/licenses/by-nc/4.0/deed.ca).
