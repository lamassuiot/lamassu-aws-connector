package service

// Define message structure for BIND_CA_AWS_POLICY --> changing on cloud_events
type JsonRPC struct {
	Method  string      `json:"method"`
	Jsonrpc string      `json:"jsonrpc"`
	Params  interface{} `json:"params"`
}

type AttachIoTCoreCAPolicyParams struct {
	CaName       string `json:"ca_name"`
	SerialNumber string `json:"serial_number"`
	Policy       string `json:"policy"`
}

type RegCodeRequestParams struct {
	CaCert string `json:"ca_cert"`
	CaName string `json:"ca_name"`
}

type CreateCAParams struct {
	CaCert           string `json:"ca_cert"`
	VerificationCert string `json:"verification_cert"`
}

type RegCodeResponse struct {
	RegistrationCode string `json:"registration_code"`
	CaName           string `json:"ca_name"`
	CaCert           string `json:"ca_cert"`
}
