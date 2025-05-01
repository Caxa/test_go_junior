# 👤 People Enrichment API

REST-сервис на Go для приёма ФИО, обогащения через внешние API (возраст, пол, национальность) и сохранения в PostgreSQL.

---

## 🚀 Запуск проекта

```bash
go run main.go
Сервер поднимается по адресу: http://localhost:8086

🧩 Используемые API для обогащения

Тип	Сервис
Возраст	agify.io
Пол	genderize.io
Национальность	nationalize.io
🧪 Примеры REST-запросов
✅ Добавление нового человека
POST /people

Пример запроса:

bash
Копировать
Редактировать
curl -X POST http://localhost:8086/people \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dmitriy",
    "surname": "Ushakov",
    "patronymic": "Vasilevich"
  }'
📸 Скриншот Postman: добавление нового человека

📄 Получение списка людей
GET /people

Пример базового запроса:

bash
Копировать
Редактировать
curl http://localhost:8086/people
Пример с фильтрами:

bash
Копировать
Редактировать
curl "http://localhost:8086/people?gender=male&nationality=RU&age=30"
Пример с пагинацией:

bash
Копировать
Редактировать
curl "http://localhost:8086/people?page=2&limit=10"
📸 Скриншот Postman: получение списка

✏️ Обновление данных человека
PUT /people/:id

Пример запроса:

bash
Копировать
Редактировать
curl -X PUT http://localhost:8086/people/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ivan",
    "surname": "Petrov",
    "patronymic": "Sergeevich"
  }'
📸 Скриншот Postman: изменение данных

❌ Удаление человека
DELETE /people/:id

Пример запроса:

bash
Копировать
Редактировать
curl -X DELETE http://localhost:8086/people/1
📸 Скриншот Postman: удаление

🧾 Пример структуры данных
JSON-запрос для добавления:

json
Копировать
Редактировать
{
  "name": "Dmitriy",
  "surname": "Ushakov",
  "patronymic": "Vasilevich" // необязательное поле
}
⚙️ Переменные окружения .env
env
Копировать
Редактировать
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=people
🧪 Тестирование
bash
Копировать
Редактировать
go test ./...
📸 Скриншот: прохождение тестов

📸 Галерея скриншотов (вставьте вручную)
✅ POST /people — добавление нового человека

✅ GET /people — получение списка с фильтрами и пагинацией

✅ PUT /people/:id — изменение данных

✅ DELETE /people/:id — удаление записи

✅ Обогащённый JSON-ответ

✅ Скриншот из Postman или cURL

📚 Swagger-документация
Swagger будет доступен по адресу:

bash
Копировать
Редактировать
http://localhost:8086/swagger/index.html
📸 Скриншот Swagger UI

📌 Автор
Разработано в рамках учебного проекта.
Контакты: email@example.com

scss
Копировать
Редактировать

Хочешь, чтобы я подготовил базовые скриншоты-заглушки (например, как выглядят запросы в Postman), чтобы ты мог сразу вставить их в проект?






Вы сказали:
всё это в маркдаун выдай пожалуйста


ChatGPT сказал:
Вот файл README.md в формате Markdown, как ты просил:

markdown
Копировать
Редактировать
# 👤 People Enrichment API

REST-сервис на Go для приёма ФИО, обогащения через внешние API (возраст, пол, национальность) и сохранения в PostgreSQL.

---

## 🚀 Запуск проекта

```bash
go run main.go
Сервер поднимается по адресу: http://localhost:8086

🧩 Используемые API для обогащения

Тип	Сервис
Возраст	agify.io
Пол	genderize.io
Национальность	nationalize.io
🧪 Примеры REST-запросов
✅ Добавление нового человека
POST /people

Пример запроса:

bash
Копировать
Редактировать
curl -X POST http://localhost:8086/people \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Dmitriy",
    "surname": "Ushakov",
    "patronymic": "Vasilevich"
  }'
📸 Скриншот Postman: добавление нового человека

📄 Получение списка людей
GET /people

Пример базового запроса:

bash
Копировать
Редактировать
curl http://localhost:8086/people
Пример с фильтрами:

bash
Копировать
Редактировать
curl "http://localhost:8086/people?gender=male&nationality=RU&age=30"
Пример с пагинацией:

bash
Копировать
Редактировать
curl "http://localhost:8086/people?page=2&limit=10"
📸 Скриншот Postman: получение списка

✏️ Обновление данных человека
PUT /people/:id

Пример запроса:

bash
Копировать
Редактировать
curl -X PUT http://localhost:8086/people/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Ivan",
    "surname": "Petrov",
    "patronymic": "Sergeevich"
  }'
📸 Скриншот Postman: изменение данных

❌ Удаление человека
DELETE /people/:id

Пример запроса:

bash
Копировать
Редактировать
curl -X DELETE http://localhost:8086/people/1
📸 Скриншот Postman: удаление

🧾 Пример структуры данных
JSON-запрос для добавления:

json
Копировать
Редактировать
{
  "name": "Dmitriy",
  "surname": "Ushakov",
  "patronymic": "Vasilevich" // необязательное поле
}
⚙️ Переменные окружения .env
env
Копировать
Редактировать
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=people
🧪 Тестирование
bash
Копировать
Редактировать
go test ./...
📸 Скриншот: прохождение тестов

📸 Галерея скриншотов (вставьте вручную)
✅ POST /people — добавление нового человека

✅ GET /people — получение списка с фильтрами и пагинацией

✅ PUT /people/:id — изменение данных

✅ DELETE /people/:id — удаление записи

✅ Обогащённый JSON-ответ

✅ Скриншот из Postman или cURL

📚 Swagger-документация
Swagger будет доступен по адресу:

bash
Копировать
Редактировать
http://localhost:8086/swagger/index.html
📸 Скриншот Swagger UI

📌 Автор
Разработано в рамках учебного проекта.
Контакты: email@example.com

