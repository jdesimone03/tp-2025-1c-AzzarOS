package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type Mensaje struct {
	Mensaje string `json:"mensaje"`
}

// CREA ARCHIVO .LOG
func ConfigurarLogger(nombreArchivoLog string) {
	logFile, err := os.OpenFile(nombreArchivoLog+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // FALTO CAMBIAR A SLOG, TE LO DEJO A VOS FABRI
	slog.Info("Logger " + nombreArchivoLog + ".log configurado")
}

func CargarConfiguracion[T any](filePath string) *T {
	var config T

	file, err := os.Open(filePath)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo abrir el archivo de configuración  (%v)", err))
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el archivo JSON (%v)", err))
		panic(err)
	}

	slog.Info(fmt.Sprintf("Configuración cargada correctamente: %+v", config))
	return &config
}

func EnviarMensaje(ip string, puerto int, endpoint string, mensaje any) {
	body, err := json.Marshal(mensaje)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo codificar el mensaje (%v)", err))
		return
	}

	url := fmt.Sprintf("http://%s:%d/%s", ip, puerto, endpoint)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo enviar mensaje a %s:%d/%s (%v)", ip, puerto, endpoint, err))
		return
	}

	// log.Printf("respuesta del servidor: %s", resp.Status)
	slog.Info(fmt.Sprintf("Respuesta de %s:%d/%s %v", ip, puerto, endpoint, resp.Status))
}

func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	mensaje, err := DecodificarMensaje[string](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Info(fmt.Sprintf("Me llego un mensaje: %+v",mensaje))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func RecibirPCB(w http.ResponseWriter, r *http.Request) {
	mensaje, err := DecodificarMensaje[PCB](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Info(fmt.Sprintf("Me llego un mensaje: %+v",mensaje))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
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
	err := http.ListenAndServe(fmt.Sprintf(":%d",puerto), nil)
	if err != nil {
		panic(err)
	}
	return err
}