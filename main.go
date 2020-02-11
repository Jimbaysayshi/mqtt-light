package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	rpio "github.com/stianeikeland/go-rpio"
)

var wg sync.WaitGroup

var pins = [10]rpio.Pin{
	rpio.Pin(17), //
	rpio.Pin(27), //
	rpio.Pin(22), //
	rpio.Pin(5),  //
	rpio.Pin(23), //
	rpio.Pin(6),  //
	rpio.Pin(13), //
	rpio.Pin(19), //
	rpio.Pin(26), //
	rpio.Pin(4),  //
}

type configuration struct {
	ClientID         string
	Topic            string
	ID               []string
	Pins             int
	Waitdelay        string
	FunctionPin      int
	FunctionOntime   int
	FunctionDelay    int
	FunctionRotation int
	Mqttuser         string
	Mqttpw           string
	Broker           string
}
type msgs struct {
	Deviceid string  `json:"deviceid"`
	Category string  `json:"category"`
	Value    float32 `json:"value"`
}

func configPins(cm chan map[string]int) {

	file, err := os.Open("conf.json")
	if err != nil {
		fmt.Println("Error while opening configurations: ")
		panic(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error while parsing configurations for GPIO: ")
		panic(err)
	}
	m := make(map[string]int)
	m["Pins"] = config.Pins
	m["FunctionPin"] = config.FunctionPin
	m["FunctionOntime"] = config.FunctionOntime
	m["FunctionDelay"] = config.FunctionDelay
	m["FunctionRotation"] = config.FunctionRotation

	cm <- m

}

/*
gpioMap initates GPIO.pin configuration and opens GPIO.memmap
*/
func gpioMap(ch chan []rpio.Pin, cma chan map[string]int) {
	defer func() {
		close(cma)
		close(ch)
	}()
	cm := make(chan map[string]int)
	go configPins(cm)

	if err := rpio.Open(); err != nil {
		fmt.Println("Error while opening GPIO memmap: ")
		panic(err)
	}
	m := <-cm
	v := m["Pins"]
	l := make([]rpio.Pin, 0, v)

	for i := 0; i < v; i++ {
		pins[i].Output()
		l = append(l, pins[i])
	}

	cma <- m
	ch <- l
	wg.Done()
}

func functional(m map[string]int, fpin uint8) {
	ontime := m["FunctionOntime"]
	delay := m["FunctionDelay"]
	rotation := m["FunctionRotation"]

	for i := 0; i < rotation; i++ {
		rpio.Pin(fpin).Write(0)
		time.Sleep(time.Millisecond * time.Duration(ontime))
		rpio.Pin(fpin).Write(1)
		time.Sleep(time.Second * time.Duration(delay))
	}
	wg.Done()

}

func isfunctional(l []rpio.Pin, i int, m map[string]int) bool {
	fpin := uint8(m["FunctionPin"])
	for idx, pin := range l {
		if pin == rpio.Pin(fpin) && i == idx {
			go functional(m, fpin)

			return true
		}
	}
	return false
}

/*
controlGpio iterates over configured rpio.Pins and writes 1/0
according to which respective action is passed as category
*/
func controlGpio(l []rpio.Pin, i int, m map[string]int) {
	fpin := uint8(m["FunctionPin"])
	wg.Add(1)
	if isfunctional(l, i, m) == false {
		for idx, pin := range l {
			switch idx {
			case i:
				if pin != rpio.Pin(fpin) {
					if pin.Read() != 0 {
						pin.Write(0)
					}

				}
			default:
				pin.Write(1)
			}
		}
		wg.Done()
	}
	wg.Wait()
}

/*
mainGpio is called by msgHandler with category which is converted to index
for array []rpio.Pin. Index defines the action for controlGpio.
*/
func mainGpio(client mqtt.Client, topic string, category string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from sub(): ", r)
			disconnectHandler(client, topic)
		}
	}()
	wg.Add(1)
	cma := make(chan map[string]int)
	ch := make(chan []rpio.Pin)
	go gpioMap(ch, cma)

	defer func() {
		err := rpio.Close()
		if err != nil {
			fmt.Println("Error while closing GPIO memmap: ")
			panic(err)
		}
	}()
	time.Sleep(50 * time.Millisecond)
	m := <-cma
	time.Sleep(50 * time.Millisecond)
	pins := <-ch
	wg.Wait()
	switch category {
	case "Low":
		controlGpio(pins, 0, m)
	case "Normal":
		controlGpio(pins, 1, m)
	case "High":
		controlGpio(pins, 2, m)
	case "4":
		controlGpio(pins, 3, m)
	case "5":
		controlGpio(pins, 4, m)
	case "6":
		controlGpio(pins, 5, m)
	case "7":
		controlGpio(pins, 6, m)
	case "8":
		controlGpio(pins, 7, m)
	case "9":
		controlGpio(pins, 8, m)
	case "10":
		controlGpio(pins, 9, m)
	}

}

func rec(client mqtt.Client, topic string, ID []string, w int) {

	if r := recover(); r != nil {
		fmt.Println("Recovered from sub(): ", r)
		disconnectHandler(client, topic)
	} else {
		sub(client, topic, ID, w)
	}
}

func devID(client mqtt.Client, topic string, deviceid string, category string, value float32, ID []string) {
	for _, id := range ID {
		if id == deviceid {
			fmt.Println(deviceid, category, value)
			mainGpio(client, topic, category)
		}
	}

}

func sub(client mqtt.Client, topic string, ID []string, w int) {

	var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

		var reading msgs
		err := json.Unmarshal([]byte(msg.Payload()), &reading)
		if err != nil {
			fmt.Println("Error while unmarshaling message: ")
			panic(err)
		}

		deviceid := reading.Deviceid
		category := reading.Category
		value := reading.Value
		devID(client, topic, deviceid, category, value, ID)

	}
	defer rec(client, topic, ID, w)
	fmt.Println("Waiting for message")
	if token := client.Subscribe(topic, 1, msgHandler); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	time.Sleep(time.Second * time.Duration(w))
}

func configClient(chm chan map[string]string, cha chan []string) {

	file, err := os.Open("conf.json")
	if err != nil {
		fmt.Println("Error while opening client configurations: ")
		panic(err)
	}

	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println("Error while closing client configurations ")
			panic(err)
		}
	}()
	decoder := json.NewDecoder(file)
	config := configuration{}
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error while parsing configurations for client: ")
		panic(err)
	}
	m := make(map[string]string)
	m["Topic"] = config.Topic
	//m["ID"] = config.ID
	m["Waitdelay"] = config.Waitdelay
	m["ClientID"] = config.ClientID
	m["Mqttuser"] = config.Mqttuser
	m["Mqttpw"] = config.Mqttpw
	m["Broker"] = config.Broker

	fmt.Println(config.ID)
	fmt.Println(config.Topic)
	fmt.Println(config.Pins)
	fmt.Printf("\nClientID: %s\n", config.ClientID)

	chm <- m
	cha <- config.ID

}

func disconnectHandler(client mqtt.Client, topic string) {

	defer clientHandler()
	wg.Add(1)
	go func() {
		time.Sleep(30 * time.Second)
		if token := client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
		fmt.Println("Unsubscribed")
		client.Disconnect(250)
		wg.Done()
	}()
	wg.Wait()
}

var lostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, e error) {
	fmt.Println("DISCONNECTED FROM BROKER")
	os.Exit(0)
}

func clientHandler() {

	chm := make(chan map[string]string)
	cha := make(chan []string)
	go configClient(chm, cha)

	m := <-chm
	w, err := strconv.Atoi(m["Waitdelay"])
	if err != nil {
		fmt.Println("Error while converting delay time: ", err)
		panic(err)
	}

	opts := &mqtt.ClientOptions{
		Servers:              nil,
		ClientID:             m["ClientID"],
		Username:             m["Mqttuser"],
		Password:             m["Mqttpw"],
		CleanSession:         false,
		Order:                true,
		WillEnabled:          false,
		WillTopic:            "",
		WillPayload:          nil,
		WillQos:              0,
		WillRetained:         false,
		ProtocolVersion:      4,
		KeepAlive:            20,
		PingTimeout:          10 * time.Second,
		ConnectTimeout:       30 * time.Second,
		MaxReconnectInterval: 10 * time.Minute,
		AutoReconnect:        true,
		Store:                nil,
		OnConnect:            nil,
		OnConnectionLost:     lostHandler,
		WriteTimeout:         0, // 0 represents timeout disabled
		MessageChannelDepth:  100,
		ResumeSubs:           false,
		HTTPHeaders:          make(map[string][]string),
	}

	opts.AddBroker(m["Broker"])
	client := mqtt.NewClient(opts)

	defer func(client mqtt.Client, topic string) {
		if r := recover(); r != nil {
			fmt.Println(" recovered in clientHandler, disconnecting: ", r)
			disconnectHandler(client, topic)
		}
	}(client, m["Topic"])

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	as := <-cha
	fmt.Println(" Connected")
	sub(client, m["Topic"], as, w)

}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(" Recovered to main: ", r)

			time.Sleep(time.Second * 10)
			clientHandler()
		}
	}()
	fmt.Println("Starting")
	clientHandler()

}
