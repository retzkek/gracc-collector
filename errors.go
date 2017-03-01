package main

// AMQPError represents an error communicating with the AMQP broker.
type AMQPError struct {
	Message string
}

func NewAMQPError(msg string) AMQPError {
	return AMQPError{Message: msg}
}

func (e AMQPError) Error() string {
	return e.Message
}

// RequestError represents an error due to an invalid request.
type RequestError struct {
	Message string
}

func NewRequestError(msg string) RequestError {
	return RequestError{Message: msg}
}

func (e RequestError) Error() string {
	return e.Message
}

// RecordError represents an error processing a record.
type RecordError struct {
	Message string
}

func NewRecordError(msg string) RecordError {
	return RecordError{Message: msg}
}

func (e RecordError) Error() string {
	return e.Message
}
