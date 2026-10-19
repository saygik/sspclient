package sspclient

import (
	"log"
	"strconv"
	"time"
)

// createInitMessage формирует init-сообщение (тип 3).
func (c *SspClient) createInitMessage() string {
	res := c.opt.ProtocolHeader // Заголовок STDP
	res += "048"                // длина заголовка сообщения
	res += "00000000"           // длина тела сообщения
	res += "03"                 // тип сообщения
	res += c.opt.SelfSSPAdress  // адрес отправителя
	res += c.opt.DestSSPAdress  // адрес получателя
	res += "001"                // длина обрабатываемой очереди
	res += "00"                 // управляющий байт
	return res
}

// handleStdpIe05 формирует ответ на сообщение типа 5.
func (c *SspClient) handleStdpIe05(mess *string) string {
	res := c.opt.ProtocolHeader
	res += "081000000000680"
	res += string([]rune(*mess)[19:67])
	res += "00"
	res += AddZeroesToStr(strconv.FormatInt(time.Now().Unix(), 10), 10) + "00"
	return res
}

// handleStdpIe06 формирует ответ на сообщение типа 6.
func (c *SspClient) handleStdpIe06(mess *string) string {
	res := c.opt.ProtocolHeader
	res += "08100000000068000"
	res += string([]rune(*mess)[21:67])
	res += "00"
	res += AddZeroesToStr(strconv.FormatInt(time.Now().Unix(), 10), 12)
	return res
}

// handleStdpIe08 формирует ответ на тестовое сообщение типа 8.
func (c *SspClient) handleStdpIe08(mess []rune) string {
	if len(mess) > 16 {
		mess[16] = mess[16] + 1
	}
	return string(mess)
}

// handleMessage разбирает входящее сообщение.
func (c *SspClient) handleMessage(str *string) MessageType {
	var msg MessageType
	runes := []rune(*str)
	if len(runes) < 17 {
		log.Println("handleMessage: message too short")
		return msg
	}
	msg.IeType = uint((runes[15]-48)*10 + (runes[16] - 48))

	switch msg.IeType {
	case 4:
		if len(runes) < 50 {
			return msg
		}
		msg.SendSSPAdr = string(runes[17:30])
		msg.DestSSPAdr = string(runes[30:43])
		msg.MessAtr = uint((runes[48]-48)*10 + (runes[49] - 48))
	case 5:
		if len(runes) < 67 {
			return msg
		}
		msg.SendSSPAdr = string(runes[21:34])
		msg.DestSSPAdr = string(runes[34:47])
		msg.MessAtr = uint((runes[19]-48)*10 + (runes[20] - 48))
		msg.Body = runes[67:]
		msg.Kvit = string(runes[55:67])
	case 6:
		if len(runes) < 69 {
			return msg
		}
		msg.SendSSPAdr = string(runes[21:34])
		msg.DestSSPAdr = string(runes[34:47])
		msg.MessAtr = uint((runes[67]-48)*10 + (runes[68] - 48))
		msg.Body = runes[69:]
		msg.Kvit = string(runes[55:67])
		log.Println(msg)
	case 8:
		if len(runes) < 43 {
			return msg
		}
		msg.SendSSPAdr = string(runes[17:30])
		msg.DestSSPAdr = string(runes[30:43])
	}

	log.Println("Message " + strconv.Itoa(int(msg.IeType)) + ":" + msg.SendSSPAdr + "->" + msg.DestSSPAdr)
	return msg
}
