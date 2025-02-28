package main

import (
	"context"
	"fmt"
		"strings"
		"testing"
		"time"
		"net/http"
		"github.com/jackc/pgx/v4"
		"math/rand"
	)
	

func TestEndToEnd(t *testing.T) {
	defer Teardown(t)

	// Проверяем, что переменные окружения для логина и пароля установлены
	if login == "" || password == "" {
		t.Fatal("Отсутствуют переменные окружения API_LOGIN и/или API_PASSWORD")
	}

	// Авторизация
	requestBody := map[string]string{
		"login":    login,
		"password": password,
	}

	// Выполняем POST запрос для авторизации
	resp, err := makeRequest(t, "POST", fmt.Sprintf("%s/api/authorize", apiBaseURL), requestBody, map[string]string{
		"Content-Type": "application/json",
	})
	if err != nil {
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
	}

	// Парсим ответ авторизации
	parseResponseBody(t, resp, &authResponse)

	// Проверяем, что refresh_token не пустой
	if authResponse.RefreshToken == "" {
		t.Error("Поле refresh_token пустое в ответе API")
	}

	refreshToken = authResponse.RefreshToken
	t.Logf("Authorization successful")

	// Создание кластера
	createClusterRequestBody := CreateClusterRequest{
		TypeID: "581690c7-adf2-4042-9953-d32b590b97e4", //FIXME: заменить на получение из /api/types
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
		FlavorID:      "94d277e1-08bd-45cd-824e-8ddaaa8325ef", //FIXME: заменить на получение из /api/flavors
		TypeName:      "Postgres Pro Enterprise",
		Az:            "GZ1",
		HAManager:     "patroni",
		HA:            false,
	}

	// Выполняем POST запрос для создания кластера
	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters", apiBaseURL), createClusterRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Ожидался статус 201, получен: %d", resp.StatusCode)
	}

	// Парсим ответ создания кластера
	parseResponseBody(t, resp, &createClusterResponse)
	clusterId = createClusterResponse.Instances[0].ClusterID
	t.Logf("Cluster created with ID: %s", clusterId)

	// Проверка статуса кластера вместо ожидания
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId, apiBaseURL), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		if err != nil {
			Teardown(t)
			t.Fatalf("Ошибка при выполнении запроса: %v", err)
		}

		parseResponseBody(t, resp, &clusterStatusResponse)

		if clusterStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Cluster status is %s, waiting for OK", clusterStatusResponse.Status)
		time.Sleep(5 * time.Second)
	}

	// Создание TableSpace
	resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId+"/tablespaces", apiBaseURL), nil, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		Teardown(t)
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}

	// Парсим ответ получения TableSpace
	parseResponseBody(t, resp, &tableSpaceResponse)
	tableSpaceId = tableSpaceResponse[0].Id

	// Создание базы данных
	createDBRequestBody := CreateDBRequest{
		Name:         "testDB",
		TableSpaceID: tableSpaceId,
	}

	// Выполняем POST запрос для создания базы данных
	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters/"+clusterId+"/databases", apiBaseURL), createDBRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		Teardown(t)
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		Teardown(t)
		t.Fatalf("Ожидался статус 201, получен: %d", resp.StatusCode)
	}

	// Парсим ответ создания базы данных
	parseResponseBody(t, resp, &createDBResponse)
	dbId = createDBResponse.Id
	t.Logf("Database created with ID: %s", dbId)

	// Ждем, пока база данных станет доступной
	for i := 0; i < 30; i++ {
		resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId+"/databases/"+dbId, apiBaseURL), nil, map[string]string{
			"Authorization": "Bearer " + refreshToken,
			"Content-Type":  "application/json",
		})
		if err != nil {
			Teardown(t)
			t.Fatalf("Ошибка при выполнении запроса: %v", err)
		}

		parseResponseBody(t, resp, &dbStatusResponse)

		if dbStatusResponse.Status == "OK" {
			break
		}
		t.Logf("Database status is %s, waiting for OK", dbStatusResponse.Status)
		time.Sleep(5 * time.Second)
	}

	// Создание пользователя кластера
	createClusterUserRequestBody := CreateClusterUserRequest{
		Databases: []string{"testDB"},
		Roles:     []string{"pg_write_all_data", "pg_read_all_data"},
		Name:      login,
		Password:  password,
	}

	// Выполняем POST запрос для создания пользователя кластера
	resp, err = makeRequest(t, "POST", fmt.Sprintf("%s/api/clusters/"+clusterId+"/users", apiBaseURL), createClusterUserRequestBody, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		Teardown(t)
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		Teardown(t)
		t.Fatalf("Ожидался статус 201, получен: %d", resp.StatusCode)
	}

	t.Logf("Database user created")

	// Запрашиваем connection string для пользователя базы данных
	resp, err = makeRequest(t, "GET", fmt.Sprintf("%s/api/clusters/"+clusterId+"/databases", apiBaseURL), nil, map[string]string{
		"Authorization": "Bearer " + refreshToken,
		"Content-Type":  "application/json",
	})
	if err != nil {
		Teardown(t)
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}

	// Парсим ответ получения пользователей базы данных
	parseResponseBody(t, resp, &responseDBUsers)

	// Формируем строку подключения с использованием логина и пароля пользователя
	conString = responseDBUsers[0].MasterConnectionString
	conString = strings.Replace(conString, "<username>", login, 1)
	conString = strings.Replace(conString, "<password>", password, 1)

	// Создание таблицы в базе данных
	ctx := context.Background()

	// Подключаемся к базе данных
	conn, err := pgx.Connect(ctx, conString)
	if err != nil {
		Teardown(t)
		t.Fatalf("не удалось подключиться к базе данных: %v", err)
	}
	defer conn.Close(ctx)

	// Создаем схему, если она не существует
	_, err = conn.Exec(ctx, `
		CREATE SCHEMA IF NOT EXISTS test_schema;
	`)
	if err != nil {
		Teardown(t)
		t.Fatalf("не удалось создать схему: %v", err)
	}
	t.Logf("Schema created")

	// Создаем таблицу, если она не существует
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS test_schema.users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) NOT NULL,
			age INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		Teardown(t)
		t.Fatalf("не удалось создать таблицу: %v", err)
	}
	t.Logf("Table created")

	// Генерируем случайные данные и вставляем их в таблицу
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Пользователь %d", i)
		email := fmt.Sprintf("user%d@example.com", i)
		age := r.Intn(50) + 18

		_, err = conn.Exec(ctx, `
			INSERT INTO test_schema.users (name, email, age)
			VALUES ($1, $2, $3)
		`, name, email, age)

		if err != nil {
			Teardown(t)
			t.Fatalf("не удалось вставить данные: %v", err)
		}
	}
	t.Logf("Data inserted")
}