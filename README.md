<h1 align="center">keqrnel</h1>

<p align="center">˚ʚ♡ɞ˚</p>

<p align="center">
  <strong>English</strong> · <a href="#русский">Русский</a>
</p>

<p align="center">
  A single proxy core: a <strong>sing-box host</strong> with an <strong>embedded xray engine</strong>.<br>
  One process instead of the usual <code>xray → sing-box</code> pair.
</p>

<p align="center">
  <a href="https://github.com/Lemonochka/keqrnel/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/Lemonochka/keqrnel/ci.yml?branch=master&label=build&style=flat-square" alt="build"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-GPL--3.0-c9b8f5?style=flat-square" alt="license"></a>
  <img src="https://img.shields.io/badge/Go-1.26+-9bc7f0?style=flat-square&logo=go&logoColor=white" alt="go">
</p>

---

## What it is

sing-box is the host: TUN, routing, DNS, sniffing, `clash_api` traffic counters.
xray-core is linked in as a library and serves one outbound type, `"xray"`:
VLESS, XHTTP, Reality, Vision, Mux, Hysteria2, and whatever else its JSON config
takes. Protocols run on xray's own code, so the wire format is what a server
configured for upstream xray expects.

```
TUN (sing-tun) ─► routing ─► outbound "proxy" (type: xray)
                                      │
                                      ▼
                              embedded xray.Instance
```

No loopback SOCKS between engines, no second process, no port to coordinate.

Written for [keqdroid/KEQDIS](https://github.com/Lemonochka/keqdroid), but it does
not depend on the app — anything that can write a sing-box config can run it.

## The `xray` outbound

```json
{
  "type": "xray",
  "tag": "proxy",
  "xray": {
    "outbounds": [
      { "protocol": "vless", "settings": { "vnext": [ ... ] }, "streamSettings": { ... } }
    ]
  }
}
```

The `xray` field is a raw xray config fragment (at minimum an `outbounds` array),
handed to the embedded engine verbatim — that is what keeps you compatible with your
servers. Everything around it is ordinary sing-box config. Full example with TUN and
routing: [config.example.json](config.example.json).

Hysteria2 needs nothing special; it is an xray outbound like any other.

sing-box's `socks` and `http` outbounds are registered too, so keqrnel also works as
a TUN front-end for a local proxy upstream — that is how AmneziaWG is handled:
`wireproxy` exposes a local SOCKS, keqrnel wraps it into the TUN.

## Build

```sh
go build ./...
go test ./...
go build -o keqrnel ./cmd/keqrnel
```

Go 1.26+ (required by xray-core). Builds natively on Windows and Linux,
cross-compiles to `linux/arm64` and `darwin/arm64`.

> sing-box's `include.*` helpers are deliberately unused: they register everything
> sing-box ships (tor, naive, ssh, dhcp-dns, and a `dns/transport/local` that does
> not compile on Windows). [core/registry.go](core/registry.go) assembles a curated
> registry instead.

## Run

```sh
keqrnel config.json     # defaults to ./config.json
keqrnel -c config.json  # xray-style flags work too, for drop-in launching
```

A TUN inbound needs administrator/root (and `wintun.dll` next to the binary on
Windows); a SOCKS/HTTP inbound does not.

The process stops on SIGINT/SIGTERM and when its stdin pipe is closed. The second
route lets a launcher ask for a clean shutdown — on Windows a hard `TerminateProcess`
leaves the TUN adapter, routes and DNS unreverted.

## Layout

```
cmd/keqrnel/     entry point
service/         New / Start / Close
core/            Box assembly, curated registries
core/localdns/   the "local" DNS transport sing-box requires as a fallback
bridge/xray/     embedded xray engine + the "xray" outbound
```

[ARCHITECTURE.md](ARCHITECTURE.md) — how the bridge works and why the registries are
curated. [ERROR-ANALYSIS.md](ERROR-ANALYSIS.md) — hazards, trade-offs, limits.

## Versions

sing-box v1.13.13 · xray-core v1.260327.0 · Go 1.26+

## License

[GPL-3.0-or-later](LICENSE). Links [sing-box](https://github.com/SagerNet/sing-box)
(GPL-3.0-or-later) and embeds [xray-core](https://github.com/XTLS/Xray-core)
(MPL-2.0). Not affiliated with either project.

---

<h2 id="русский">Русский</h2>

<p align="center">
  <a href="#keqrnel">English</a> · <strong>Русский</strong>
</p>

Единое прокси-ядро: **host на sing-box со встроенным движком xray**. Один процесс
вместо привычной пары `xray → sing-box`.

### Что это

sing-box — хост: TUN, роутинг, DNS, sniffing, счётчики трафика через `clash_api`.
xray-core влинкован как библиотека и обслуживает один тип аутбаунда, `"xray"`:
VLESS, XHTTP, Reality, Vision, Mux, Hysteria2 и всё прочее, что принимает его
JSON-конфиг. Протоколы исполняет сам xray, поэтому на проводе ровно то, чего ждёт
сервер, настроенный на свежий xray.

Ни петли через локальный SOCKS между движками, ни второго процесса, ни порта,
который надо согласовывать.

Написано для [keqdroid/KEQDIS](https://github.com/Lemonochka/keqdroid), но от
приложения не зависит — запустит что угодно, умеющее написать sing-box-конфиг.

### Аутбаунд `xray`

```json
{
  "type": "xray",
  "tag": "proxy",
  "xray": {
    "outbounds": [
      { "protocol": "vless", "settings": { "vnext": [ ... ] }, "streamSettings": { ... } }
    ]
  }
}
```

Поле `xray` — сырой фрагмент xray-конфига (как минимум массив `outbounds`), который
отдаётся встроенному движку как есть; на этом и держится совместимость с вашими
серверами. Всё вокруг — обычный конфиг sing-box. Полный пример с TUN и роутингом:
[config.example.json](config.example.json).

Hysteria2 не требует ничего особенного — это такой же xray-аутбаунд.

Аутбаунды `socks` и `http` из sing-box тоже зарегистрированы, так что keqrnel годится
и как TUN-фронтенд к локальному прокси-аплинку — так работает AmneziaWG: `wireproxy`
поднимает локальный SOCKS, keqrnel заворачивает его в TUN.

### Сборка

```sh
go build ./...
go test ./...
go build -o keqrnel ./cmd/keqrnel
```

Нужен Go 1.26+ (требование xray-core). Нативно собирается под Windows и Linux,
кросс-компилируется под `linux/arm64` и `darwin/arm64`.

> Хелперы `include.*` из sing-box намеренно не используются: они регистрируют всё,
> что sing-box вообще умеет (tor, naive, ssh, dhcp-dns и `dns/transport/local`,
> который не компилируется под Windows). Вместо них курированный набор registry
> собран в [core/registry.go](core/registry.go).

### Запуск

```sh
keqrnel config.json     # по умолчанию ./config.json
keqrnel -c config.json  # флаги в стиле xray тоже принимаются, для drop-in запуска
```

TUN-инбаунду нужны права администратора/root (и `wintun.dll` рядом с бинарём на
Windows), SOCKS/HTTP-инбаунду — нет.

Процесс останавливается по SIGINT/SIGTERM и при закрытии своего stdin-пайпа. Второй
путь нужен, чтобы лаунчер мог попросить корректное завершение: на Windows жёсткий
`TerminateProcess` оставляет TUN-адаптер, маршруты и DNS неоткаченными.

### Структура

```
cmd/keqrnel/     точка входа
service/         New / Start / Close
core/            сборка Box, курированные registry
core/localdns/   транспорт "local" для DNS, который sing-box требует как фолбэк
bridge/xray/     встроенный движок xray + аутбаунд "xray"
```

[ARCHITECTURE.md](ARCHITECTURE.md) — как устроен мост и зачем курированные registry.
[ERROR-ANALYSIS.md](ERROR-ANALYSIS.md) — грабли, компромиссы, ограничения.

### Версии

sing-box v1.13.13 · xray-core v1.260327.0 · Go 1.26+

### Лицензия

[GPL-3.0-or-later](LICENSE). Линкует [sing-box](https://github.com/SagerNet/sing-box)
(GPL-3.0-or-later) и встраивает [xray-core](https://github.com/XTLS/Xray-core)
(MPL-2.0). С этими проектами не аффилирован.
