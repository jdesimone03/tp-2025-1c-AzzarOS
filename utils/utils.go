package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"utils/structs"
)

func EnviarMensaje(ip string, puerto int, endpoint string, mensaje any) string {
	body, err := json.Marshal(mensaje)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo codificar el mensaje (%v)", err))
		return ""
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, puerto, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo enviar mensaje a %s:%d/%s (%v)", ip, puerto, endpoint, err))
		return ""
	}
	defer resp.Body.Close()

	var resData structs.Respuesta
	respuesta, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	err = json.Unmarshal(respuesta, &resData)
	if err != nil {
		return ""
	}

	// log.Printf("respuesta del servidor: %s", resp.Status)
	slog.Info(fmt.Sprintf("Respuesta de %s:%d/%s %v", ip, puerto, endpoint, resData))
	return resData.Mensaje
}

// Función genérica para decodificar un mensaje del body
func DecodificarMensaje[T any](r *http.Request) (*T, error) {
	var mensaje T
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&mensaje)
	if err != nil {
		return nil, err
	}
	return &mensaje, nil
}

func IniciarServidor(puerto int) error {
	slog.Info(fmt.Sprintf("Inicializando servidor en el puerto %d",puerto))
	err := http.ListenAndServe(fmt.Sprintf(":%d", puerto), nil)
	if err != nil {
		panic(err)
	}
	return err
}
