package sspclient

// Options — параметры подключения к SSP-серверу.
type Options struct {
	SelfSSPAdress  string // Автоответ КИ
	SelfIPAdress   string // IP-адрес сервера на котором запускается драйвер
	SelfIPPort     int    // Порт с которого нужно устанавливать соединение
	DestSSPAdress  string // Автоответ целевого КИ
	DestIPAdress   string // IP-адрес целевого КИ
	DestIPPort     int    // Порт целевого КИ
	SizeOfBuff     uint   // Буфер отправки сообщений (протокол ССП)
	ProtocolHeader string // Заголовок протокола (по умолчанию "STDP")
	MessageBuff    uint   // Буфер канала сообщений
}
