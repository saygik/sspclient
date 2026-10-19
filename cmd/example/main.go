package main

import (
	"log"

	"github.com/saygik/sspclient/sspclient"
)

func main() {
	opt := &sspclient.Options{
		SelfSSPAdress: "(1 2)1234 5)67",
		SelfIPAdress:  "0.0.0.0",
		SelfIPPort:    0,
		DestSSPAdress: "(1 2)1234 5)68",
		DestIPAdress:  "127.0.0.1",
		DestIPPort:    1234,
	}

	client := sspclient.New(opt, "sspclient.log")

	err := client.Start(func(m sspclient.MessageType) {
		log.Printf("Получено сообщение: type=%d from=%s to=%s body=%s\n",
			m.IeType, m.SendSSPAdr, m.DestSSPAdr, string(m.Body))
	})
	if err != nil {
		log.Fatal(err)
	}

	select {}
}
