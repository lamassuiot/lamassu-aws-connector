package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	awsIot "github.com/aws/aws-sdk-go/service/iot"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/lamassuiot/lamassu-aws-connector/pkg/server/store"
	"github.com/lamassuiot/lamassu-aws-connector/pkg/server/utils"
	lamassucaclient "github.com/lamassuiot/lamassuiot/pkg/ca/client"
	caDTO "github.com/lamassuiot/lamassuiot/pkg/ca/common/dto"
	"golang.org/x/exp/slices"
)

type Service interface {
	//Responses received from AWS via SQS
	HandleUpdateConfiguration(ctx context.Context, config interface{}) error
	HandleUpdateThingConfiguration(ctx context.Context, deviceID string, config interface{}) error
	HandleUpdateCertificateStatus(ctx context.Context, caName string, serialNumber string, status string) error
	HandleUpdateCAStatus(ctx context.Context, caName string, caSerialNumber string, caID string, status string) error

	// HTTP methods
	Health(ctx context.Context) bool
	GetConfiguration(ctx context.Context) (interface{}, error)
	GetThingConfiguration(ctx context.Context, deviceID string) (AWSThing, error)
	SignRegistrationCode(ctx context.Context, caName string, caCert string, SerialNumber string) error
	AttachIoTCorePolicy(ctx context.Context, caName string, SerialNumber string, Policy string) error
	UpdateCAStatusRequest(ctx context.Context, caName string, status string, certificateID string) error
	UpdateCertStatusRequest(ctx context.Context, caName string, certSerialNumber string, status string, deviceCert string, caCert string) error
}

type awsService struct {
	ID              string
	logger          log.Logger
	LamassuCaClient lamassucaclient.LamassuCaClient
	db              store.DB
	awsIotSvc       *awsIot.IoT
	accountID       string
}

func NewAwsConnectorService(logger log.Logger, connectorId string, lamassuCAClient lamassucaclient.LamassuCaClient, db store.DB, awsIotSvc *awsIot.IoT, awsAccountID string) (s Service) {
	return &awsService{
		logger:          logger,
		LamassuCaClient: lamassuCAClient,
		ID:              connectorId,
		db:              db,
		awsIotSvc:       awsIotSvc,
		accountID:       awsAccountID,
	}
}

func (s *awsService) Health(ctx context.Context) bool {
	return true
}

func (s *awsService) AttachIoTCorePolicy(ctx context.Context, caName string, SerialNumber string, Policy string) error {
	listCAsResponse, err := s.awsIotSvc.ListCACertificates(&awsIot.ListCACertificatesInput{
		PageSize: aws.Int64(25),
	})

	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to list CA certificates")
		return err
	}

	for _, ca := range listCAsResponse.Certificates {
		tagsResponse, err := s.awsIotSvc.ListTagsForResource(&awsIot.ListTagsForResourceInput{
			ResourceArn: ca.CertificateArn,
		})

		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to list tags for CA certificate")
			return err
		}

		nameTagIdx := slices.IndexFunc(tagsResponse.Tags, func(tag *awsIot.Tag) bool {
			return *tag.Key == "lamassuCAName"
		})

		if nameTagIdx == -1 {
			level.Error(s.logger).Log("err", err, "msg", "Failed to find lamassuCAName field tag for CA [AWS ID]="+*ca.CertificateId)
			continue
		}

		nameTag := tagsResponse.Tags[nameTagIdx]
		if *nameTag.Value == caName {
			now := time.Now()
			policyName := "lamassu-iot-policy_" + caName + "_" + now.Format("2006-01-02_15-04-05")
			templateBody := fmt.Sprintf(
				`{"Parameters":{"AWS::IoT::Certificate::CommonName":{"Type":"String"},"AWS::IoT::Certificate::Country":{"Type":"String"},"AWS::IoT::Certificate::Id":{"Type":"String"},"AWS::IoT::Certificate::SerialNumber":{"Type":"String"}},"Resources":{"thing":{"Type":"AWS::IoT::Thing","Properties":{"ThingName":{"Ref":"AWS::IoT::Certificate::CommonName"},"ThingGroups":["LAMASSU"],"AttributePayload":{}}},"certificate":{"Type":"AWS::IoT::Certificate","Properties":{"CertificateId":{"Ref":"AWS::IoT::Certificate::Id"},"Status":"ACTIVE"}},"policy":{"Type":"AWS::IoT::Policy","Properties":{"PolicyName":"%s"}}}}`,
				policyName)

			_, err = s.awsIotSvc.CreatePolicy(&awsIot.CreatePolicyInput{
				PolicyDocument: aws.String(Policy),
				PolicyName:     aws.String(policyName),
				Tags: []*awsIot.Tag{
					{
						Key:   aws.String("serialNumber"),
						Value: aws.String(SerialNumber),
					},
				},
			})

			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to create policy")
				return err
			}

			_, err = s.awsIotSvc.UpdateCACertificate(&awsIot.UpdateCACertificateInput{
				CertificateId:             ca.CertificateId,
				NewAutoRegistrationStatus: aws.String("ENABLE"),
				NewStatus:                 aws.String("ACTIVE"),
				RegistrationConfig: &awsIot.RegistrationConfig{
					RoleArn:      aws.String("arn:aws:iam::" + s.accountID + ":role/JITPRole"),
					TemplateBody: aws.String(templateBody),
				},
				RemoveAutoRegistration: aws.Bool(false),
			})

			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to update CA certificate")
				return err
			}
		}
	}

	s.db.DeleteAWSIoTCoreConfig(ctx)

	return err
}

func (s *awsService) SignRegistrationCode(ctx context.Context, caName string, caCert string, SerialNumber string) error {
	registrationCode, err := s.awsIotSvc.GetRegistrationCode(&awsIot.GetRegistrationCodeInput{})
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to get registration code")
		return err
	}

	subj := pkix.Name{
		CommonName: *registrationCode.RegistrationCode,
	}
	rawSubj := subj.ToRDNSequence()
	asn1Subj, _ := asn1.Marshal(rawSubj)
	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	// Generate Private key for verification certificate
	privKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privKey)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to create registraion code CSR")
		return err
	}

	// Generate CSR for verification certificate
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to parse registraion code CSR")
		return err
	}

	// Sign verification certificate CSR
	crt, _, err := s.LamassuCaClient.SignCertificateRequest(ctx, caDTO.Pki, caName, csr, false, csr.Subject.CommonName)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to sign registration code.")
		return err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})
	data, _ := base64.StdEncoding.DecodeString(caCert)
	caCert = string(data)
	serialN := &awsIot.Tag{
		Key:   aws.String("serialNumber"),
		Value: aws.String(SerialNumber),
	}
	caname := &awsIot.Tag{
		Key:   aws.String("lamassuCAName"),
		Value: aws.String(caName),
	}
	tags := []*awsIot.Tag{serialN, caname}
	_, err = s.awsIotSvc.RegisterCACertificate(&awsIot.RegisterCACertificateInput{
		CaCertificate:           aws.String(caCert),
		VerificationCertificate: aws.String(string(certPEM)),
		Tags:                    tags,
	})

	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to register CA certificate")
		return err
	}

	s.db.DeleteAWSIoTCoreConfig(ctx)

	return nil
}

func (s *awsService) UpdateCAStatusRequest(ctx context.Context, caName string, status string, certificateID string) error {
	if status == "REVOKED" {
		status = "INACTIVE"
	}

	_, err := s.awsIotSvc.UpdateCACertificate(&awsIot.UpdateCACertificateInput{
		CertificateId: aws.String(certificateID),
		NewStatus:     aws.String(status),
	})
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to Update CA certificate status")
		return err
	}

	err = s.db.DeleteAWSIoTCoreConfig(ctx)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to delete AWS IOT Core config")
		return err
	}

	return nil
}

func (s *awsService) UpdateCertStatusRequest(ctx context.Context, deviceID string, certSerialNumber string, status string, deviceCert string, caCert string) error {
	searchResult, err := s.awsIotSvc.SearchIndex(&awsIot.SearchIndexInput{QueryString: aws.String("thingName:" + deviceID)})
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to search device in IoT Core")
		return err
	}

	if len(searchResult.Things) == 1 {
		thingsPrincipalResponse, err := s.awsIotSvc.ListThingPrincipals(&awsIot.ListThingPrincipalsInput{
			MaxResults: aws.Int64(int64(25)),
			ThingName:  aws.String(deviceID),
		})
		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to list things principals")
			return err
		}
		updatedThingCertifcate := false
		for _, principal := range thingsPrincipalResponse.Principals {
			splitiedPrincipal := strings.Split(*principal, ":")
			certificateID := strings.Replace(splitiedPrincipal[len(splitiedPrincipal)-1], "cert/", "", 1)
			describeCertificateResponse, err := s.awsIotSvc.DescribeCertificate(&awsIot.DescribeCertificateInput{
				CertificateId: aws.String(certificateID),
			})
			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to describe things certificate")
				return err
			}

			block, _ := pem.Decode([]byte(*describeCertificateResponse.CertificateDescription.CertificatePem))
			crt, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to decode crt")
				return err
			}

			if utils.InsertNth(utils.ToHexInt(crt.SerialNumber), 2) == certSerialNumber {
				_, err = s.awsIotSvc.UpdateCertificate(&awsIot.UpdateCertificateInput{
					CertificateId: aws.String(certificateID),
					NewStatus:     aws.String(status),
				})
				if err != nil {
					level.Error(s.logger).Log("err", err, "msg", "Failed to Update certificate status")
					return err
				}

				updatedThingCertifcate = true
			}
		}
		if !updatedThingCertifcate {
			level.Debug(s.logger).Log("msg", "The device does not have the certificate yet. Registering manually")
		}
	} else if len(searchResult.Things) > 1 {
		level.Error(s.logger).Log("msg", "Inconsistent thing repo: More than one result for [DeviceID]="+deviceID)
	} else {
		level.Debug(s.logger).Log("msg", "No results with device ID")

		_, err = s.awsIotSvc.CreateThing(&awsIot.CreateThingInput{
			ThingName: aws.String(deviceID),
		})
		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to register device in IoT Core")
			return err
		}

		registerCertificateResponse, err := s.awsIotSvc.RegisterCertificate(&awsIot.RegisterCertificateInput{
			CaCertificatePem: &caCert,
			CertificatePem:   &deviceCert,
			Status:           &status,
		})

		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to register certificate in IoT Core")
			return err
		}

		s.awsIotSvc.AttachThingPrincipal(&awsIot.AttachThingPrincipalInput{
			Principal: registerCertificateResponse.CertificateArn,
			ThingName: aws.String(deviceID),
		})
	}

	s.db.DeleteAWSIoTCoreThingConfig(ctx, deviceID)
	return nil
}

func (s *awsService) GetConfiguration(ctx context.Context) (interface{}, error) {
	endpointAddress := ""
	endpointInfo, err := s.awsIotSvc.DescribeEndpoint(&iot.DescribeEndpointInput{EndpointType: aws.String("iot:Data")})
	if err == nil {
		endpointAddress = *endpointInfo.EndpointAddress
	}

	awsCAs := make([]interface{}, 0)
	listCAsResponse, err := s.awsIotSvc.ListCACertificates(&awsIot.ListCACertificatesInput{AscendingOrder: aws.Bool(true), PageSize: aws.Int64(int64(30))})
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to search device in IoT Core")
		return make(map[string]interface{}), err
	}

	for _, ca := range listCAsResponse.Certificates {
		tagsResponse, err := s.awsIotSvc.ListTagsForResource(&awsIot.ListTagsForResourceInput{
			ResourceArn: ca.CertificateArn,
		})

		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to list tags")
			continue
		}

		nameTagIdx := slices.IndexFunc(tagsResponse.Tags, func(tag *awsIot.Tag) bool {
			return *tag.Key == "lamassuCAName"
		})

		if nameTagIdx == -1 {
			level.Error(s.logger).Log("err", err, "msg", "Failed to find lamassuCAName field tag for CA [AWS ID]="+*ca.CertificateId)
			continue
		}

		nameTag := tagsResponse.Tags[nameTagIdx]
		caDescription, err := s.awsIotSvc.DescribeCACertificate(&awsIot.DescribeCACertificateInput{
			CertificateId: ca.CertificateId,
		})

		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to describe CA certificate")
			continue
		}

		type awsCAConfig struct {
			Name           string    `json:"name"`
			ARN            string    `json:"arn"`
			ID             string    `json:"id"`
			Status         string    `json:"status"`
			CreationDate   time.Time `json:"creation_date"`
			PolicyStatus   string    `json:"policy_status"`
			PolicyName     string    `json:"policy_name,omitempty"`
			PolicyDocument string    `json:"policy_document,omitempty"`
		}

		if caDescription.RegistrationConfig != nil {
			caTemplate := *caDescription.RegistrationConfig.TemplateBody
			type parsedCaTemplateType struct {
				Resources struct {
					Policy struct {
						Properties struct {
							PolicyName string `json:"PolicyName"`
						} `json:"Properties"`
					} `json:"policy"`
				} `json:"Resources"`
			}
			var parsedCaTemplate parsedCaTemplateType
			json.Unmarshal([]byte(caTemplate), &parsedCaTemplate)
			policyName := parsedCaTemplate.Resources.Policy.Properties.PolicyName
			policyResponse, err := s.awsIotSvc.GetPolicy(&awsIot.GetPolicyInput{PolicyName: &policyName})

			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to get policy response")
				caConfig := awsCAConfig{
					Name:         *nameTag.Value,
					ARN:          *ca.CertificateArn,
					ID:           *ca.CertificateId,
					Status:       *ca.Status,
					CreationDate: *ca.CreationDate,
					PolicyName:   policyName,
					PolicyStatus: "Inconsistent",
				}
				awsCAs = append(awsCAs, caConfig)
			} else {
				caConfig := awsCAConfig{
					Name:           *nameTag.Value,
					ARN:            *ca.CertificateArn,
					ID:             *ca.CertificateId,
					Status:         *ca.Status,
					CreationDate:   *ca.CreationDate,
					PolicyName:     policyName,
					PolicyDocument: *policyResponse.PolicyDocument,
					PolicyStatus:   "Active",
				}
				awsCAs = append(awsCAs, caConfig)
			}
		} else {
			caConfig := awsCAConfig{
				Name:         *nameTag.Value,
				ARN:          *ca.CertificateArn,
				ID:           *ca.CertificateId,
				Status:       *ca.Status,
				CreationDate: *ca.CreationDate,
				PolicyStatus: "NoPolicy",
			}
			awsCAs = append(awsCAs, caConfig)
		}

	}

	type awsConfig struct {
		AwsAccountID    string        `json:"account_id"`
		IotCoreEndpoint string        `json:"iot_core_endpoint"`
		RegisteredCAs   []interface{} `json:"registered_cas"`
	}
	return &awsConfig{
		AwsAccountID:    s.accountID,
		IotCoreEndpoint: endpointAddress,
		RegisteredCAs:   awsCAs,
	}, nil

}

func (s *awsService) GetThingConfiguration(ctx context.Context, deviceID string) (AWSThing, error) {
	searchResult, err := s.awsIotSvc.SearchIndex(&awsIot.SearchIndexInput{QueryString: aws.String("thingName:" + deviceID)})
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to search device in IoT Core")
		return AWSThing{}, err
	}
	if len(searchResult.Things) == 1 {
		thingResult := searchResult.Things[0]
		var thing AWSThing
		thing.DeviceID = deviceID
		thing.Status = 200
		thing.Config.AWSThingID = *thingResult.ThingId
		thing.Config.Certificates = []AWSThingCertificate{}

		fmt.Println(*thingResult.Connectivity.Connected)
		if *thingResult.Connectivity.Connected {
			thing.Config.LastConnection = int(*thingResult.Connectivity.Timestamp)
		}

		principalResponse, err := s.awsIotSvc.ListThingPrincipals(&awsIot.ListThingPrincipalsInput{
			ThingName:  thingResult.ThingName,
			MaxResults: aws.Int64(int64(25)),
		})

		if err != nil {
			level.Error(s.logger).Log("err", err, "msg", "Failed to list thing principals")
			return thing, nil
		}

		for _, principal := range principalResponse.Principals {
			splitiedPrincipal := strings.Split(*principal, ":")
			certificateID := strings.Replace(splitiedPrincipal[len(splitiedPrincipal)-1], "cert/", "", 1)
			certificateResponse, err := s.awsIotSvc.DescribeCertificate(&awsIot.DescribeCertificateInput{CertificateId: &certificateID})
			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to describe thing cetificate")
				return thing, nil
			}

			block, _ := pem.Decode([]byte(*certificateResponse.CertificateDescription.CertificatePem))

			crt, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to decode crt")
				return thing, nil
			}

			thingCrt := AWSThingCertificate{
				ARN:          *certificateResponse.CertificateDescription.CertificateArn,
				ID:           *certificateResponse.CertificateDescription.CertificateId,
				SerialNumber: utils.InsertNth(utils.ToHexInt(crt.SerialNumber), 2),
				Status:       *certificateResponse.CertificateDescription.Status,
				UpdateDate:   *certificateResponse.CertificateDescription.LastModifiedDate,
				CaName:       crt.Issuer.CommonName,
			}
			thing.Config.Certificates = append(thing.Config.Certificates, thingCrt)
		}

		return thing, nil
	}
	return AWSThing{}, nil
}

func (s *awsService) HandleUpdateCAStatus(ctx context.Context, caName string, caSerialNumber string, caID string, status string) error {
	level.Info(s.logger).Log("msg", "invalidating config cache due to CA update. caName:"+caName+" caSerialNumber:"+caSerialNumber+" caID:"+caID+" status:"+status)
	s.db.DeleteAWSIoTCoreConfig(ctx)
	return nil
}

func (s *awsService) HandleUpdateConfiguration(ctx context.Context, config interface{}) error {
	s.db.UpdateAWSIoTCoreConfig(ctx, config)
	return nil
}

func (s *awsService) HandleUpdateThingConfiguration(ctx context.Context, deviceID string, config interface{}) error {
	s.db.UpdateAWSIoTCoreThingConfig(ctx, deviceID, config)
	return nil
}

func (s *awsService) HandleUpdateCertificateStatus(ctx context.Context, caName string, serialNumber string, status string) error {
	if status == "REVOKED" {
		err := s.LamassuCaClient.RevokeCert(ctx, caDTO.Pki, caName, serialNumber)
		if err != nil {
			switch err.(type) {
			case *lamassucaclient.AlreadyRevokedError:
				return nil
			default:
				return err
			}

		}
	}
	return nil
}
