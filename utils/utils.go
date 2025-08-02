package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"utils/logueador"
)

func EnviarMensaje(ip string, puerto string, endpoint string, mensaje any) string {
	logueador.Debug("Enviando mensaje a %s:%s/%s con el contenido: %+v", ip, puerto, endpoint, mensaje)
	body, err := json.Marshal(mensaje)
	if err != nil {
		logueador.Error("No se pudo codificar el mensaje (%v)", err)
		return ""
	}

	url := fmt.Sprintf("http://%s:%s/%s", ip, puerto, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logueador.Error("No se pudo enviar mensaje a %s:%s/%s (%v)", ip, puerto, endpoint, err)
		return ""
	}
	defer resp.Body.Close()

	respuesta, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	respuestaStr := string(respuesta)

	// log.Printf("respuesta del servidor: %s", resp.Status)
	logueador.Debug("Respuesta de %s:%s/%s %v", ip, puerto, endpoint, respuestaStr)
	return respuestaStr
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

func IniciarServidor(puerto string) error {
	logueador.Info("Inicializando servidor en el puerto %s",puerto)
	err := http.ListenAndServe(":" + puerto, nil)
	if err != nil {
		panic(err)
	}
	return err
}

func ParsearNombreInstruccion(v any) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem() // Dereference if it's a pointer
	}

	name := t.Name()
	if strings.HasSuffix(name, "Instruction") {
		name = strings.TrimSuffix(name, "Instruction") // Remove "Instruction"
	}

	var result strings.Builder
	for i, char := range name {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result.WriteRune('_') // Add underscore before uppercase letters after the first one
		}
		result.WriteRune(char)
	}

	return strings.ToUpper(result.String()) // Convert to uppercase
}