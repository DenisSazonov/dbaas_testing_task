# Тестовое задание DBaaS

Этот проект содержит e2e тест для API Database as a Service (DBaaS). Тест выполняет следующие шаги:
1. Авторизация через API.
2. Создаёт двухнодовый кластер Postgres.
3. Создаёт базу данных.
4. Создаёт пользователя базы данных.
5. Подключается к базе данных
6. Создаёт схему данных и таблицу. Добавляет в таблицу произвольные данные.
7. Создаёт дамп базы данных
8. Очищает созданную таблицу
9. Восстанавливат базу данных из дампа
10. Проверяет что записи в таблице успешно восстановлены
11. После выполнения тест удаляет кластер и дамп базы данных

## Требования

Перед началом работы убедитесь, что у вас установлены следующие компоненты:
- [Go](https://golang.org/doc/install) (версия 1.16 или выше).
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

## Настройка проекта

1. **Клонируйте репозиторий:**
    ```sh
    git clone https://github.com/DenisSazonov/DBaaS_test_task.git
    cd dbaas_testing_task
    ```

2. **Настройте переменные окружения:**
    **Для Windows (PowerShell):**
    ```powershell
    $Env:API_LOGIN="your_api_login"
    $Env:API_PASSWORD="your_api_password"
    $Env:API_BASE_URL="https://example.ru"
    ```

    **Для Linux (Shell):**
    ```sh
    export API_LOGIN=your_api_login
    export API_PASSWORD=your_api_password
    export API_BASE_URL=https://example.ru
    ```

3. **Установите зависимости:**
    Проект использует пакет [pgx](http://_vscodecontentref_/2) для подключения к PostgreSQL. Установите его с помощью:
    ```sh
    go get github.com/jackc/pgx/v4
    ```

    Также проект использует библиотеку [testify](https://github.com/stretchr/testify). Установите её с помощью:
    ```sh
    go get github.com/stretchr/testify
    ```

## Запуск тестов

1. **Запустите тест:**
    ```sh
    go test -v
    ```

    Эта команда выполнит функцию [TestEndToEnd](http://_vscodecontentref_/3) в файле [dbaas_test.go](http://_vscodecontentref_/4), которая выполняет всю последовательность операций, описанных выше.

## Структура проекта

- [dbaas_test.go](http://_vscodecontentref_/5): Содержит основную тестовую функцию [TestEndToEnd](http://_vscodecontentref_/6), которая выполняет e2e тест.
- [http_helpers.go](http://_vscodecontentref_/7): Содержит вспомогательные функции для выполнения HTTP-запросов и разбора ответов.
- [models.go](http://_vscodecontentref_/8): Содержит структуры данных, используемые в проекте.
- [go.mod](http://_vscodecontentref_/13): Содержит информацию о зависимостях и модулях Go, используемых в проекте.
- [go.sum](http://_vscodecontentref_/14): Содержит контрольные суммы для зависимостей, указанных в go.mod.
- [helpers.go](http://_vscodecontentref_/15): Содержит вспомогательные функции для выполнения различных операций, таких как авторизация и очистка данных.

![Ироничная шутка](https://cdn66.printdirect.ru/cache/product/2b/15/8307709/tov/all/480z480_front_2258_0_0_0_7ae301566b4e4201ef18ba45ec30.jpg)
