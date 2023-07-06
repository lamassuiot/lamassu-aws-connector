package api

type HandleUpdateCertificateStatusInput struct {
	CaName       string
	SerialNumber string
	Status       string
}

type HandleUpdateCAStatusInput struct {
	CaName         string
	CaSerialNumber string
	CaID           string
	Status         string
}

type DeviceShadow struct {
}

type ShadowIssuer struct {
	Url string `json:"url"`
}

type ShadowCaCerts struct {
	UpdateCaCerts bool `json:"update_cacerts"`
}
type ShadowIdentityCert struct {
	Rotate bool `json:"rotate"`
}

type ShadowDevo struct {
	Update bool `json:"update"`
}
type ShadowSlots struct {
	Devo ShadowDevo `json:"devo"`
}

type Desired struct {
	CaCerts      ShadowCaCerts      `json:"ca_certs"`
	Issuer       ShadowIssuer       `json:"issuer"`
	IdentityCert ShadowIdentityCert `json:"identity_cert"`
	Slots        ShadowSlots        `json:"slots"`
}

type StatePayload struct {
	Desired Desired `json:"desired"`
}

type DeviceShadowPayload struct {
	State StatePayload `json:"state"`
}
