package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/streadway/amqp"
)

type MensajeFinanzas struct {
	Order  []string
	Status string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func HacerCalculos(mensaje []byte, balance []float64) []float64 {
	mensajeStruct := MensajeFinanzas{}
	json.Unmarshal(mensaje, &mensajeStruct)
	valor, _ := strconv.ParseFloat(mensajeStruct.Order[2], 64)
	intentos, _ := strconv.ParseFloat(mensajeStruct.Order[6], 64)
	if mensajeStruct.Order[3] == "pyme" {
		if mensajeStruct.Status == "No entregado" && mensajeStruct.Order[5] == "true" {
			balance = append(balance, valor*0.3-10*intentos)
			ganancia := fmt.Sprintf("%f", valor*0.3-10*intentos)
			f, err := os.OpenFile("registro.txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString("Pedido: " + mensajeStruct.Order[0] + "(" + mensajeStruct.Order[1] + ") " + mensajeStruct.Status + " Balance:" + ganancia)
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()

		} else if mensajeStruct.Status == "No entregado" && mensajeStruct.Order[5] == "false" {
			balance = append(balance, -10*intentos)
			ganancia := fmt.Sprintf("%f", -10*intentos)
			f, err := os.OpenFile("registro.txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString("Pedido: " + mensajeStruct.Order[0] + "(" + mensajeStruct.Order[1] + ") " + mensajeStruct.Status + " Balance:" + ganancia + ":\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()
		} else if mensajeStruct.Status == "Entregado" && mensajeStruct.Order[5] == "true" {
			balance = append(balance, valor*1.3-10*intentos)
			ganancia := fmt.Sprintf("%f", valor*1.3-10*intentos)
			f, err := os.OpenFile("registro.txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString("Pedido: " + mensajeStruct.Order[0] + "(" + mensajeStruct.Order[1] + ") " + mensajeStruct.Status + " Balance:" + ganancia + ":\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()

		} else if mensajeStruct.Status == "Entregado" && mensajeStruct.Order[5] == "false" {
			balance = append(balance, valor*1.3-10*intentos)
			ganancia := fmt.Sprintf("%f", valor*1.3-10*intentos)
			f, err := os.OpenFile("registro.txt", os.O_APPEND|os.O_WRONLY, 0644)
			_, err = f.WriteString("Pedido: " + mensajeStruct.Order[0] + "(" + mensajeStruct.Order[1] + ") " + mensajeStruct.Status + " Balance:" + ganancia + ":\n")
			if err != nil {
				log.Fatal("whoops")
			}
			err = f.Close()

		}
	} else if mensajeStruct.Order[3] == "retail" {
		balance = append(balance, valor-10*intentos)
		ganancia := fmt.Sprintf("%f", valor-10*intentos)
		f, err := os.OpenFile("registro.txt", os.O_APPEND|os.O_WRONLY, 0644)
		_, err = f.WriteString("\nPedido: " + mensajeStruct.Order[0] + "(" + mensajeStruct.Order[1] + ") " + mensajeStruct.Status + "Ganancia:" + ganancia + ":\n")
		if err != nil {
			log.Fatal("whoops")
		}
		err = f.Close()
	}
	return balance
}
func PrintBalance(balance []float64) {
	suma := float64(0)
	for _, b := range balance {
		suma += b
	}
	fmt.Printf("El balance final es: %f", suma)
}

func SetupCloseHandler(balance []float64) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		PrintBalance(balance)
		os.Exit(0)
	}()
}
func ObtenerGanancia(linea string) float64 {
	lineaSeparada := strings.Split(linea, ":")
	flotante, _ := strconv.ParseFloat(lineaSeparada[2], 64)
	return flotante
}
func main() {
	var balanceGeneral []float64
	file, err2 := os.Open("registro.txt")
	if err2 != nil {
		log.Fatal(err2)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ganancia := ObtenerGanancia(scanner.Text())
		balanceGeneral = append(balanceGeneral, ganancia)
	}
	conn, err := amqp.Dial("amqp://guest:guest@10.10.28.47:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	mensajes, err := ch.Consume(
		"Finanzas",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	failOnError(err, "Failed to declare a queue")
	forever := make(chan bool)
	go func() {
		for d := range mensajes {
			balanceGeneral = HacerCalculos(d.Body, balanceGeneral)
		}
	}()
	SetupCloseHandler(balanceGeneral)

	fmt.Printf("Succesfully connected to logistica\n")
	fmt.Printf("esperando reportes\n")
	<-forever
}
