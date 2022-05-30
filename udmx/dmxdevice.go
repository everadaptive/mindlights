package udmx

type DmxDevice interface {
	Open()
	Close()
	SetChannelColor(channel uint16, value uint16)
	Render()
}
