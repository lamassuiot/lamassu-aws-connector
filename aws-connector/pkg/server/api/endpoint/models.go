package endpoint

type EmptyRequest struct{}

type HealthResponse struct {
	Healthy bool  `json:"healthy,omitempty"`
	Err     error `json:"-"`
}

type DispatchAttachIoTCorePolicyRequest struct {
	Policy       string `json:"policy"`
	CaName       string `json:"ca_name"`
	SerialNumber string `json:"serial_number"`
}

type DispatchAttachIoTCorePolicyResponse struct {
}

type DispatchRegistrationCodeRequestRequest struct {
	CaName       string `json:"ca_name"`
	CaCert       string `json:"ca_cert"`
	SerialNumber string `json:"serial_number"`
}
type DispatchUpdateCaStatusRequest struct {
	CaName        string `json:"ca_name"`
	Status        string `json:"status"`
	CertificateID string `json:"certificate_id"`
}
type DispatchUpdateCertStatusRequest struct {
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
