package gracc

type Record interface {
	Id() string
	ToJSON(indent string) ([]byte, error)
	Raw() []byte
}
