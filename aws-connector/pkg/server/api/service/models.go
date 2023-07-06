package service

import (
	"time"
)

//-------------

type UpdateAWSConfiguration struct {
	CAName string `json:"ca_name"`
	Policy string `json:"policy"`
}

//-------------

type AWSConfiguration struct {
	AwsAccountID    string `json:"account_id"`
	IotCoreEndpoint string `json:"iot_core_endpoint"`
}

type AWSCAConfiguration struct {
	Name           string    `json:"name"`
	ARN            string    `json:"arn"`
	ID             string    `json:"id"`
	Status         string    `json:"status"`
	CreationDate   time.Time `json:"creation_date"`
	PolicyStatus   string    `json:"policy_status"`
	PolicyName     string    `json:"policy_name,omitempty"`
	PolicyDocument string    `json:"policy_document,omitempty"`
}

//-------------

type AWSThingConfig struct {
	Certificates   []AWSThingCertificate `json:"certificates"`
	LastConnection int                   `json:"last_connection"`
}

type AWSThingCertificate struct {
	ARN          string    `json:"arn"`
	ID           string    `json:"id"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
	UpdateDate   time.Time `json:"update_date"`
	CaName       string    `json:"ca_name"`
}
