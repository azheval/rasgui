# rasgui

[![Build](https://github.com/azheval/rasgui/actions/workflows/build.yml/badge.svg)](https://github.com/azheval/rasgui/actions/workflows/build.yml)

## English

`rasgui` is a portable cross-platform web application for administering 1C server infrastructure through `rac` and `ras`.

It provides:

- web GUI for clusters, infobases, sessions, connections, and related administrative operations;
- flexible RBAC with roles, scoped permissions, and UI visibility based on access rights;
- support for remote `RAS` hosts with arbitrary ports;
- support for multiple local `rac` / `ras` versions through toolchain profiles;
- favorites, guided operator flows, audit log, and multilingual UI (`en`, `be`, `ru`).

Quick start:

```powershell
go run ./cmd/rasgui
```

Open:

- `http://localhost:8099`

Default account:

- login: `admin`
- password: `admin123`

## Беларуская

`rasgui` — гэта пераноснае кросплатформеннае вэб-прыкладанне для адміністравання сервернай інфраструктуры 1С праз `rac` і `ras`.

Прыкладанне дае:

- вэб-інтэрфейс для кластараў, інфабаз, сеансаў, злучэнняў і іншых адміністрацыйных дзеянняў;
- гнуткую RBAC-мадэль з ролямі, scope-правамі і схаваннем недаступных элементаў інтэрфейсу;
- падтрымку аддаленых `RAS` з адвольным хостам і портам;
- падтрымку некалькіх лакальных версій `rac` / `ras` праз профілі toolchain;
- абранае, аператарскія сцэнарыі, аўдыт і інтэрфейс на `en`, `be`, `ru`.

Хуткі запуск:

```powershell
go run ./cmd/rasgui
```

Адкрыць:

- `http://localhost:8099`

Пачатковы ўліковы запіс:

- лагін: `admin`
- пароль: `admin123`

## Русский

`rasgui` — это переносимое кроссплатформенное веб-приложение для администрирования серверной инфраструктуры 1С через `rac` и `ras`.

Приложение предоставляет:

- web GUI для кластеров, инфобаз, сеансов, соединений и других административных операций;
- гибкую RBAC-модель с ролями, scope-правами и скрытием недоступных элементов интерфейса;
- поддержку удаленных `RAS` с произвольным хостом и портом;
- поддержку нескольких локальных версий `rac` / `ras` через профили toolchain;
- избранное, операторские сценарии, аудит и интерфейс на `en`, `be`, `ru`.

Быстрый запуск:

```powershell
go run ./cmd/rasgui
```

Открыть:

- `http://localhost:8099`

Начальная учетная запись:

- логин: `admin`
- пароль: `admin123`
