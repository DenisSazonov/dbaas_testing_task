package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestEndToEnd(t *testing.T) {
	defer Teardown(t)

	// Проверяем наличие переменных окружения
	assert.NotEmpty(t, login, "Отсутствует переменная окружения API_LOGIN")
	assert.NotEmpty(t, password, "Отсутствует переменная окружения API_PASSWORD")
	assert.NotEmpty(t, apiBaseURL, "Отсутствует переменная окружения API_BASE_URL")

	// Шаг 1: Авторизация через API
	requestBody := map[string]string{
		"login":    login,
		"password": password,
	}

	resp, err := makeRequest(t, "POST", fmt.Sprintf("%s/api/authorize", apiBaseURL), requestBody, map[string]string{
		"Content-Type": "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на авторизацию")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Ожидался статус 200")

	parseResponseBody(t, resp, &authResponse)
	assert.NotEmpty(t, authResponse.RefreshToken, "Поле refresh_token пустое в ответе API")

	refreshToken = authResponse.RefreshToken
	t.Logf("Authorization successful")

	// Шаг 2: Создаём двухнодовый кластер Postgres
	typeId = GetTypeID(t)
	assert.NotEmpty(t, typeId,"Ошибка при получении TypeID")
	flavorId = GetFlavorID(t)
	assert.NotEmpty(t, flavorId, "Ошибка при получении FlavorID")
	// Наполняем и отправляем запрос на создание кластера
	createClusterRequestBody := CreateClusterRequest{
		TypeID: typeId,
		Options: Options{
			MaximumLagOnFailover:  1048576,
			WalArchiveMode:        false,
			AutoRestart:           false,
			Production:            false,
			EnableSynchronousMode: false,
			DisableAutofailover:   false,
		},
		DiskSize:      3221225472,
		Mode:          "create",
		ReplicasCount: 1,
		CreationMode:  "empty",
		Name:          "test",
		FlavorID:      flavorId,
		TypeName:      "Postgres Pro Enterprise",
		Az:            "GZ1",
		HAManager:     "patroni",
		HA:            false,
	}

	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters", apiBaseURL), createClusterRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на создание кластера")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался статус 201")

	parseResponseBody(t, resp, &createClusterResponse)
	clusterId = createClusterResponse.Instances[0].ClusterID
	t.Logf("Cluster created with ID: %s", clusterId)
	//Ждём пока кластер перейдёт в состояние OK
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId, apiBaseURL), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о кластере")

		parseResponseBody(t, resp, &clusterStatusResponse)

		if clusterStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Cluster status is %s, waiting for OK at %s", clusterStatusResponse.Status, time.Now().Format("2006-01-02 15:04:05.000"))
		time.Sleep(5 * time.Second)
	}
	// Получаем список tablespace и используем дефолтный
	resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId+"/tablespaces", apiBaseURL), nil, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о tablespace")

	parseResponseBody(t, resp, &tableSpaceResponse)
	tableSpaceId = tableSpaceResponse[0].Id

	// Шаг 3: Создаём базу данных
	createDBRequestBody := CreateDBRequest{
		Name:         "testDB",
		TableSpaceID: tableSpaceId,
	}
	// Наполняем и отправляем запрос на создание базы данных
	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters/"+clusterId+"/databases", apiBaseURL), createDBRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на создание базы данных")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался статус 201")

	parseResponseBody(t, resp, &createDBResponse)
	dbId = createDBResponse.Id
	t.Logf("Database created with ID: %s", dbId)
	// Ждём пока база данных перейдёт в состояние OK
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId+"/databases/"+dbId, apiBaseURL), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о базе данных")

		parseResponseBody(t, resp, &dbStatusResponse)

		if dbStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Database status is %s, waiting for OK at %s", dbStatusResponse.Status, time.Now().Format("2006-01-02 15:04:05.000"))
		time.Sleep(5 * time.Second)
	}
    // Шаг 4: Создаём пользователя базы данных
	createClusterUserRequestBody := CreateClusterUserRequest{
		Databases: []string{"testDB"},
		Roles:     []string{"pg_write_all_data", "pg_read_all_data"},
		Name:      login,
		Password:  password,
	}
	// Наполняем и отправляем запрос на создание пользователя
	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters/"+clusterId+"/users", apiBaseURL), createClusterUserRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на создание пользователя")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался статус 201")

	t.Logf("Database user created with login: %s", login)

    // Шаг 5: Подключаемся к базе данных
	resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId+"/databases", apiBaseURL), nil, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о базах данных")

	// Парсим ответ и формируем connection string
	parseResponseBody(t, resp, &responseDBUsers)

	conString = responseDBUsers[0].MasterConnectionString
	conString = strings.Replace(conString, "<username>", login, 1)
	conString = strings.Replace(conString, "<password>", password, 1)
	// Подключаемся к базе данных
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, conString)
	assert.NoError(t, err, "не удалось подключиться к базе данных")
	defer conn.Close(ctx)
    // Шаг 6: Создаём схему данных и таблицу. Добавляем в таблицу произвольные данные

	_, err = conn.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS test_schema;
	`)
	assert.NoError(t, err, "не удалось создать схему данных")
	t.Logf("Schema created successfully")
	// Создаём таблицу users
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_schema.users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			age INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	assert.NoError(t, err, "не удалось создать таблицу")
	t.Logf("Table 'Users' created")

	// Добавляем в таблицу произвольные данные
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Пользователь %d", i)
		email := fmt.Sprintf("user%d@example.com", i)
		age := r.Intn(50) + 18

		_, err = conn.Exec(ctx, `
			INSERT INTO test_schema.users (name, email, age)
			VALUES ($1, $2, $3)
		`, name, email, age)

		assert.NoError(t, err, "не удалось вставить данные")
	}
	t.Logf("Random data inserted inserted into the table")

    // Шаг 7: Создаём дамп базы данных
	createDumpRequestBody := map[string]string{
		"name": "testBackup",
	}
	// Наполняем и отправляем запрос на создание дампа
	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters/%s/databases/%s/dumps", apiBaseURL, clusterId, dbId), createDumpRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Ожидался статус 201")

	var createDumpResponse struct {
		ID string `json:"id"`
	}
	parseResponseBody(t, resp, &createDumpResponse)
	dumpId = createDumpResponse.ID
	t.Logf("Database dump created with ID: %s", dumpId)
	// Ждём пока дамп перейдёт в состояние OK
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/dumps/%s", apiBaseURL, dumpId), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о дампе")

		var dumpStatusResponse struct {
			Status string `json:"status"`
		}
		parseResponseBody(t, resp, &dumpStatusResponse)

		if dumpStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Dump status is %s, waiting for OK at %s", dumpStatusResponse.Status, time.Now().Format("2006-01-02 15:04:05.000"))
		time.Sleep(5 * time.Second)
	}

    // Шаг 8: Очищаем созданную таблицу
	_, err = conn.Exec(ctx, `
		TRUNCATE TABLE test_schema.users;
	`)
	assert.NoError(t, err, "не удалось очистить таблицу")
	t.Logf("Table truncated")
    
	// Шаг 9: Восстанавливаем базу данных из дампа
	restoreDumpRequestBody := map[string]interface{}{
		"dump_id":       dumpId,
		"mode":          "full",
		"restore_users": false,
	}

	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters/%s/databases/%s/dump_restore", apiBaseURL, clusterId, dbId), restoreDumpRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	assert.NoError(t, err, "Ошибка при выполнении запроса на восстановление базы данных из дампа")
	// Ждём пока дамп перейдёт в состояние OK
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/dumps/%s", apiBaseURL, dumpId), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о дампе")

		var dumpStatusResponse struct {
			Status string `json:"status"`
		}
		parseResponseBody(t, resp, &dumpStatusResponse)

		if dumpStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Dump status is %s, waiting for OK at %s", dumpStatusResponse.Status, time.Now().Format("2006-01-02 15:04:05.000"))
		time.Sleep(5 * time.Second)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Ожидался статус 200")
	// Ждём пока база данных перейдёт в состояние OK после восстановления
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/%s/databases/%s", apiBaseURL, clusterId, dbId), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		assert.NoError(t, err, "Ошибка при выполнении запроса на получение информации о базе данных")

		parseResponseBody(t, resp, &dbStatusResponse)

		if dbStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Database status is %s, waiting for OK at %s", dbStatusResponse.Status, time.Now().Format("2006-01-02 15:04:05.000"))
		time.Sleep(5 * time.Second)
	}

	// Шаг 10: Проверяем что записи в таблице успешно восстановлены
	rows, err := conn.Query(ctx, `
		SELECT name, email, age
		FROM test_schema.users;
	`)
	assert.NoError(t, err, "не удалось выполнить запрос")
	defer rows.Close()

	var restoredUsers []struct {
		Name  string
		Email string
		Age   int
	}

	for rows.Next() {
		var user struct {
			Name  string
			Email string
			Age   int
		}
		err := rows.Scan(&user.Name, &user.Email, &user.Age)
		assert.NoError(t, err, "не удалось прочитать данные")
		restoredUsers = append(restoredUsers, user)
	}

	assert.Equal(t, 10, len(restoredUsers), "Ожидалось 10 записей")
	if assert.Equal(t, 10, len(restoredUsers), "Ожидалось 10 записей") {
		t.Logf("Data restored successfully")
	}
}
