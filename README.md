# Burch Motorsport

Личный пит-уолл: телеметрия, трансляции и коммунити Формулы 1.
Сайт: [burchmotorsport.ru](https://burchmotorsport.ru)

## Что умеет

- **Телеметрия** `/live` — тайминг любой сессии с 2023 года: позиции, отрывы, сектора, шины, карта трассы с машинами, погода, race control, штрафы и лимиты трассы, радио команд, графики позиций по кругам и стратегии шин, прогноз личного зачёта и Кубка конструкторов «если ничего не изменится». Будущие сессии показывают обратный отсчёт и прошлогоднюю гонку на той же трассе.
- **Трансляция** `/watch` — видео с RuTube и телеметрия на одном экране, рядом или поверх картинки, с полноэкранным режимом. Сессия определяется по названию видео.
- **Главная** — ближайшая сессия с расписанием уикенда по МСК и подиум последней.
- **Пасхалка** — восемь тапов по логотипу. Ave Dominus Nox.

## Стек

| Слой | Технологии |
| --- | --- |
| Фронт | Vue 3.5, TypeScript, Vite 8, Vue Router, Pinia, VueUse, Lucide. Своя дизайн-система в духе необрутализма, без UI-китов |
| Бэк | Go 1.25, только стандартная библиотека. Кеширующий прокси над [OpenF1](https://openf1.org) |
| Инфра | Docker Compose, Caddy с авто-TLS, GitHub Actions |

Данные берутся из OpenF1. Бэк единственный, кто в него ходит: каждый эндпоинт кешируется со своим TTL, завершённые сессии навсегда, так что число зрителей на лимиты OpenF1 не влияет.

## Структура

```
api/    Go-сервис: /api/live, /api/track, /api/meetings, /api/sessions,
        /api/next, /api/resolve, /api/radio, /api/health
web/    Vue-приложение по Feature-Sliced Design:
        src/app       точка входа, роутер, SEO
        src/pages     home, live, watch, design-system
        src/widgets   header, next-race, live-panel
        src/features  theme-toggle, session-picker, rutube-player,
                      theater-mode, night-lords
        src/entities  session (загрузка сессии, карта трассы, плашка пилота)
        src/shared    ui (Bm*-компоненты), api, lib, assets/brand
```

Правила вёрстки и токены: `web/brand/BRAND_FRONTEND.md`, `web/src/shared/assets/brand/`.

## Запуск

Нужны Docker и Node 22+ с pnpm.

```bash
# бэк в контейнере, порт 8090
docker compose up -d api

# фронт с hot reload, /api проксируется в бэк
cd web && pnpm install && pnpm dev
```

Весь стек как на проде, через Caddy на `http://localhost`:

```bash
docker compose up -d --build
```

Проверки:

```bash
cd web && pnpm type-check && pnpm vitest run
docker run --rm -v "$PWD/api:/src" -w /src golang:1.25-alpine go test ./...
```

## Переменные окружения

Файл `.env` в корне, читается Docker Compose. В git не попадает.

| Переменная | Зачем |
| --- | --- |
| `SITE_ADDRESS` | Домены для Caddy, например `burchmotorsport.ru, www.burchmotorsport.ru`. Без неё сайт на `:80` без TLS |
| `OPENF1_USERNAME`, `OPENF1_PASSWORD` | Спонсорский аккаунт OpenF1. Без него работает только архив: у бесплатного тарифа нет живых данных и лимит 30 запросов в минуту |
| `RADIO_RELAY` | Адрес второго инстанса API в регионе, откуда доступен CDN F1, если основной сервер под гео-блоком |

`GET /api/health` показывает тариф OpenF1 и статус авторизации.

## Деплой

Пуш в `main` запускает GitHub Actions: изменения в `api/` пересобирают только бэк, в `web/` только фронт. Воркфлоу заходят на сервер по SSH, делают `git pull` и `docker compose up -d --build <сервис>`. Секреты репозитория: `SSH_HOST`, `SSH_USER`, `SSH_KEY`.

Первый запуск на чистом сервере:

```bash
curl -fsSL https://get.docker.com | sh
git clone https://github.com/IlyaBurch/burchmotorsport /opt/burchmotorsport
cd /opt/burchmotorsport
printf 'SITE_ADDRESS=burchmotorsport.ru, www.burchmotorsport.ru\n' > .env
docker compose up -d --build
```
