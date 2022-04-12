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
type DispatchUpdateCaStatusCodeRequest struct {
	CaName string `json:"ca_name"`
	CaType string `json:"ca_type"`
}
type DispatchUpdateCertStatusCodeRequest struct {
	CaName       string `json:"ca_name"`
	Status       string `json:"status"`
	SerialNumber string `json:"serial_number"`
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
	Config   []interface{} `json:"config"`
	DeviceID string        `json:"device_id"`
}
