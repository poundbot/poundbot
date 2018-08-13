package types

// RESTError Generic REST error response
type RESTError struct {
	StatusCode int
	Error      string
}
