package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"utils/structs"
)

// CREA ARCHIVO .LOG
func ConfigurarLogger(nombreArchivoLog string) {
	logFile, err := os.OpenFile(nombreArchivoLog+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile)
	slog.Info("Logger " + nombreArchivoLog + ".log configurado")
}

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

func RecibirMensaje(w http.ResponseWriter, r *http.Request) {
	mensaje, err := DecodificarMensaje[string](r)
	if err != nil {
		slog.Error(fmt.Sprintf("No se pudo decodificar el mensaje (%v)", err))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Error al decodificar mensaje"))
		return
	}

	slog.Info(fmt.Sprintf("Me llego un mensaje: %+v",mensaje))

	respuesta := structs.Respuesta {
		Mensaje: fmt.Sprint(http.StatusOK),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(respuesta)
}

func RecibirPCB(w http.ResponseWriter, r *http.Request) {
	mensaje, err := DecodificarMensaje[structs.PCB](r)
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


// log obligatorio para cuando se crea un nuevo proceso (1)
func NuevoProceso(nuevoPCB structs.PCB) {
	log.Printf("Se crea el proceso %d en estado %s", nuevoPCB.PID, nuevoPCB.Estado)
}