# keqrnel — анализ ошибок и риски

Глубокий разбор по реальным исходникам sing-box v1.13.13 / xray-core v1.260327.0.
Статусы: **FIXED** (исправлено), **OK** (проверено, корректно), **ACCEPTED**
(осознанный компромисс), **TODO** (требует внимания на следующих этапах).

## Жизненный цикл / ресурсы

- **FIXED — утечка xray-инстанса при частичном фейле `box.New`.**
  Движок стартовал в `New()` (во время конструирования Box). Если другой
  outbound падал позже, наш `Close()` не вызывался → инстанс утекал. Особенно
  опасно на Android (частые релоады). Решение: bridge-outbound реализует
  `adapter.Lifecycle` — движок стартует в `Start(StartStateStart)`, закрывается
  в `Close()`. Парность гарантирована. См. [bridge/xray/outbound.go](bridge/xray/outbound.go).

- **OK — `Close()` действительно вызывается.** `outbound.Manager.Close()`
  закрывает каждый outbound через `io.Closer`; при замене/удалении (релоад) —
  `common.Close(existsOutbound)`. Наш `Close() error` подходит под оба пути.

- **FIXED — гонка движка при остановке.** Доступ к `engine` под `sync.RWMutex`;
  после `Close()` поле обнуляется, `currentEngine()` отдаёт явную ошибку вместо
  nil-разыменования при гонке dial/close.

- **ACCEPTED — до ~60 с висящая UDP-горутина после закрытия сессии.**
  xray `dispatcherConn` чистит связанный `connEntry` по таймеру неактивности
  (1 мин), а не мгновенно при `Close()`. Для UDP приемлемо; число горутин
  ограничено числом активных сессий.

## Сеть / корректность данных

- **OK — UDP-маршрутизация по назначению.** Каждый UDP-flow из TUN вызывает
  `ListenPacket` отдельно → свой `dispatcherConn`; `WriteTo` кладёт назначение
  в каждый пакет (`buffer.UDP`). Так как xray-инстанс у нас = один outbound
  (роутинг на стороне sing-box), мультиплекс на один линк проблемы не создаёт.

- **OK — типы соединений.** `xnet.Conn`/`xnet.PacketConn` у xray — алиасы
  stdlib `net.Conn`/`net.PacketConn`, совместимы с `N.Dialer` напрямую.

- **OK — адреса.** `host()` отдаёт FQDN либо строковый IP; `xnet.ParseAddress`
  корректно различает домен/IPv4/IPv6.

- **ACCEPTED — `DialContext` только TCP.** UDP через `DialContext` отклоняется
  явной ошибкой; весь UDP идёт штатным путём `ListenPacket`. Соответствует
  модели sing-box TUN NAT.

## Платформа / сборка

- **FIXED — sing-box v1.13.13 не собирается под Windows.** Его пин sing-tun
  переименовал `MyInterface()` → `MyInterfaces()`, а `dns/transport/local/
  resolv_windows.go` зовёт старое имя. Это windows-only файл (Android/Linux не
  затронуты). Решение: не использовать `include.*` (тянет весь легаси, включая
  этот пакет), а собирать курированные registry в [core/registry.go](core/registry.go).

- **FIXED — обязательный `local` DNS-транспорт.** sing-box на старте создаёт
  дефолтный DNS-fallback типа `local` (box.go ~327). Без регистрации:
  `default DNS server fallback: transport type not found: local`. Реализован
  свой кросс-платформенный транспорт на `net.Resolver`
  ([core/localdns/local.go](core/localdns/local.go)) — как делает libbox для Android.

- **OK — целевые платформы собираются.** Нативно Windows/amd64, кросс
  linux/amd64 и linux/arm64 (Android-арх) — все `go build ./...` = 0.

## Конфиг / поведение

- **OK — паритет с серверами.** Поле `xray` принимает сырой xray-конфиг 1:1;
  протоколы исполняет сам xray → байт-в-байт совпадение с серверами на свежем
  xray (VLESS/XHTTP/Reality/Vision/Mux/Hysteria2).

- **OK — ошибки конфига не паникуют.** Малформед-конфиг и пустой `xray`-фрагмент
  отдают ошибку (`TestInvalidConfigReports`, проверка в `New`).

- **TODO — реальный interop Reality/XHTTP/Hysteria2.** Проверено, что движок
  гоняет трафик (freedom round-trip + curl через socks→xray→интернет, HTTP 200).
  Полный interop с конкретными серверами требует прогона с боевым конфигом —
  невозможно здесь без её серверов. Риск только в трансляции шер-линк → xray
  JSON, которая переиспользуется из keqdroid как есть.

## Покрытие тестами

- `bridge/xray`: round-trip реального коннекта через `core.Dial`.
- `core`: парс example-конфига через курированные registry.
- `service`: полный New/Start/Close с роут-правилами; ошибка на битом конфиге.
- Рантайм-смоук: бинарник + `curl` через socks → xray-движок → интернет.
