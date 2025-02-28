package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

// const apiBaseURL = "https://pub.dbaas.postgrespro.ru"

var (
	apiBaseURL   string
	createDBResponse CreateDBResponse
	responseDBUsers []ResponseDBUsers
	clusterStatusResponse ClusterStatusResponse
	createClusterResponse CreateClusterResponse
	tableSpaceResponse []TableSpaceResponse
	authResponse AuthResponse
	refreshToken string
	clusterId    string
	conString    string
	login        string
	password     string
	tableSpaceId string
	dbId         string
)

// Функция инициализации, которая считывает логин и пароль из переменных окружения
func init() {
	apiBaseURL = os.Getenv("API_BASE_URL")
	login = os.Getenv("API_LOGIN")
	password = os.Getenv("API_PASSWORD")
}

// Функция для удаления кластера, используется для очистки после тестов
func Teardown(t *testing.T) {
	// Создание запроса на удаление кластера
	delreq, err := http.NewRequest(
		"DELETE", fmt.Sprintf("%s/api/clusters/"+clusterId, apiBaseURL), nil)
	if err != nil {
		t.Fatalf("Ошибка при создании запроса на удаление: %v", err)
	}
	delreq.Header.Add("Content-Type", "*/*")
	delreq.Header.Add("Authorization", "Bearer "+refreshToken)
	
	// Отправка запроса
	delclient := &http.Client{}
	delresp, err := delclient.Do(delreq)
	if (err != nil) {
		t.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer delresp.Body.Close()

	// Проверка статуса ответа
	if delresp.StatusCode != http.StatusNoContent {
		t.Fatalf("Ожидался статус 204, получен: %d", delresp.StatusCode)
	}
	t.Logf("Cluster deleted")
}

// Функция авторизации, используется для получения токена
func Authorize(t *testing.T) {
	// Проверка наличия логина и пароля в переменных окружения
	if login == "" || password == "" {
		t.Fatal("Отсутствуют переменные окружения API_LOGIN и/или API_PASSWORD")
	}

	// Создание тела запроса в формате JSON
	requestBody, err := json.Marshal(map[string]string{
		"login":    login,
		"password": password,
	})
	if err != nil {
		t.Fatalf("Ошибка при создании JSON запроса: %v", err)
	}

	// Отправка запроса на авторизацию
	resp, err := http.Post(
		fmt.Sprintf("%s/api/authorize", apiBaseURL),
		"application/json",
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		t.Fatalf("Ошибка при выполнении запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
	}

	// Декодирование ответа
	var authResponse AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		t.Fatalf("Ошибка при декодировании ответа: %v", err)
	}

	// Проверка наличия токена в ответе
	if authResponse.RefreshToken == "" {
		t.Error("Поле refresh_token пустое в ответе API")
	}

	// Сохранение токена
	refreshToken = authResponse.RefreshToken
	t.Logf("Токен успешно получен")
}