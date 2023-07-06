package endpoint

type AttachIoTCorePolicyRequest struct {
	Policy       string `json:"policy"`
	CaName       string `json:"ca_name"`
	SerialNumber string `json:"serial_number"`
}

type AttachIoTCorePolicyResponse struct {
}

type RegistrationCodeRequestRequest struct {
	CaName       string `json:"ca_name"`
	CaCert       string `json:"ca_cert"`
	SerialNumber string `json:"serial_number"`
}
type UpdateCaStatusRequest struct {
	CaName        string `json:"ca_name"`
	Status        string `json:"status"`
	CertificateID string `json:"certificate_id"`
}
type UpdateCertStatusRequest struct {
	DeviceID     string `json:"device_id"`
	Status       string `json:"status"`
	SerialNumber string `json:"serial_number"`
	DeviceCert   string `json:"device_cert"`
	CaCert       string `json:"ca_cert"`
}
type SignRegistrationCodeRequest struct {
	CaName           string `json:"ca_name"`
	CaCert           string `json:"ca_cert"`
	SerialNumber     string `json:"serial_number"`
	RegistrationCode string `json:"registration_code"`
}
type CreateCAResponse struct {
}

type GetConfigurationResponse struct {
	Config interface{}
}

type GetThingConfigurationRequest struct {
	DeviceID string
}
type GetThingConfigurationResponse struct {
	ThingConfig interface{}
}

type UpdateConfigurationRequest struct {
	Config map[string]interface{}
}

type UpdateThingConfigurationRequest struct {
	Config   interface{}
	DeviceID string
}

type HandleUpdateCertStatusCodeRequest struct {
	CaName       string `json:"ca_name"`
	Status       string `json:"status"`
	SerialNumber string `json:"serial_number"`
}

type HandleUpdateCAStatusCodeRequest struct {
	CaID           string `json:"ca_id"`
	CaName         string `json:"ca_name"`
	CaSerialNumber string `json:"ca_serial_number"`
	Status         string `json:"status"`
}
