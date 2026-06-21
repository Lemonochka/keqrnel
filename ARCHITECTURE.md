# keqrnel — единое ядро

Цель: одно Go-ядро вместо цепочки `xray → sing-box` и отдельных ядер.
Берём лучшее из каждого, ничего не пишем с нуля.

## Принцип

**Host = sing-box, протокол-движок = встроенный xray.**

- sing-box несёт: TUN (`sing-tun`), роутинг, DNS, нативный Hysteria2,
  Android-биндинг (`libbox`).
- xray-core встраивается как библиотека и обслуживает **один** bridge-outbound:
  VLESS + XHTTP + Reality + Vision + Mux — ровно теми реализациями, что в xray,
  чтобы был байт-в-байт паритет с серверами (все на свежем xray).

```
TUN (sing-tun) ─► routing ─► outbounds:
                    ├─ hysteria2      (sing-box native)
                    └─ xray-bridge ─► embedded xray.Instance
                                      VLESS/XHTTP/Reality/Vision/Mux
```

Один процесс, один TUN, без loopback-SOCKS между движками.

## Bridge-механизм

xray-core экспортит `core.Dial(ctx, instance, dest) (net.Conn, error)` — дилит
через свой роутинг/outbound. Кастомный sing-box outbound в `DialContext`
зовёт `core.Dial` встроенного инстанса. Чистый `net.Conn`, без петлевания.

## Конфиг

Общий sing-box-config. У xray-bridge outbound — **вложенный сырой xray-JSON
фрагмент**, скармливается встроенному xray как есть. Свой outbound из
xray-конфига вставляется 1:1 → гарантия совместимости с серверами.

## Скоуп

| Берём | Откуда | Статус |
|---|---|---|
| TUN, роутинг, DNS, Hysteria2, libbox | sing-box | host |
| VLESS/XHTTP/Reality/Vision/Mux | xray-core (as lib) | bridge |
| ~~AmneziaWG~~ | — | **остаётся отдельным ядром, не трогаем** |

## Роадмап

0. Форк/сборка sing-box-базы + `libbox`, TUN живой.
1. **embedded xray + bridge-outbound** (обязательная база). ← старт
2. Hysteria2 в общий роутинг/конфиг.
3. Единая JSON-схема, выпиливание легаси.

## Структура пакетов (план)

```
cmd/keqrnel/    desktop-entry для тестов
core/           оркестрация host'а (sing-box instance)
bridge/xray/    embedded xray + bridge-outbound (core.Dial)
config/         единая схема + xray-фрагмент
```
