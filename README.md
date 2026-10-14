<h1 align="center">keqrnel</h1>

<p align="center">˚ʚ♡ɞ˚</p>

<p align="center">
  <strong>English</strong> · <a href="#русский">Русский</a>
</p>

<p align="center">
  A single proxy core: a <strong>sing-box host</strong> with an <strong>embedded xray engine</strong>.<br>
  TUN, routing and DNS from sing-box; the protocols run on xray's own code.
</p>

<p align="center">
  <a href="https://github.com/Lemonochka/keqrnel/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/Lemonochka/keqrnel/ci.yml?branch=master&label=build&style=flat-square" alt="build"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-c9b8f5?style=flat-square" alt="license"></a>
  <img src="https://img.shields.io/badge/Go-1.26+-9bc7f0?style=flat-square&logo=go&logoColor=white" alt="go">
</p>

---

## What it is

sing-box is the host: TUN, routing, DNS, sniffing, `clash_api` traffic counters.
xray-core is linked in as a library and serves one outbound type, `"xray"`, which
handles VLESS, XHTTP, Reality, Vision, Mux, Hysteria2 and anything else its JSON
config takes. Protocols run on xray's own code, so the wire format is what a server
configured for upstream xray expects.

```
TUN (sing-tun) ─► routing ─► outbound "proxy" (type: xray)
                                      │
                                      ▼
                              embedded xray.Instance
```

## The `xray` outbound

```json
{
  "type": "xray",
  "tag": "proxy",
  "xray": {
    "outbounds": [
      {
        "protocol": "vless",
        "settings": { "address": "example.com", "port": 443, "id": "...", "encryption": "none" },
        "streamSettings": { ... }
      }
    ]
  }
}
```

The `xray` field is a raw xray config fragment, at minimum an `outbounds` array,
handed to the embedded engine as-is. Everything around it is ordinary sing-box
config. Full example with TUN and routing: [config.example.json](config.example.json).

Current xray puts the server and credentials directly under `settings`. Older
configs that wrap them in a `vnext` array still parse, so both forms work.

Optional `sniffing` takes the block of an xray inbound (`enabled`, `destOverride`,
`metadataOnly`, `routeOnly`) and makes xray sniff every connection it is handed.
Under TUN sing-box passes the destination as an IP even when it sniffed the domain
itself, so without it xray's `domain` and `geosite` rules never match there.

Hysteria2 is an xray outbound like any other and goes in the same fragment.
sing-box's `socks` and `http` outbounds are registered as well, so keqrnel can also
front a local proxy upstream.

## Build

```sh
go run ./tools/patchxray
go test -modfile=go.patched.mod ./...
go build -modfile=go.patched.mod -tags with_gvisor -o keqrnel ./cmd/keqrnel
```

keqrnel is built against the xray-core pinned in `go.mod` with `patches/xray-*.patch`
applied. `tools/patchxray` downloads that revision, applies the patches (git must be on
PATH) and writes `go.patched.mod`, which differs from `go.mod` by one `replace`. A plain
`go build` still works and gives the unpatched core; the patch headers say what that
loses, and `TestEngineUDPKeepsBurst` fails there on purpose. `with_gvisor` brings the
`gvisor` and `mixed` TUN stacks.

Go 1.26+, required by xray-core. Native on Windows and Linux, cross-compiles to
`linux/arm64` and `darwin/arm64`.

## Run

```sh
keqrnel config.json     # defaults to ./config.json
keqrnel -c config.json  # xray-style flags work too, for drop-in launching
```

A TUN inbound needs administrator/root, plus `wintun.dll` next to the binary on
Windows. A SOCKS/HTTP inbound does not.

The process stops on SIGINT/SIGTERM, and on stdin EOF. A launcher that holds a stdin
pipe can close it to request a clean shutdown rather than killing the process, which
would leave the TUN adapter, routes and DNS unreverted.

## Layout

```
cmd/keqrnel/     entry point
service/         New / Start / Close
core/            Box assembly, curated registries
core/localdns/   the "local" DNS transport sing-box requires as a fallback
bridge/xray/     embedded xray engine + the "xray" outbound
```

How the bridge works, and why the registries are curated rather than pulled in with
sing-box's `include.*`: [ARCHITECTURE.md](ARCHITECTURE.md).

## Versions

sing-box v1.13.19 · xray-core v26.7.28 · Go 1.26+

## License

[GPL-3.0-or-later](LICENSE). Links [sing-box](https://github.com/SagerNet/sing-box)
(GPL-3.0-or-later) and embeds [xray-core](https://github.com/XTLS/Xray-core)
(MPL-2.0). Not affiliated with either project.

---

<h2 id="русский">Русский</h2>

<p align="center">
  <a href="#keqrnel">English</a> · <strong>Русский</strong>
</p>

Единое прокси-ядро: **host на sing-box со встроенным движком xray**. TUN, роутинг и
DNS — от sing-box, протоколы исполняет код самого xray.

### Что это

sing-box — хост: TUN, роутинг, DNS, sniffing, счётчики трафика через `clash_api`.
xray-core влинкован как библиотека и обслуживает один тип аутбаунда, `"xray"`,
который тянет VLESS, XHTTP, Reality, Vision, Mux, Hysteria2 и всё прочее, что
принимает его JSON-конфиг. Протоколы исполняет сам xray, поэтому на проводе ровно
то, чего ждёт сервер, настроенный на свежий xray.

### Аутбаунд `xray`

```json
{
  "type": "xray",
  "tag": "proxy",
  "xray": {
    "outbounds": [
      {
        "protocol": "vless",
        "settings": { "address": "example.com", "port": 443, "id": "...", "encryption": "none" },
        "streamSettings": { ... }
      }
    ]
  }
}
```

Поле `xray` — сырой фрагмент xray-конфига, как минимум массив `outbounds`, который
отдаётся встроенному движку как есть. Всё вокруг — обычный конфиг sing-box. Полный
пример с TUN и роутингом: [config.example.json](config.example.json).

Актуальный xray кладёт адрес сервера и креды прямо в `settings`. Старые конфиги,
которые заворачивают их в массив `vnext`, тоже разбираются — работают обе формы.

Необязательный `sniffing` принимает блок входа xray (`enabled`, `destOverride`,
`metadataOnly`, `routeOnly`), и xray распознаёт каждое переданное ему соединение.
Под TUN sing-box отдаёт назначение адресом, даже если сам узнал домен, и без этого
правила xray по `domain` и `geosite` там не срабатывают.

Hysteria2 — такой же xray-аутбаунд, кладётся в тот же фрагмент. Аутбаунды `socks` и
`http` из sing-box тоже зарегистрированы, так что keqrnel умеет стоять и перед
локальным прокси-аплинком.

### Сборка

```sh
go run ./tools/patchxray
go test -modfile=go.patched.mod ./...
go build -modfile=go.patched.mod -tags with_gvisor -o keqrnel ./cmd/keqrnel
```

keqrnel собирается с xray-core из `go.mod`, на который наложены `patches/xray-*.patch`.
`tools/patchxray` скачивает эту ревизию, накладывает патчи (нужен git в PATH) и пишет
`go.patched.mod`, который отличается от `go.mod` одной строкой `replace`. Обычный
`go build` по-прежнему работает и даёт ядро без патчей; что при этом теряется, написано
в шапках патчей, а `TestEngineUDPKeepsBurst` там нарочно падает. `with_gvisor` даёт
стеки TUN `gvisor` и `mixed`.

Нужен Go 1.26+, это требование xray-core. Нативно под Windows и Linux,
кросс-компиляция под `linux/arm64` и `darwin/arm64`.

### Запуск

```sh
keqrnel config.json     # по умолчанию ./config.json
keqrnel -c config.json  # флаги в стиле xray тоже принимаются, для drop-in запуска
```

TUN-инбаунду нужны права администратора/root и `wintun.dll` рядом с бинарём на
Windows. SOCKS/HTTP-инбаунду — нет.

Процесс останавливается по SIGINT/SIGTERM и по EOF на stdin. Лаунчер, который держит
stdin-пайп, может закрыть его и попросить корректное завершение вместо убийства
процесса — иначе останутся неоткаченными TUN-адаптер, маршруты и DNS.

### Структура

```
cmd/keqrnel/     точка входа
service/         New / Start / Close
core/            сборка Box, курированные registry
core/localdns/   транспорт "local" для DNS, который sing-box требует как фолбэк
bridge/xray/     встроенный движок xray + аутбаунд "xray"
```

Как устроен мост и почему registry собраны вручную, а не через `include.*` из
sing-box: [ARCHITECTURE.md](ARCHITECTURE.md).

### Версии

sing-box v1.13.19 · xray-core v26.7.28 · Go 1.26+

### Лицензия

[GPL-3.0-or-later](LICENSE). Линкует [sing-box](https://github.com/SagerNet/sing-box)
(GPL-3.0-or-later) и встраивает [xray-core](https://github.com/XTLS/Xray-core)
(MPL-2.0). С этими проектами не аффилирован.
