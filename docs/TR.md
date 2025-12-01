# Техническое задание

Всего 5 сервисов:

1. [ ] API Gateway (REST от Клиента -> gRPC)
2. [ ] Telemetry Service (gRPC от Дрона -> Kafka(Producer))
3. [ ] Tracking Service (Redis, гео-поиск ближайшего)
4. [ ] Order Service (Database, бизнес-логика)
5. [ ] Dispatch Service (Диспетчер, оркестратор)

6. [ ] Drone Emulator (optional)

По инфраструктуре:

1. [ ] Nginx
2. [ ] Docker Compose (общий для всего проекта)
3. [ ] Dockerfile (свой для каждого сервиса)
4. [ ] Gitlab CI/CD (сборка, линтер, тесты и т.д.)

## Telemetry Service

Для каждого дрона - свое соединение для обмена данными
Нужно:

1. [ ] Реализовать `sync.Map` где хранятся данные вида `{ID: Connection}`, добавлять в эту мапу id дрона и его коннект
   при первом соединении
2. [ ] Принимать данные от дрона
    * Валидировать их: проверять что батарея в пределах `0-100`, координаты в пределах допустимой территории, статус
      должен быть одним из описанных в .proto контракте, иначе обрывать соединение
    * Отправлять в Kafka топик `telemetry.raw`данные от дрона
3. [ ] Реализовать gRPC метод SendCommand (вызывается Dispatch сервисом):
    * Найти соединение в `sync.Map` по id дрона
    * Отправить эту команду дрону
4. [ ] Реализовать отправку событий от дрона `ARRIVED`, `PICKED_UP`, `DROPPED_CARGO` в Kafka топик `events`

[Proto контракт](../proto/telemetry.proto)

## Tracking Service

Отвечает за хранение данных дронов в Redis (про Redis Geo можно почитать [тут](https://habr.com/ru/articles/679994/))
Нужно:

1. [ ] Подключиться к Kafka Consumer Group `tracking-group`
2. [ ] Читать топик `telemetry.raw`:
    * Обновлять гео-индекс: `GEOADD drones <lon> <lat> <drone_id>`
    * Обновлять статус: `HSET drone:<id> battery <val> status <val>`
3. [ ] Реализовать gRPC метод `FindNearest`:
    * Принимать координаты склада
    * Делать `GEOSEARCH ... BYRADIUS 30 km` в Redis, фильтровать по статусу `FREE` и заряду `>20%`, сортировать по `ASC`
    * Возвращать ID ближайшего дрона (первый после сортировки)
4. [ ] Реализовать gRPC метод `GetDroneLocation` (возвращать из Redis данные дрона)
5. [ ] Реализовать gRPC метод `SetStatus` (менять статус дрона в Redis)

[Proto контракт](../proto/tracking.proto)

## Order Service

Отвечает за создание заказа и получения информации по нему
Нужно:

1. [ ] Поднять миграции БД (таблица `orders`)
2. [ ] Реализовать gRPC метод `CreateOrder`:
    * Сохранить заказ в БД со статусом `PENDING`
    * Вызвать `dispatch.AssignDrone`
    * Обновить заказ (`ASSIGNED` в случае успешного ответа от `dispatch.AssignDrone` и `FAILED` в ином случае)
3. [ ] Реализовать gRPC метод `GetOrder` (возвращать данные заказа по его ID)
4. [ ] Реализовать gRPC метод `UpdateStatus` (обновлять статус заказа, метод вызывает `dispatch` сервис)

[Proto контракт](../proto/order.proto)

## Dispatch Service

Привязывает дрона к заказу и отправляет ему команды (через `telemetry` сервис)
Нужно:

1. [ ] Реализовать метод `AssignOrder`:
    * Найти ближайший склад к точке заказа
    * Вызвать `tracking.FindNearest` для получения ближайшего дрона к этому складу
    * Вызвать `tracking.SetStatus(BUSY)` если дрон найден, иначе вернуть `success: false`
    * Вызвать `telemetry.SendCommand(FLY_TO(Store))`
2. [ ] Читать сообщения из Kafka топика `events`:
    * Если `ARRIVED_AT_STORE` -> шлем `PICKUP_CARGO`
    * Если `PICKED_UP_CARGO` -> шлем `FLY_TO(Client)`
    * Если `ARRIVED_AT_CLIENT` -> шлем `DROP_CARGO`
    * Если `DROPPED_CARGO` -> шлем `FLY_TO(Base)`, `order.UpdateStatus(COMPLETED)` и `tracking.SetStatus(FREE)`
    * Если `ARRIVED_AT_BASE` -> шлем `CHARGE` и `tracking.SetStatus(CHARGING)` если `battery < 80%`, иначе просто
      `WAIT` (статус остается `FREE`)
    * Если `FULLY_CHARGED` -> шлем `WAIT` и `tracking.SetStatus(FREE)`

Proto контракт:

[Proto контракт](../proto/dispatch.proto)

## API Gateway

Сервер на `echo`, принимает REST HTTP запросы от клиентского приложения (фронтенда), не содержит бизнес-логики,
занимается только маршрутизацией и базовой валидации
Нужно:

1. [ ] Поднять HTTP сервер на порту `8080`
2. [ ] POST `/api/v1/orders`:
    * Принимает JSON:
   ```json
       {
           "user_id": "<uuid>", 
           "items": ["bread", "milk", ..], 
           "delivery_location": {"lat": 55.7531753, "lon": 37.6104283}
       }
   ```
    * Валидирует данные (есть ли координаты, юзер с таким id, не пустой ли список)
    * Делает gRPC запрос `order.CreateOrder`
    * Возвращает JSON `201 Created`:
   ```json
   {
       "order_id": "<uuid>",
       "status": "PENDING" | "ASSIGNED",
       "estimated_time": "15 min"
   }
   ```
3. [ ] GET `/api/v1/orders`:
    * Делает gRPC запрос `order.GetOrder`
    * Если есть `drone_id` делает `tracking.GetDroneLocation`
    * Собирает агрегированный JSON ответ:
   ```json
   {
       "order_id": "<uuid>",
       "status": "ASSIGNED",
       "drone": {
           "id": "<uuid>",
           "location": {
               "lat": 55.7289473,
               "lon": 37.7457302
           },
           "battery": 85
       }
   }
   ```

*В будущем добавить регистрацию/авторизацию пользователя и CRUD в БД*

[OpenAPI спецификация](../services/api-gateway/openapi.yaml)

## Drone Emulator (отдельный клиентский скрипт, не относится к сервисам)

*Отличие эмулятора от симулятора в том, что эмулятор воспроизводит интерфейс и поведение системы, а симулятор точную
физику и поведение дрона в реальном мире.
В MVP нашего проекта используем **эмулятор**, потому что он не учитывает гравитацию, погодные условия и препятствия. Ему
важно, чтобы на команду "Лети туда" он начал менять координаты с заданной скоростью*

Запускает N "виртуальных дронов", подключается с каждого дрона (отдельной функции) к Telemetry Service
Нужно:

1. [ ] Для каждого дрона запустить отдельную горутину
2. [ ] Логика одного дрона:
    * Инициализация:
        * Сгенерировать UUID
        * Выбрать случайную стартовую точку внутри полигона
        * Установить `Battery 100%`, `Status = FREE`
    * Соединение:
        * Установить gRPC Stream соединение с `Telemetry Service` (метод `Link`)
    * Цикл жизни (Tick Loop):
        * Раз в N миллисекунд:
            * Если статус `BUSY` (летим):
                * Изменить `Lat, Lon` в сторону `Target` с фиксированной скоростью (например 10 м/с)
                * Уменьшить заряд батареи (например -1% раз в 15 секунд)
                * Проверить: если расстояние до цели <5 метров (учет погрешности) -> отправить событие (`ARRIVED`,
                  `PICKED_UP`, `DROPPED`)
            * Если статус `CHARGING` (на базе на зарядке)
                * Увеличить заряд (+1% раз в 5 секунд)
                * Если 100% -> отправить `FULLY_CHARGED`, статус `FREE`
            * Отправка телеметрии:
                * Отправить текущие `Lat, Lon, Battery, Status, Event` в стрим `stream.Send()`
                * Сбросить `EVENT=NONE` после отправки (ждем команды)
    * Обработка команд (Recv Loop):
        * В параллельной горутине читать `stream.Recv()`
        * Команда `FLY_TO`:
            * Запомнить `Target`
            * Обновить внутренний `TargetType`
            * Поставить статус = `BUSY`
        * Команда `PICKUP_CARGO` / `DROP_CARGO`
            * Заснуть на 3 секунды (имитация работы)
            * Выставить флаг для отправки `PICKED_UP` / `DROPPED` события в следующем тике
        * Команда `CHARGE`
            * Поставить статус = `CHARGING`

Алгоритм движения (примерный):

```go
deltaLat := (targetLat - currentLat) * speedFactor
deltaLon := (targetLon - currentLon) * speedFactor
currentLat += deltaLat
currentLon += deltaLon
```

*В будущем можно прикрутить еще один скрипт, который отображает дронов на карте.
Данные берутся из Tracking Service*
