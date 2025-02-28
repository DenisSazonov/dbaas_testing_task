package main

import (
    "bytes"
    "encoding/json"
    "io"
    "net/http"
    "testing"
)


func makeRequest(t *testing.T, method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
    var bodyReader io.Reader
    if body != nil {
        bodyBytes, err := json.Marshal(body)
        if err != nil {
            t.Fatalf("Ошибка при сериализации тела запроса: %v", err)
        }
        bodyReader = bytes.NewReader(bodyBytes)
    }

    req, err := http.NewRequest(method, url, bodyReader)
    if err != nil {
        t.Fatalf("Ошибка при создании запроса: %v", err)
    }

    for key, value := range headers {
        req.Header.Add(key, value)
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("Ошибка при выполнении запроса: %v", err)
    }

    return resp, err
}

func parseResponseBody(t *testing.T, resp *http.Response, result interface{}) {
    defer resp.Body.Close()
    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatalf("Ошибка при чтении тела ответа: %v", err)
    }
    err = json.Unmarshal(bodyBytes, result)
    if err != nil {
        t.Fatalf("Ошибка при разборе JSON: %v", err)
    }
}