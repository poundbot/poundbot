package types

// RESTError Generic REST error response
type RESTError struct {
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}
