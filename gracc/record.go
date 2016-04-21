package gracc

type Record interface {
	Id() string
	Flatten() map[string]string
	Raw() []byte
}
