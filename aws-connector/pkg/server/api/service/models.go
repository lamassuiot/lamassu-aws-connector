package service

import "time"

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

type AWSThingConfig struct {
	AWSThingID     string                `json:"aws_id"`
	Certificates   []AWSThingCertificate `json:"certificates"`
	DeviceID       string                `json:"device_id"`
	LastConnection int                   `json:"last_connection"`
}

type AWSThingCertificate struct {
	ARN          string    `json:"arn"`
	ID           string    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
	UpdateDate   time.Time `json:"update_date"`
}
