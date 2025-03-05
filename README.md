# Распределённый вычислитель арифметических выражений
Данный проект распределённо вычисляет арифметические выражения. Сервер, когда ему присылают выражение, разбивает его на задачи, которые отправляются агентам на вычисление.
## Установка
1. Установите язык программирования [Golang](https://go.dev/dl/).
2. Установите текстовый редактор [Visual Studio Code](https://code.visualstudio.com/).
3. Установите систему контроля версий [Git](https://git-scm.com/downloads).
4. Создайте папку и откройте ее в Visual Studio Code.
5. В проекте слева нажмите на 4 квадратика - Extensions. В поле поиска введите go и скачайте первый модуль под названием Go.
6. Создайте клон репозитория с GitHub. Для этого в терминале Visual Studio Code введите следующую команду:
```
git clone https://github.com/kingofhandsomes/distributed-calculator-go
```
## Использование в Windows
1. В терминале Visual Studio Code перейдите в папку calculation_go с помощью команды:
```
cd distributed-calculator-go
```
2. Введите команду ниже для установки пакета gorilla/mux, если потребуется:
```
go get github.com/gorilla/mux
```
3. Введите команду ниже для установки пакета godotenv, если потребуется:
```
go get github.com/joho/godotenv
```
4. В файле variables.env измените переменные среды, если Вы хотите поменять программы:
- PORT - отвечает за порт, на котором будет работать сервер, принимает значения от 0 до 9999, по-умолчанию раверн 8080;
- TIME_ADDITION_MS - отвечает за время в миллисекундах, которое будет имитировать работу операции сложение, принимает значение от 0 до бесконечности, по-умолчанию 1;
- TIME_SUBTRACTION_MS - отвечает за время в миллисекундах, которое будет имитировать работу операции вычитание, принимает значение от 0 до бесконечности, по-умолчанию 1;
- TIME_MULTIPLICATIONS_MS - отвечает за время в миллисекундах, которое будет имитировать работу операции умножение, принимает значение от 0 до бесконечности, по-умолчанию 1;
- TIME_DIVISIONS_MS - отвечает за время в миллисекундах, которое будет имитировать работу операции деление, принимает значение от 0 до бесконечности, по-умолчанию 1;
- COMPUTING_POWER - отвечает за количество одновременно работающих агентов, которые решают математические операции, принимает значение от 0 до бесконечности, по-умолчанию 1;
5. Запустите веб-сервис, введя следующую команду:
```
go run cmd/main.go
```
## Отправка выражения на вычисление
В выражении можно использовать:
- \+ (сложение);
- \- (вычитание);
- \* (умножение);
- / (деление);
- скобки (открывающиеся, закрывающиеся);
- (-2) (отрицательные числа, указываются в скобках);
- 2.1 (вещественные числа);
- (-2.2) (отрицательные вещественные числа);  
Чтобы отправить выражение на вычисление, необходимо открыть приложение Git Bash и ввести команду:
```
curl --location --request POST 'localhost:<ПОРТ>/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression":"<ВЫРАЖЕНИЕ>"}'
```
Примеры отправки запроса:
1. Удачный:
```
curl --location --request POST 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression":"1+(-2)-3/(-4.1)*5"}'
```
2. Неудачные:
- Неверный метод, необходим POST, статус код 405:
```
curl --location --request GET 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression":"1+(-2)-3/(-4.1)*5"}'
```
- Неверная структура (ей является неверная json-структура при отправке (например, '{""}') или неверное выражение, в котором будет находиться иные символы, лишнее количество скобок, деление на ноль, неправильное написание вещественного и отрицательного числа (например, 2.1.1 или -2 без скобок)), статус код 422:
```
curl --location --request POST 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression":"1/0+((-2.1.1)-$3/(-4.1)*-5"}'
```
Результат запроса:
```
invalid data
```
## Вывод состояния всех выражений
Примеры отправки запроса:
1. Удачный:
```
curl --location --request GET 'localhost:8080/api/v1/expressions'
```
Результат запроса:
```
{"expressions":[{"id":1,"status":"resolved","result":2.6585365853658542},{"id":2,"status":"resolved","result":2}]}
```
2. Неудачный:
- Неверный метод, необходим GET, статус код 405:
```
curl --location --request POST 'localhost:8080/api/v1/expressions'
```
## Вывод состояния одного выражения
После expressions/ надо указать id выражения:  
Примеры отправки запроса:
1. Удачный (если на сервере есть выражение с id - 1):
```
curl --location --request GET 'localhost:8080/api/v1/expressions/1'
```
Результат запроса:
```
{"expression":{"id":1,"status":"resolved","result":2.6585365853658542}}
```
2. Неудачные
- Неверный метод, необходим GET, статус код 405:
```
curl --location --request POST 'localhost:8080/api/v1/expressions/1'
```
- Неверный id выражения (на сервере нет выражения с таким id), статус код 404:
```
curl --location --request GET 'localhost:8080/api/v1/expressions/2'
```
Результат запроса:
```
there is no such expression
```
### Работа с задачами
Следующие запросы предназначены ТОЛЬКО для агентов, поэтому их не стоит вызывать.
1. Примеры взятия задачи для решения:
- Удачный:
```
curl --location --request GET 'localhost:8080/internal/task'
```
Результат запроса:
```
{"task":{"id": 1, "arg1": 2, "arg2": 2, "operation": "+", "operation_time": 0}}
```
- Неудачный:  
На сервере закончились нерешенные задачи, статус код 404:
```
curl --location --request GET 'localhost:8080/internal/task'
```
Результат запроса:
```
there is no such expression
```
2. Примеры принятия решения для задачи:
- Удачный (если на сервере есть задача с id - 1):
```
curl --location --request POST 'localhost:8080/internal/task' --header 'Content-Type: application/json' --data '{"id":1,"result":4,"operation_time":1}'
```
- Неудачные:
- - На сервере нет задачи с таким id (если на сервере нет задачи с id - 2), статус код 404:
```
curl --location --request POST 'localhost:8080/internal/task' --header 'Content-Type: application/json' --data '{"id":2,"result":4,"operation_time":1}'
```
Результат запроса:
```
there is no such expression
```
- - Неверно указана json-структура, статус код 422:
```
curl --location --request POST 'localhost:8080/internal/task' --header 'Content-Type: application/json' --data '{""}'
```
Результат запроса:
```
invalid data
```
