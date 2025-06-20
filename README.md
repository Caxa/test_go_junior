# 👤 People Enrichment API

REST-сервис на Go для приёма ФИО, обогащения через внешние API (возраст, пол, национальность) и сохранения в PostgreSQL.

---

## 🚀 Запуск проекта


go run main.go
Сервер поднимается по адресу: http://localhost:8086

🧩 Используемые API для обогащения



Возраст	agify.io

Пол	genderize.io

Национальность	nationalize.io

## 🛠 Управление проектом через Makefile
Проект использует Makefile для автоматизации часто используемых команд:

### 🧩 Основные команды

make run       - Запуск приложения

make test      - Запуск тестов

make lint      - Проверка кода линтером

make build     - Сборка проекта

make clean     - Очистка артефактов сборки

## 🗃 Работа с миграциями

make migrate-up          - Применить все новые миграции

make migrate-down        - Откатить последнюю миграцию

make migrate-status      - Показать текущую версию миграции

make migrate-force-reset - Принудительный сброс миграций (опасно!)

## Примеры использования миграций:


### Применить миграции
$ make migrate-up
> migrate -path db/migrations -database "postgres://user:pass@localhost:5432/db" up


### Откатить одну миграцию
$ make migrate-down
> migrate -path db/migrations -database "postgres://user:pass@localhost:5432/db" down 1

### Проверить статус
$ make migrate-status
> migrate -path db/migrations -database "postgres://user:pass@localhost:5432/db" version


## 🧪 Примеры REST-запросов 

### ✅ Добавление нового человека
POST /people

Пример запроса:

curl -X POST http://localhost:8086/people \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dmitriy",
    "surname": "Ushakov",
    "patronymic": "Vasilevich"
  }'


![Alt text](image.png)

### 📄 Получение списка людей
GET /people

http://localhost:8086/people?name=Ivan&limit=2&offset=0

![Alt text](image-2.png)
---

### ✏️ Обновление данных человека
PUT /people/:id

Пример запроса:
curl -X PUT http://localhost:8086/people/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ivan",
    "surname": "Petrov",
    "patronymic": "Sergeevich"
  }'
![Alt text](image-3.png)

---

### ❌ Удаление человека
DELETE /people/:id

Пример запроса:
curl -X DELETE http://localhost:8086/people/1
📸 Скриншот Postman: удаление
![Alt text](image-4.png)

---
## ⚙️ Переменные окружения .env



DB_HOST=localhost

DB_PORT=5432

DB_USER=postgres

DB_PASSWORD=postgres

DB_NAME=people


## 🧪 Тестирование

make test
![Alt text](image-13.png)

📸 Галерея скриншотов 

✅ POST /people — добавление нового человека

✅ GET /people — получение списка с фильтрами и пагинацией

✅ PUT /people/:id — изменение данных

✅ DELETE /people/:id — удаление записи

✅ Обогащённый JSON-ответ

✅ Скриншот из Postman или cURL

---
## 📚 Swagger-документация
Swagger будет доступен по адресу:


http://localhost:8086/swagger/index.html
![Alt text](image-6.png)



## 🗄️ База данных

📦 Проект использует PostgreSQL для хранения обогащённых данных о людях.

### 🔍 Структура таблицы

Ниже представлены скриншоты структуры базы данных, созданной с помощью миграций:

📸 _Скриншот 1: Таблица `people` — общая структура_
![Таблица People](image-10.png)

📸 _Скриншот 2: Типы и ограничения колонок_
![Типы данных](image-11.png)

📸 _Скриншот 3: Представление записей в таблице_
![Данные](image-12.png)


## 📌 Автор

Контакты: airat.sharushev@gmail.com
тг @AiratSharushev