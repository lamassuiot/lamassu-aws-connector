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

type UpdateConfigurationRequest struct {
	Config map[string]interface{}
}
