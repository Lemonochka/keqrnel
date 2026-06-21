# keqrnel

Единое прокси-ядро для keqdroid: **host на sing-box + встроенный движок xray**.
Заменяет цепочку `xray → sing-box` и отдельные ядра одним бинарём.

- **sing-box** несёт: TUN, роутинг, DNS, sniffing, Android-биндинг (`libbox`).
- **xray-core** встроен как библиотека и обслуживает outbound типа `"xray"`:
  VLESS / XHTTP / Reality / Vision / Mux / **Hysteria2** — реализациями самого
  xray, поэтому поведение байт-в-байт совпадает с серверами на свежем xray.
- **AmneziaWG** сюда не входит — остаётся отдельным ядром (так и задумано).

Подробности — в [ARCHITECTURE.md](ARCHITECTURE.md).

## Как это устроено

```
TUN (sing-tun) ─► routing ─► outbound "proxy" (type: xray)
                                      │
                                      ▼
                              embedded xray.Instance
                              VLESS/XHTTP/Reality/Vision/Mux/Hysteria2
```

Outbound `{"type":"xray","xray":{ <сырой xray-конфиг> }}` поднимает встроенный
xray-инстанс и дислатчит через него трафик (`core.Dial`/`DialUDP`). Поле `xray`
— это обычный xray-конфиг (как минимум массив `outbounds`), вставляется 1:1.

## Сборка

Целевая платформа — **Android** (через gomobile/`libbox`). Локально:

```sh
go build ./...            # нативно (Windows/Linux/macOS)
go test  ./...            # тесты: round-trip через движок + парс конфига
go build -o keqrnel ./cmd/keqrnel
```

> Мы намеренно НЕ используем `include.*` из sing-box (он тащит весь легаси —
> tor, naive, ssh, dhcp-dns и сломанный на Windows `dns/transport/local`).
> Вместо этого в [core/registry.go](core/registry.go) собран курированный набор
> registry. Это и выкидывает легаси, и чинит нативную Windows-сборку.

## Запуск

```sh
keqrnel config.json       # по умолчанию ./config.json
```

TUN требует прав администратора/root. Пример конфига —
[config.example.json](config.example.json): TUN + роутинг + xray-outbound с
VLESS/XHTTP/Reality/Vision. Подставь свои `YOUR_*` значения.

## Hysteria2

Hysteria2 идёт через тот же встроенный движок — отдельный outbound не нужен.
Твой `hysteria2://…` шер-линк конвертируется в xray-outbound и кладётся в поле
`xray`. Пример формы xray-outbound для Hysteria2:

```json
{
  "outbounds": [
    {
      "protocol": "hysteria2",
      "settings": {
        "servers": [{ "address": "sub.example.art", "port": 9000 }],
        "auth": "YOUR_PASSWORD",
        "obfs": { "type": "salamander", "password": "YOUR_OBFS_PASSWORD" }
      },
      "streamSettings": {
        "network": "hysteria",
        "security": "tls",
        "tlsSettings": { "serverName": "sub.example.art", "alpn": ["h3"] }
      }
    }
  ]
}
```

Конкретная схема полей — по версии твоего xray; парсер шер-линков из keqdroid
переиспользуется как есть, bridge принимает уже готовый xray-фрагмент.

## Структура

```
cmd/keqrnel/     entry-point (desktop, для тестов)
core/            сборка Box + курированные registry
bridge/xray/     встроенный xray-движок + bridge-outbound (core.Dial)
config.example.json
```

## Версии

sing-box v1.13.13, xray-core v1.260327.0 (требует Go ≥ 1.26).
