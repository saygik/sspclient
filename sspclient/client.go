package sspclient

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/saygik/sspclient/dkoiencode"
)

// SspClient — клиент SSP.
type SspClient struct {
	ReadChan         chan MessageType // канал на чтение сообщений
	connectionStatus bool             // статус соединения
	conn             net.Conn         // соединение
	dialer           net.Dialer
	opt              *Options
	open             bool
}

// New создаёт нового клиента SSP.
func New(opt *Options, logfilename string) *SspClient {
	if opt.ProtocolHeader == "" {
		opt.ProtocolHeader = "STDP"
	}
	if opt.MessageBuff == 0 {
		opt.MessageBuff = 1
	}

	client := &SspClient{
		ReadChan:         make(chan MessageType, 100),
		connectionStatus: false,
		opt:              opt,
	}
	client.dialer = net.Dialer{
		LocalAddr: &net.TCPAddr{
			IP:   net.ParseIP(opt.SelfIPAdress),
			Port: opt.SelfIPPort,
		},
	}

	logfile, err := os.OpenFile(logfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logfile)

	return client
}

// GetConnectionStatus возвращает статус соединения.
func (c *SspClient) GetConnectionStatus() bool {
	return c.connectionStatus
}

// GetStatus — синоним GetConnectionStatus (для совместимости).
func (c *SspClient) GetStatus() bool {
	return c.connectionStatus
}

// SendMessageAway отправляет сообщение (тип 5).
func (c *SspClient) SendMessageAway(id int64, message *MessageType) (string, uint) {
	if !ValidSSPAdr(message.SendSSPAdr) {
		return "", 2
	}
	if !ValidSSPAdr(message.DestSSPAdr) {
		return "", 3
	}
	if len(message.Body) == 0 {
		return "", 1
	}
	if !c.connectionStatus {
		return "", 4
	}

	message.Kvit = generateKvit(id)
	messageID := generateMessageID(id)

	totalLen := 30 + len(message.Body)
	var res string
	res = c.opt.ProtocolHeader
	res += "067"
	res += AddZeroesToStr(strconv.Itoa(totalLen), 8)
	res += "05"
	res += "90"
	res += "00"
	res += message.SendSSPAdr
	res += message.DestSSPAdr
	res += AddZeroesToStr(strconv.Itoa(totalLen), 8)
	res += message.Kvit
	res += "ID=" + messageID + " " + string(message.Body)

	if err := c.sendMessage(res); err != nil {
		return "", 4
	}
	return messageID, 0
}

// sendMessage отправляет строку в кодировке ДКОИ-8.
func (c *SspClient) sendMessage(message string) error {
	temp := []rune(message)
	_, err := c.conn.Write(dkoiencode.Encode(&temp))
	return err
}

// getConnectionString формирует строку подключения.
func (c *SspClient) getConnectionString() string {
	return fmt.Sprintf("%s:%s", c.opt.DestIPAdress, strconv.Itoa(c.opt.DestIPPort))
}

// reconnect переподключается к SSP-серверу.
func (c *SspClient) reconnect() {
	var err error
	fmt.Printf("Reconnecting to SSP server %s...\n", c.opt.DestIPAdress)
	c.conn, err = c.dialer.Dial("tcp", c.getConnectionString())
	if err != nil {
		log.Println(err)
		c.connectionStatus = false
		time.Sleep(time.Second * 2)
		return
	}
	c.connectionStatus = true
}

// Start запускает клиент.
func (c *SspClient) Start(getMessage func(MessageType)) error {
	fmt.Printf("Connecting to SSP server %s...\n", c.opt.DestIPAdress)
	conn, err := c.dialer.Dial("tcp", c.getConnectionString())
	if err != nil {
		log.Println(err)
		c.connectionStatus = false
		return err
	}
	c.conn = conn
	c.connectionStatus = true
	c.open = true

	// Рутина доставки сообщений из канала пользователю.
	go func() {
		for c.open {
			d, ok := <-c.ReadChan
			if !ok {
				return
			}
			getMessage(d)
		}
	}()

	// Рутина поддержки соединения с SSP.
	go func() {
		for c.open {
			if !c.connectionStatus {
				c.reconnect()
			}
			if c.connectionStatus {
				if err := c.sendMessage(c.createInitMessage()); err != nil {
					log.Println(err)
				} else {
					c.getMessageRoutine()
				}
				c.connectionStatus = false
				if c.conn != nil {
					c.conn.Close()
				}
				go c.disconnected()
			}
		}
	}()

	return nil
}

// disconnected отправляет сообщение об отключении в канал.
func (c *SspClient) disconnected() {
	c.ReadChan <- MessageType{
		IeType:     4,
		MessAtr:    9,
		SendSSPAdr: "",
		DestSSPAdr: "",
		Kvit:       "",
		Body:       []rune(""),
	}
}

// Stop останавливает клиент.
func (c *SspClient) Stop() {
	c.open = false
	c.connectionStatus = false
	if c.conn != nil {
		c.conn.Close()
	}
}

// getMessageRoutine читает и обрабатывает входящие сообщения.
func (c *SspClient) getMessageRoutine() {
	for c.open {
		header := make([]byte, 17)
		n, err := c.conn.Read(header)
		if n == 0 || err != nil {
			log.Println(err)
			return
		}
		temp := dkoiencode.Decode(&header)

		index := strings.Index(string(temp), "STDP")
		if index < 0 {
			continue
		}

		if index > 0 {
			tempbuf := make([]byte, index)
			n, err = c.conn.Read(tempbuf)
			if n == 0 || err != nil {
				log.Println(err)
				return
			}
			header = append(header[index:], tempbuf...)
			temp = dkoiencode.Decode(&header)
		}

		if len(temp) < 17 {
			continue
		}

		lheader, _ := strconv.Atoi(string(temp[4:7]))
		lmessage, _ := strconv.Atoi(string(temp[7:15]))
		ieType, _ := strconv.Atoi(string(temp[15:]))

		var body []byte
		bodyLen := lheader + lmessage - 17
		if bodyLen > 0 {
			body = make([]byte, 0, bodyLen)
			buf := make([]byte, 1)
			for len(body) < bodyLen {
				m, err := c.conn.Read(buf)
				if m == 0 || err != nil {
					log.Println(err)
					return
				}
				body = append(body, buf...)
			}
		} else {
			body = make([]byte, 0)
		}

		fullMessage := string(temp) + string(dkoiencode.Decode(&body))

		switch ieType {
		case 4:
			log.Println("Connection established")
			c.ReadChan <- c.handleMessage(&fullMessage)
		case 5:
			log.Println("Got new message!")
			resp := c.handleStdpIe05(&fullMessage)
			c.ReadChan <- c.handleMessage(&fullMessage)
			if err := c.sendMessage(resp); err != nil {
				log.Println(err)
			}
		case 6:
			log.Println("Your message accepted!")
			c.ReadChan <- c.handleMessage(&fullMessage)
		case 7:
			log.Println("Connection Dropped")
			return
		case 8:
			resp := c.handleStdpIe08([]rune(fullMessage))
			if err := c.sendMessage(resp); err != nil {
				log.Println(err)
			}
		case 9:
			// ответ на тестовое соединение
		case 10:
			// предупреждение
		}
	}
}
