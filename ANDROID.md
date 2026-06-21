# keqrnel на Android (keqdroid)

Ядро (`service.Instance`) — обычный Go и полностью тестируется на десктопе.
Для Android не хватает только платформенного слоя VpnService.

## Что нужно от платформы

TUN на Android нельзя открыть из Go напрямую: fd выдаёт `VpnService`, а исходящие
сокеты нужно защищать (`VpnService.protect`), иначе они уйдут в петлю TUN. В
sing-box это закрывается интерфейсом `adapter.PlatformInterface`:

```go
service.New(configJSON, service.Options{ Platform: platform })
```

`service.New` инжектит платформу в контекст (`service.ContextWith[
adapter.PlatformInterface]`), и `box.New` подхватывает её для TUN/монитора
интерфейсов/защиты сокетов. Десктоп: `Platform: nil`.

## Почему нельзя просто `gomobile bind` на наш `service`

`adapter.PlatformInterface` содержит Go-типы (`tun.Tun`, `*tun.Options`,
`tun.DefaultInterfaceMonitor`) — gomobile их не байндит. Поэтому sing-box даёт
`experimental/libbox` c **упрощённым** интерфейсом (`libbox.PlatformInterface`,
`libbox.TunInterface`), который адаптируется к `adapter.PlatformInterface`
внутри. keqdroid (Kotlin) реализует именно упрощённый интерфейс.

## Рекомендуемый путь интеграции

1. **gomobile-биндинг.** Пакет `mobile` (Kotlin-facing), экспортит:
   `Start(configJSON string, platform PlatformInterface) error`, `Stop()`.
   Внутри — адаптер `libbox.PlatformInterface` → `adapter.PlatformInterface`,
   затем `service.New(..., service.Options{Platform: adapted})`.
   Переиспользовать адаптер из `libbox` (он уже написан и проверен).
2. **Сборка:** `gomobile bind -target=android -androidapi 21 ./mobile`
   (нужны Android SDK + NDK, `gomobile init`). Получаем `.aar` для keqdroid.
3. **Kotlin-сторона:** `VpnService` отдаёт tun fd и `protect()` через
   реализацию `PlatformInterface`; конфиг строится существующим парсером
   keqdroid (включая шер-линк → xray-фрагмент в поле `xray`).

## Что уже готово к этому

- `service.New(config, Options{Platform})` — сейм принят, компилируется,
  кросс-собирается под linux/arm64 (Android-арх).
- Курированные registry + свой `local` DNS — ровно та же причина, по которой
  libbox не использует дефолтный `dns/transport/local`.

## Что осталось (требует устройства/NDK, здесь не собирается)

- Пакет `mobile` с адаптером `libbox.PlatformInterface`.
- `gomobile bind` → `.aar`.
- Прогон на устройстве: VpnService fd, защита сокетов, реальный interop
  Reality/XHTTP/Hysteria2 с боевыми серверами.
