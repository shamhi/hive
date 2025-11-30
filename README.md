## Общая архитектура
https://excalidraw.com/#json=aiozIsaqS5UeVLdOy6A-v,b71CCzkSmriyXZnau0n6Mg

Всего 6 компонентов:
1. Nginx (прокси для Telemetry Service и API Gateway)
2. API Gateway (REST от Клиента -> gRPC)
3. Telemetry Service (gRPC от Дрона -> Kafka(Producer) -> SendCommand)
4. Tracking Service (Redis, гео-поиск ближайшего)
5. Order Service (Database, бизнес-логика)
6. Dispatch Service (Диспетчер, оркестратор)


### Задачи по сервисам
#### Telemetry Service
Для каждого дрона - свое соединение для обмена данными
Нужно:
1. Реализовать `sync.Map` где хранятся данные вида `{ID: Connection}`, добавлять в эту мапу id дрона и его коннект при первом соединении
2. Принимать данные от дрона
    * Валидировать их: проверять что батарея в пределах `0-100`, координаты в пределах допустимой территории, статус должен быть одним из описанных в .proto контракте, иначе обрывать соединение
    * Отправлять в Kafka топик `telemetry.raw`данные от дрона
3. Реализовать gRPC метод SendCommand (вызывается Dispatch сервисом):
    * Найти соединение в `sync.Map` по id дрона
    * Отправить эту команду дрону
4. Реализовать отправку событий от дрона `ARRIVED`, `PICKED_UP`, `DROPPED_CARGO` в Kafka топик `events`

Proto контракт:
```proto
syntax = "proto3";  
  
package telemetry;  
option go_package = "hive/gen/telemetry";  
  
service TelemetryService {  
  rpc Link(stream DroneTelemetry) returns (stream ServerCommand);  
  rpc SendCommand(DispatchCommandRequest) returns (DispatchCommandResponse);  
}  
  
enum DroneStatus {  
  STATUS_UNKNOWN = 0;  
  STATUS_FREE = 1;  
  STATUS_BUSY = 2;  
  STATUS_CHARGING = 3;  
}  
  
enum DroneEvent {  
  EVENT_NONE = 0;  
  EVENT_ARRIVED_AT_STORE = 1;  
  EVENT_PICKED_UP_CARGO = 2;  
  EVENT_ARRIVED_AT_CLIENT = 3;  
  EVENT_DROPPED_CARGO = 4;  
  EVENT_ARRIVED_AT_BASE = 5;  
  EVENT_FULLY_CHARGED = 6;  
}  
  
enum Action {  
  ACTION_WAIT = 0;  
  ACTION_FLY_TO = 1;  
  ACTION_PICKUP_CARGO = 2;  
  ACTION_DROP_CARGO = 3;  
  ACTION_CHARGE = 4;  
}  
  
enum TargetType {  
  TARGET_POINT = 0;  
  TARGET_STORE = 1;  
  TARGET_CLIENT = 2;  
  TARGET_BASE = 3;  
}  
  
message Location {  
  double lat = 1;  
  double lon = 2;  
}  
  
message DroneTelemetry {  
  string drone_id = 1;  
  Location drone_location = 2;  
  int32 battery = 3;  
  DroneStatus status = 4;  
  int64 timestamp = 5;  
  DroneEvent event = 6;  
}  
  
message ServerCommand {  
  string command_id = 1;  
  Action action = 2;  
  Location target = 3;  
  TargetType type = 4;  
}  
  
message DispatchCommandRequest {  
  string drone_id = 1;  
  Action action = 2;  
  Location target = 3;  
  TargetType type = 4;  
}  
  
message DispatchCommandResponse {  
  bool success = 1;  
}
```


#### Tracking Service
Отвечает за хранение данных дронов в Redis
Нужно:
1. Подключиться к Kafka Consumer Group `tracking-group`
2. Читать топик `telemetry.raw`:
    * Обновлять гео-индекс: `GEOADD drones <lon> <lat> <drone_id>`
    * Обновлять статус: `HSET drone:<id> battery <val> status <val>`
3. Реализовать gRPC метод `FindNearest`:
    * Принимать координаты склада
    * Делать `GEOSEARCH ... BYRADIUS 30 km` в Redis, фильтровать по статусу `FREE` и заряду `>20%`, сортировать по `ASC`
    * Возвращать ID ближайшего дрона (первый после сортировки)
4. Реализовать gRPC метод `GetDroneLocation` (возвращать из Redis данные дрона)
5. Реализовать gRPC метод `SetStatus` (менять статус дрона в Redis)

Proto контракт:
```proto
syntax = "proto3";  
  
package tracking;  
option go_package = "hive/gen/tracking";  
  
service TrackingService {  
  rpc FindNearest(FindNearestRequest) returns (FindNearestResponse);  
  rpc GetDroneLocation(GetDroneLocationRequest) returns (GetDroneLocationResponse);  
  rpc SetStatus(SetStatusRequest) returns (SetStatusResponse);  
}  
  
enum DroneStatus {  
  STATUS_UNKNOWN = 0;  
  STATUS_FREE = 1;  
  STATUS_BUSY = 2;  
  STATUS_CHARGING = 3;  
}  
  
message Location {  
  double lat = 1;  
  double lon = 2;  
}  
  
message FindNearestRequest {  
  Location store_location = 1;  
}  
  
message FindNearestResponse {  
  string drone_id = 1;  
  bool found = 2;  
  double distance_meters = 3;  
}  
  
message GetDroneLocationRequest {  
  string drone_id = 1;  
}  
  
message GetDroneLocationResponse {  
  Location drone_location = 1;  
  int32 battery = 2;  
}  
  
message SetStatusRequest {  
  string drone_id = 1;  
  DroneStatus status = 2;  
}  
  
message SetStatusResponse {  
  bool success = 1;  
}
```


#### Order Service
Отвечает за создание заказа и получения информации по нему
Нужно:
1. Поднять миграции БД (таблица `orders`)
2. Реализовать gRPC метод `CreateOrder`:
    * Сохранить заказ в БД со статусом `PENDING`
    * Вызвать `dispatch.AssignDrone`
    * Обновить заказ (`ASSIGNED` в случае успешного ответа от `dispatch.AssignDrone` и `FAILED` в ином случае)
3. Реализовать gRPC метод `GetOrder` (возвращать данные заказа по его ID)
4. Реализовать gRPC метод `UpdateStatus` (обновлять статус заказа, метод вызывает `dispatch` сервис)

Proto контракт:
```proto
syntax = "proto3";  
  
package order;  
option go_package = "hive/gen/order";  
  
service OrderService {  
  rpc CreateOrder(CreateOrderRequest) returns (CreateOrderResponse);  
  rpc GetOrder(GetOrderRequest) returns (GetOrderResponse);  
  rpc UpdateStatus(UpdateStatusRequest) returns (UpdateStatusResponse);  
}  
  
enum OrderStatus {  
  CREATED = 0;  
  PENDING = 1;  
  ASSIGNED = 2;  
  COMPLETED = 3;  
  FAILED = 4;  
}  
  
message Location {  
  double lat = 1;  
  double lon = 2;  
}  
  
message CreateOrderRequest {  
  string user_id = 1;  
  repeated string items = 2;  
  Location delivery_location = 3;  
}  
  
message CreateOrderResponse {  
  string order_id = 1;  
  OrderStatus status = 2;  
  string estimated_time = 3;  
}  
  
message GetOrderRequest {  
  string order_id = 1;  
}  
  
message GetOrderResponse {  
  string order_id = 1;  
  OrderStatus status = 2;  
  string drone_id = 3;  
}  
  
message UpdateStatusRequest {  
  string order_id = 1;  
  OrderStatus new_status = 2;  
  string drone_id = 3;  
}  
  
message UpdateStatusResponse {  
  bool success = 1;  
}
```


#### Dispatch Service
Привязывает дрона к заказу и отправляет ему команды (через `telemetry` сервис)
Нужно:
1. Реализовать метод `AssignOrder`:
    * Найти ближайший склад к точке заказа
    * Вызвать `tracking.FindNearest` для получения ближайшего дрона к этому складу
    * Вызвать `tracking.SetStatus(BUSY)` если дрон найден, иначе вернуть `success: false`
    * Вызвать `telemetry.SendCommand(FLY_TO(Store))`
2. Читать сообщения из Kafka топика `events`:
    * Если `ARRIVED_AT_STORE` -> шлем `PICKUP_CARGO`
    * Если `PICKED_UP_CARGO` -> шлем `FLY_TO(Client)`
    * Если `ARRIVED_AT_CLIENT` -> шлем `DROP_CARGO`
    * Если `DROPPED_CARGO` -> шлем `FLY_TO(Base)`, `order.UpdateStatus(COMPLETED)` и `tracking.SetStatus(FREE)`
    * Если `ARRIVED_AT_BASE` -> шлем `CHARGE` и `tracking.SetStatus(CHARGING)` если `battery < 80%`, иначе просто `WAIT` (статус остается `FREE`)
    * Если `FULLY_CHARGED` -> шлем `WAIT` и `tracking.SetStatus(FREE)`

Proto контракт:
```proto
syntax = "proto3";  
  
package dispatch;  
option go_package = "hive/gen/dispatch";  
  
service DispatchService {  
  rpc AssignDrone(AssignDroneRequest) returns (AssignDroneResponse);  
}  
  
message Location {  
  double lat = 1;  
  double lon = 2;  
}  
  
message AssignDroneRequest {  
  string order_id = 1;  
  Location delivery_location = 2;  
}  
  
message AssignDroneResponse {  
  bool success = 1;  
  string drone_id = 2;  
}
```


#### API Gateway
Сервер на `echo`, принимает REST HTTP запросы от клиентского приложения (фронтенда), не содержит бизнес-логики, занимается только маршрутизацией и базовой валидации
Нужно:
1. Поднять HTTP сервер на порту `8080`
2. POST `/api/v1/orders`:
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
3. GET `/api/v1/orders`:
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

OpenAPI спецификация:
```openapi
openapi: 3.0.1
info:
  title: Hive Drone Delivery API
  description: API Gateway для взаимодействия клиентов с системой доставки дронами.
  version: 1.0.0
servers:
  - url: http://localhost:8080
    description: Local Development

paths:
  # ---------------------------------------------------------------------------
  # ORDERS API
  # ---------------------------------------------------------------------------
  /api/v1/orders:
    post:
      summary: Создать новый заказ
      operationId: createOrder
      tags:
        - Orders
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrderRequest'
      responses:
        '201':
          description: Заказ успешно создан
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CreateOrderResponse'
        '400':
          description: Ошибка валидации (неверные координаты, пустой список)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
        '500':
          description: Внутренняя ошибка сервера

  /api/v1/orders/{id}:
    get:
      summary: Получить статус заказа
      description: Возвращает статус заказа и, если назначен дрон, его текущее местоположение.
      operationId: getOrder
      tags:
        - Orders
      parameters:
        - name: id
          in: path
          required: true
          description: ID заказа (UUID)
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Информация о заказе
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetOrderResponse'
        '404':
          description: Заказ не найден
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  # ---------------------------------------------------------------------------
  # USERS API (FUTURE)
  # Пока закомментировано для MVP
  # ---------------------------------------------------------------------------
  # /api/v1/auth/register:
  #   post:
  #     summary: Регистрация пользователя
  #     tags: [Users]
  #     requestBody:
  #       content:
  #         application/json:
  #           schema:
  #             type: object
  #             properties:
  #               email: {type: string, format: email}
  #               password: {type: string}
  #               name: {type: string}
  #     responses:
  #       '200':
  #         description: Успешная регистрация
  #
  # /api/v1/users/{id}:
  #   get:
  #     summary: Профиль пользователя
  #     tags: [Users]
  #     parameters:
  #       - name: id
  #         in: path
  #         required: true
  #         schema: {type: string, format: uuid}
  #     responses:
  #       '200':
  #         description: Профиль
  #         content:
  #           application/json:
  #             schema:
  #               $ref: '#/components/schemas/User'


components:
  schemas:
    # --- Basic Types ---
    Location:
      type: object
      required:
        - lat
        - lon
      properties:
        lat:
          type: number
          format: double
          example: 55.7531753
          description: Широта (-90 to 90)
        lon:
          type: number
          format: double
          example: 37.6104283
          description: Долгота (-180 to 180)

    # --- Requests ---
    CreateOrderRequest:
      type: object
      required:
        - user_id
        - items
        - delivery_location
      properties:
        user_id:
          type: string
          format: uuid
          example: "550e8400-e29b-41d4-a716-446655440000"
        items:
          type: array
          items:
            type: string
          minItems: 1
          example: [ "pizza", "coke" ]
        delivery_location:
          $ref: '#/components/schemas/Location'

    # --- Responses ---
    CreateOrderResponse:
      type: object
      properties:
        order_id:
          type: string
          format: uuid
        status:
          $ref: '#/components/schemas/OrderStatus'
        estimated_time:
          type: string
          example: "15 min"

    GetOrderResponse:
      type: object
      properties:
        order_id:
          type: string
          format: uuid
        status:
          $ref: '#/components/schemas/OrderStatus'
        drone:
          $ref: '#/components/schemas/DroneInfo'
          description: Информация о дроне (присутствует, если статус ASSIGNED/DELIVERING)

    # --- Models ---
    OrderStatus:
      type: string
      enum:
        - PENDING
        - ASSIGNED
        - COMPLETED
        - FAILED
      example: "ASSIGNED"

    DroneInfo:
      type: object
      properties:
        id:
          type: string
        location:
          $ref: '#/components/schemas/Location'
        battery:
          type: integer
          minimum: 0
          maximum: 100
          example: 85

    User:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        email: { type: string }

    Error:
      type: object
      properties:
        code:
          type: integer
          example: 400
        message:
          type: string
          example: "Delivery location is out of service area"
```



#### Drone Emulator (отдельный клиентский скрипт, не относится к сервисам)
*Отличие эмулятора от симулятора в том, что эмулятор воспроизводит интерфейс и поведение системы, а симулятор точную физику и поведение дрона в реальном мире.
В MVP нашего проекта используем **эмулятор**, потому что он не учитывает гравитацию, погодные условия и препятствия. Ему важно, чтобы на команду "Лети туда" он начал менять координаты с заданной скоростью*

Запускает N "виртуальных дронов", подключается с каждого дрона (отдельной функции) к Telemetry Service
Нужно:
1. Для каждого дрона запустить отдельную горутину
2. Логика одного дрона:
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
                * Проверить: если расстояние до цели <5 метров (учет погрешности) -> отправить событие (`ARRIVED`, `PICKED_UP`, `DROPPED`)
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
