package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
)

var (
	apiBaseURL            string
	createDBResponse      CreateDBResponse
	responseDBUsers       []ResponseDBUsers
	clusterStatusResponse ClusterStatusResponse
	createClusterResponse CreateClusterResponse
	tableSpaceResponse    []TableSpaceResponse
	authResponse          AuthResponse
	refreshToken          string
	clusterId             string
	conString             string
	login                 string
	password              string
	tableSpaceId          string
	dbId                  string
	dumpId                string
	typeId   			  string
	flavorId 			  string
)

// Функция инициализации, которая считывает логин, пароль и эндроинт api из переменных окружения
func init() {
	apiBaseURL = os.Getenv("API_BASE_URL")
	login = os.Getenv("API_LOGIN")
	password = os.Getenv("API_PASSWORD")
}
// Функция для получения flavorId по имени
func GetFlavorID(t *testing.T) string {
	// Создание запроса
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/flavors", apiBaseURL), nil)
	if err != nil {
		t.Fatalf("Ошибка при создании запроса: %v", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+refreshToken)

	// Отправка запроса
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
	}

	// Декодирование ответа
	var flavorsResponse []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&flavorsResponse); err != nil {
		t.Fatalf("Ошибка при декодировании ответа: %v", err)
	}

	// Поиск flavorId по имени
	for _, flavor := range flavorsResponse {
		if flavor.Name == "STD3-1-1" {
			flavorId = flavor.ID
			return flavorId
		}
	}

	t.Error("Flavor с именем STD3-1-1 не найден")
	return ""
}
// Функция для получения type_id по версии
func GetTypeID(t *testing.T) string {
    // Создание запроса
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/types", apiBaseURL), nil)
    if err != nil {
        t.Fatalf("Ошибка при создании запроса: %v", err)
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", "Bearer "+refreshToken)

    // Отправка запроса
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("Ошибка при отправке запроса: %v", err)
    }
    defer resp.Body.Close()

    // Проверка статуса ответа
    if resp.StatusCode != http.StatusOK {
        t.Fatalf("Ожидался статус 200, получен: %d", resp.StatusCode)
    }

    // Декодирование ответа
    var typesResponse []struct {
        ID      string `json:"id"`
        Version string `json:"version"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&typesResponse); err != nil {
        t.Fatalf("Ошибка при декодировании ответа: %v", err)
    }

    // Поиск type_id по версии
    for _, typ := range typesResponse {
        if typ.Version == "17.2.2" {
            return typ.ID
        }
    }

    t.Error("Type с версией Postgres Pro Enterprise не найден")
    return ""
}

// Функция для удаления кластера, используется для очистки после тестов
func Teardown(t *testing.T) {
	// Создание запроса на удаление дампа
	dumpReq, err := http.NewRequest(
		"DELETE", fmt.Sprintf("%s/api/dumps/%s", apiBaseURL, dumpId), nil)
	if err != nil {
		t.Fatalf("Ошибка при создании запроса на удаление дампа: %v", err)
	}
	dumpReq.Header.Add("Content-Type", "application/json")
	dumpReq.Header.Add("Authorization", "Bearer "+refreshToken)
	
	// Отправка запроса
	dumpClient := &http.Client{}
	dumpResp, err := dumpClient.Do(dumpReq)
	if err != nil {
		t.Fatalf("Ошибка при отправке запроса на удаление дампа: %v", err)
	}
	defer dumpResp.Body.Close()

	// Проверка статуса ответа
	if dumpResp.StatusCode != http.StatusNoContent {
		t.Fatalf("Ожидался статус 204, получен: %d", dumpResp.StatusCode)
	}
	t.Logf("Deleted dump with ID: %s", dumpId)	

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
	if err != nil {
		t.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer delresp.Body.Close()

	// Проверка статуса ответа
	if delresp.StatusCode != http.StatusNoContent {
		t.Fatalf("Ожидался статус 204, получен: %d", delresp.StatusCode)
	}
	t.Logf("Deleted cluster with ID: %s", clusterId)
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
