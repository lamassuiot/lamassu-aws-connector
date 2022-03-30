package client

type AWSConfig struct {
	IotCoreEndpoint string         `json:"iot_core_endpoint"`
	AccountID       string         `json:"account_id"`
	CAs             []AWSIotCoreCA `json:"registered_cas"`
}
type AWSIotCoreCA struct {
	ARN            string `json:"arn"`
	CertificateID  string `json:"id"`
	CreationDate   string `json:"creation_date"`
	CAName         string `json:"name"`
	Status         string `json:"status"`
	PolicyStatus   string `json:"policy_status,omitempty"`
	PolicyDocumnet string `json:"policy_document,omitempty"`
}
type awsCreateIotCoreCA struct {
	CaName       string `json:"ca_name"`
	CaCert       string `json:"ca_cert"`
	SerialNumber string `json:"serial_number"`
}
type awsIotCoreCAAttachPolicy struct {
	Policy       string `json:"policy"`
	CaName       string `json:"ca_name"`
	SerialNumber string `json:"serial_number"`
}
