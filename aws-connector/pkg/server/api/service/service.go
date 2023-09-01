package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsIot "github.com/aws/aws-sdk-go/service/iot"
	awsIotData "github.com/aws/aws-sdk-go/service/iotdataplane"
	"github.com/aws/aws-sdk-go/service/sqs"
	awsSts "github.com/aws/aws-sdk-go/service/sts"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	api "github.com/lamassuiot/aws-connector/pkg/common"
	"github.com/lamassuiot/aws-connector/pkg/server/store"
	"github.com/lamassuiot/aws-connector/pkg/server/utils"
	lamassuCAClient "github.com/lamassuiot/lamassuiot/pkg/ca/client"
	cProvderApi "github.com/lamassuiot/lamassuiot/pkg/cloud-provider/common/api"
	cProviderService "github.com/lamassuiot/lamassuiot/pkg/cloud-provider/server/api/service"
	lamassuDevManagerClient "github.com/lamassuiot/lamassuiot/pkg/device-manager/client"
	devApi "github.com/lamassuiot/lamassuiot/pkg/device-manager/common/api"
	lamassudmsclient "github.com/lamassuiot/lamassuiot/pkg/dms-manager/client"
	dmsApi "github.com/lamassuiot/lamassuiot/pkg/dms-manager/common/api"
	"github.com/lamassuiot/lamassuiot/pkg/v3/models"
	serviceV3 "github.com/lamassuiot/lamassuiot/pkg/v3/services"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Service interface {
	cProviderService.Service
	//Responses received from AWS via SQS
	HandleUpdateCertificateStatus(ctx context.Context, input *api.HandleUpdateCertificateStatusInput) error
	HandleUpdateCAStatus(ctx context.Context, input *api.HandleUpdateCAStatusInput) error
	HandleCloudEvents(ctx context.Context, event cloudevents.Event) error
	GetAccountID() string
	GetDefaultRegion() string
}

type awsService struct {
	ID                   string
	lamassuCAClient      serviceV3.CAService
	dmsClient            lamassudmsclient.LamassuDMSManagerClient
	devManagerClient     lamassuDevManagerClient.LamassuDeviceManagerClient
	db                   store.DB
	awsIotSvc            *awsIot.IoT
	awsIotData           *awsIotData.IoTDataPlane
	accountID            string
	accountDefaultRegion string
	sqsOutboundURL       string
	sqsSvc               *sqs.SQS
}

func NewAwsConnectorService(connectorId string, lamassuCAClient serviceV3.CAService, dmsClient lamassudmsclient.LamassuDMSManagerClient, devManagerClient lamassuDevManagerClient.LamassuDeviceManagerClient, db store.DB, awsDefaultRegion string, awsKeyID string, awsKeySecret string, awsSQSOutboundQueueName string) (Service, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(awsDefaultRegion),
		Credentials: credentials.NewStaticCredentials(awsKeyID, awsKeySecret, ""),
	}))
	sqsSvc := sqs.New(sess)
	awsSts := awsSts.New(sess, aws.NewConfig())
	awsSvc := awsIot.New(sess, aws.NewConfig())
	awsIotDataPlane := awsIotData.New(sess, aws.NewConfig())
	awsIotDataPlane.Endpoint = "https://a3penyvxwz0v8m-ats.iot.eu-west-1.amazonaws.com"
	awsIdentity, err := awsSts.GetCallerIdentity(nil)
	if err != nil {
		log.Fatal("Could not get AWS Identity: ", err)
	}
	sqsOutboundURL := "https://sqs." + awsDefaultRegion + ".amazonaws.com/" + *awsIdentity.Account + "/" + awsSQSOutboundQueueName
	return &awsService{
		lamassuCAClient:      lamassuCAClient,
		dmsClient:            dmsClient,
		devManagerClient:     devManagerClient,
		ID:                   connectorId,
		db:                   db,
		awsIotSvc:            awsSvc,
		awsIotData:           awsIotDataPlane,
		accountID:            *awsIdentity.Account,
		accountDefaultRegion: awsDefaultRegion,
		sqsOutboundURL:       sqsOutboundURL,
		sqsSvc:               sqsSvc,
	}, nil
}

func (s *awsService) Health() bool {
	return true
}

func (s *awsService) GetAccountID() string {
	return s.accountID
}

func (s *awsService) GetDefaultRegion() string {
	return s.accountDefaultRegion
}

func (s *awsService) UpdateDeviceDigitalTwinReenrollmentStatus(ctx context.Context, input *cProvderApi.UpdateDeviceDigitalTwinReenrollmentStatusInput) (*cProvderApi.UpdateDeviceDigitalTwinReenrollmentStatusOutput, error) {
	var shadowPayload api.DeviceShadowPayload
	if input.SlotID == "default" {
		shadowPayload = api.DeviceShadowPayload{
			State: api.StatePayload{
				Desired: api.Desired{
					Issuer: api.ShadowIssuer{},
					IdentityCert: api.ShadowIdentityCert{
						Rotate: input.ForceReenroll,
					},
					Slots: api.ShadowSlots{
						Devo: api.ShadowDevo{},
					},
				},
			},
		}
	} else {
		shadowPayload = api.DeviceShadowPayload{
			State: api.StatePayload{
				Desired: api.Desired{
					Issuer:       api.ShadowIssuer{},
					IdentityCert: api.ShadowIdentityCert{},
					Slots: api.ShadowSlots{
						Devo: api.ShadowDevo{
							Update: input.ForceReenroll,
						},
					},
				},
			},
		}
	}

	payloadBytes, err := json.Marshal(shadowPayload)
	if err != nil {
		log.Warn("Error Marshalling shadow payload: ", err)
	}

	device, err := s.devManagerClient.GetDeviceById(ctx, input.DeviceID)
	if err != nil {
		log.Warn("Error obtaining the device ", err)
	}
	dms, err := s.dmsClient.GetDMSByName(ctx, &dmsApi.GetDMSByNameInput{
		Name: device.DmsName,
	})
	if err != nil {
		log.Warn("Error obtaining the dms ", err)
	}
	if dms.Aws.ShadowType == dmsApi.ShadowTypeClassic {
		_, err = s.awsIotData.UpdateThingShadow(&awsIotData.UpdateThingShadowInput{
			ThingName: &input.DeviceID,
			Payload:   payloadBytes,
		})
		if err != nil {
			log.Warn("Error updating thing shadow: ", err)
		}
	} else {
		_, err = s.awsIotData.UpdateThingShadow(&awsIotData.UpdateThingShadowInput{
			ThingName:  &input.DeviceID,
			ShadowName: aws.String("lamassu-identity"),
			Payload:    payloadBytes,
		})
		if err != nil {
			log.Warn("Error updating thing shadow: ", err)

		}
	}

	return &cProvderApi.UpdateDeviceDigitalTwinReenrollmentStatusOutput{}, nil
}

func (s *awsService) UpdateConfiguration(ctx context.Context, input *cProvderApi.UpdateConfigurationInput) (*cProvderApi.UpdateConfigurationOutput, error) {
	marshalConfig, err := json.Marshal(input.Configuration)
	if err != nil {
		return &cProvderApi.UpdateConfigurationOutput{}, err
	}

	var awsConfig UpdateAWSConfiguration
	err = json.Unmarshal(marshalConfig, &awsConfig)
	if err != nil {
		return &cProvderApi.UpdateConfigurationOutput{}, fmt.Errorf("invalid configuration for AWS connector")
	}

	listCAsResponse, err := s.awsIotSvc.ListCACertificates(&awsIot.ListCACertificatesInput{
		PageSize: aws.Int64(25),
	})

	if err != nil {
		return &cProvderApi.UpdateConfigurationOutput{}, err
	}

	for _, ca := range listCAsResponse.Certificates {
		tagsResponse, err := s.awsIotSvc.ListTagsForResource(&awsIot.ListTagsForResourceInput{
			ResourceArn: ca.CertificateArn,
		})

		if err != nil {
			return &cProvderApi.UpdateConfigurationOutput{}, err
		}

		nameTagIdx := slices.IndexFunc(tagsResponse.Tags, func(tag *awsIot.Tag) bool {
			return *tag.Key == "lamassuCAName"
		})

		if nameTagIdx == -1 {
			log.Warn(fmt.Sprintf("Failed to find lamassuCAName field tag for CA [AWS ID]= %s", *ca.CertificateId))
			continue
		}

		nameTag := tagsResponse.Tags[nameTagIdx]
		if *nameTag.Value == awsConfig.CAName {
			now := time.Now()
			policyName := "lms" + strings.ReplaceAll(awsConfig.CAName, " ", "-") + "_" + now.Format("2006-01-02_15-04-05")
			templateName := "lms_" + now.Format("2006-01-02_15-04-05")
			templateBody := fmt.Sprintf(
				`{"Parameters":{"AWS::IoT::Certificate::CommonName":{"Type":"String"},"AWS::IoT::Certificate::Country":{"Type":"String"},"AWS::IoT::Certificate::Id":{"Type":"String"},"AWS::IoT::Certificate::SerialNumber":{"Type":"String"}},"Resources":{"thing":{"Type":"AWS::IoT::Thing","Properties":{"ThingName":{"Ref":"AWS::IoT::Certificate::CommonName"},"ThingGroups":["LAMASSU"],"AttributePayload":{}}},"certificate":{"Type":"AWS::IoT::Certificate","Properties":{"CertificateId":{"Ref":"AWS::IoT::Certificate::Id"},"Status":"ACTIVE"}},"policy":{"Type":"AWS::IoT::Policy","Properties":{"PolicyName":"%s"}}}}`,
				policyName)

			_, err = s.awsIotSvc.CreatePolicy(&awsIot.CreatePolicyInput{
				PolicyDocument: aws.String(awsConfig.Policy),
				PolicyName:     aws.String(policyName),
				Tags: []*awsIot.Tag{
					{
						Key:   aws.String("lamassuCAName"),
						Value: aws.String(awsConfig.CAName),
					},
				},
			})

			if err != nil {
				log.Error("could not create IoT core Policy")
				return &cProvderApi.UpdateConfigurationOutput{}, err
			}

			_, err = s.awsIotSvc.CreateProvisioningTemplate(&awsIot.CreateProvisioningTemplateInput{
				TemplateName:        aws.String(templateName),
				Enabled:             aws.Bool(true),
				Description:         aws.String("Created by AWS connector"),
				TemplateBody:        aws.String(templateBody),
				ProvisioningRoleArn: aws.String("arn:aws:iam::" + s.accountID + ":role/JITPRole"),
				Type:                aws.String("JITP"),
			})

			if err != nil {
				log.Error("could not create IoT Provisioning Template")
				return &cProvderApi.UpdateConfigurationOutput{}, err
			}

			_, err = s.awsIotSvc.UpdateCACertificate(&awsIot.UpdateCACertificateInput{
				CertificateId:             ca.CertificateId,
				NewAutoRegistrationStatus: aws.String("ENABLE"),
				NewStatus:                 aws.String("ACTIVE"),
				RegistrationConfig: &awsIot.RegistrationConfig{
					TemplateName: aws.String(templateName),
				},
				RemoveAutoRegistration: aws.Bool(false),
			})

			if err != nil {
				log.Error("could not create update CA Certificate")
				return &cProvderApi.UpdateConfigurationOutput{}, err
			}
		}
	}

	s.db.DeleteAWSIoTCoreConfig(ctx)

	return &cProvderApi.UpdateConfigurationOutput{}, nil
}

func (s *awsService) RegisterCA(ctx context.Context, input *cProvderApi.RegisterCAInput) (*cProvderApi.RegisterCAOutput, error) {
	registrationCode, err := s.awsIotSvc.GetRegistrationCode(&awsIot.GetRegistrationCodeInput{})
	if err != nil {
		return &cProvderApi.RegisterCAOutput{}, err
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
		return &cProvderApi.RegisterCAOutput{}, err
	}

	// Generate CSR for verification certificate
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return &cProvderApi.RegisterCAOutput{}, err
	}

	// Sign verification certificate CSR
	singOutput, err := s.lamassuCAClient.SignCertificate(serviceV3.SignCertificateInput{
		CAID:        input.CAName,
		CertRequest: (*models.X509CertificateRequest)(csr),
		Subject: models.Subject{
			CommonName: csr.Subject.CommonName,
		},
		SignVerbatim: false,
	})
	if err != nil {
		return &cProvderApi.RegisterCAOutput{}, err
	}

	verificationCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: singOutput.Certificate.Raw})
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: singOutput.Certificate.Raw})

	serialN := &awsIot.Tag{
		Key:   aws.String("serialNumber"),
		Value: aws.String(input.SerialNumber),
	}
	caname := &awsIot.Tag{
		Key:   aws.String("lamassuCAName"),
		Value: aws.String(input.CAName),
	}
	tags := []*awsIot.Tag{serialN, caname}
	_, err = s.awsIotSvc.RegisterCACertificate(&awsIot.RegisterCACertificateInput{
		CaCertificate:           aws.String(string(caPEM)),
		VerificationCertificate: aws.String(string(verificationCertPEM)),
		Tags:                    tags,
	})

	if err != nil {
		log.Error("Could not register CA Certificate in AWS")
		return &cProvderApi.RegisterCAOutput{}, err
	}

	s.db.DeleteAWSIoTCoreConfig(ctx)

	return &cProvderApi.RegisterCAOutput{}, nil
}

func (s *awsService) UpdateCAStatus(ctx context.Context, input *cProvderApi.UpdateCAStatusInput) (*cProvderApi.UpdateCAStatusOutput, error) {
	newStatus := input.Status
	if input.Status == "REVOKED" {
		newStatus = "INACTIVE"
	}

	awsCA := s.getAWSCAByName(input.CAName)
	if awsCA == nil {
		return &cProvderApi.UpdateCAStatusOutput{}, errors.New("CA not found in AWS IoT")
	}

	_, err := s.awsIotSvc.UpdateCACertificate(&awsIot.UpdateCACertificateInput{
		CertificateId: aws.String(*awsCA.CertificateId),
		NewStatus:     aws.String(newStatus),
	})
	if err != nil {
		return &cProvderApi.UpdateCAStatusOutput{}, err
	}

	err = s.db.DeleteAWSIoTCoreConfig(ctx)
	if err != nil {
		return &cProvderApi.UpdateCAStatusOutput{}, err
	}

	return &cProvderApi.UpdateCAStatusOutput{}, nil
}

func (s *awsService) UpdateDMSCaCerts(ctx context.Context, input *cProvderApi.UpdateDMSCaCertsInput) (*cProvderApi.UpdateDMSCaCertsOutput, error) {
	cas, err := s.dmsClient.CACerts(ctx, input.Name)
	if err != nil {
		return &cProvderApi.UpdateDMSCaCertsOutput{}, err
	}
	var caCerts []byte
	for _, ca := range cas {
		cacert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw})
		caCerts = append(caCerts, cacert...)
	}
	topic := "dt/lms/well-known/" + input.Name + "/cacerts"
	retained := true
	_, err = s.awsIotData.Publish(&awsIotData.PublishInput{
		Topic:   &topic,
		Payload: caCerts,
		Retain:  &retained,
	})

	dms, err := s.dmsClient.GetDMSByName(ctx, &dmsApi.GetDMSByNameInput{
		Name: input.Name,
	})
	if err != nil {
		log.Warn("Error obtaining the dms ", err)
	}

	_, err = s.devManagerClient.IterateDevicesbyDMSWithPredicate(ctx, &devApi.IterateDevicesByDMSWithPredicateInput{
		DmsName: input.Name,
		PredicateFunc: func(d *devApi.Device) {
			shadowPayload := api.DeviceShadowPayload{
				State: api.StatePayload{
					Desired: api.Desired{
						CaCerts: api.ShadowCaCerts{
							UpdateCaCerts: true,
						},
						Issuer:       api.ShadowIssuer{},
						IdentityCert: api.ShadowIdentityCert{},
						Slots:        api.ShadowSlots{},
					},
				},
			}
			payloadBytes, err := json.Marshal(shadowPayload)
			if err != nil {
				log.Warn("Error Marshalling shadow payload: ", err)
			}
			if dms.Aws.ShadowType == dmsApi.ShadowTypeClassic {
				_, err = s.awsIotData.UpdateThingShadow(&awsIotData.UpdateThingShadowInput{
					ThingName: &d.ID,
					Payload:   payloadBytes,
				})
				if err != nil {
					log.Warn("Error updating thing shadow: ", err)
				}
			} else {
				_, err = s.awsIotData.UpdateThingShadow(&awsIotData.UpdateThingShadowInput{
					ThingName:  &d.ID,
					ShadowName: aws.String("lamassu-identity"),
					Payload:    payloadBytes,
				})
				if err != nil {
					log.Warn("Error updating thing shadow: ", err)

				}
			}

		},
	})
	if err != nil {
		return &cProvderApi.UpdateDMSCaCertsOutput{}, err
	}
	return &cProvderApi.UpdateDMSCaCertsOutput{}, nil
}

func (s *awsService) UpdateDeviceCertificateStatus(ctx context.Context, input *cProvderApi.UpdateDeviceCertificateStatusInput) (*cProvderApi.UpdateDeviceCertificateStatusOutput, error) {
	deviceID := input.DeviceID
	splitedDeviceID := strings.Split(deviceID, ":")
	if len(splitedDeviceID) == 2 { // device is using SLOT ID
		deviceID = splitedDeviceID[1]
	}

	searchResult, err := s.awsIotSvc.SearchIndex(&awsIot.SearchIndexInput{QueryString: aws.String("thingName:" + deviceID)})
	if err != nil {
		log.Error("could not use aws iot search index: ", err)
		return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
	}

	if len(searchResult.Things) == 1 {
		thingsPrincipalResponse, err := s.awsIotSvc.ListThingPrincipals(&awsIot.ListThingPrincipalsInput{
			MaxResults: aws.Int64(int64(25)),
			ThingName:  aws.String(deviceID),
		})
		if err != nil {
			log.Error("could not list iot thing principals: ", err)
			return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
		}
		updatedThingCertifcate := false
		for _, principal := range thingsPrincipalResponse.Principals {
			splitiedPrincipal := strings.Split(*principal, ":")
			certificateID := strings.Replace(splitiedPrincipal[len(splitiedPrincipal)-1], "cert/", "", 1)
			describeCertificateResponse, err := s.awsIotSvc.DescribeCertificate(&awsIot.DescribeCertificateInput{
				CertificateId: aws.String(certificateID),
			})
			if err != nil {
				log.Error("could not describe iot certificate: ", err)
				return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
			}

			block, _ := pem.Decode([]byte(*describeCertificateResponse.CertificateDescription.CertificatePem))
			crt, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
			}

			if utils.InsertNth(utils.ToHexInt(crt.SerialNumber), 2) == input.SerialNumber {
				_, err = s.awsIotSvc.UpdateCertificate(&awsIot.UpdateCertificateInput{
					CertificateId: aws.String(certificateID),
					NewStatus:     aws.String(string(input.Status)),
				})
				if err != nil {
					log.Error("could not update iot certificate: ", err)
					return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
				}

				updatedThingCertifcate = true
			}
		}
		if !updatedThingCertifcate {
			log.Info("The device does not have the certificate yet. Registering manually")
		}
	} else if len(searchResult.Things) > 1 {
		log.Warn(fmt.Sprintf("Inconsistent thing repo: More than one result for [DeviceID]= %s", deviceID))
	} else {
		log.Info("No results with device ID")

		_, err = s.awsIotSvc.CreateThing(&awsIot.CreateThingInput{
			ThingName: aws.String(deviceID),
		})
		if err != nil {
			log.Error("could not create iot thing: ", err)
			return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
		}

		caOutput, err := s.lamassuCAClient.GetCAByID(serviceV3.GetCAByIDInput{
			CAID: input.CAName,
		})
		if err != nil {
			return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
		}

		certificateOutput, err := s.lamassuCAClient.GetCertificateBySerialNumber(serviceV3.GetCertificatesBySerialNumberInput{
			SerialNumber: input.SerialNumber,
		})
		if err != nil {
			return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
		}

		certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateOutput.Certificate.Raw})
		caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caOutput.Certificate.Certificate.Raw})

		registerCertificateResponse, err := s.awsIotSvc.RegisterCertificate(&awsIot.RegisterCertificateInput{
			CaCertificatePem: aws.String(string(caPEM)),
			CertificatePem:   aws.String(string(certificatePEM)),
			Status:           &input.Status,
		})

		if err != nil {
			log.Error("could not register iot certificate: ", err)
			return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, err
		}

		_, err = s.awsIotSvc.AttachThingPrincipal(&awsIot.AttachThingPrincipalInput{
			Principal: registerCertificateResponse.CertificateArn,
			ThingName: aws.String(deviceID),
		})

		if err != nil {
			log.Error("could not attach iot thing principal: ", err)
		}
	}

	// s.db.DeleteAWSIoTCoreThingConfig(ctx, input.DeviceID)
	return &cProvderApi.UpdateDeviceCertificateStatusOutput{}, nil
}

func (s *awsService) GetConfiguration(ctx context.Context, input *cProvderApi.GetConfigurationInput) (*cProvderApi.GetConfigurationOutput, error) {
	endpointAddress := ""
	endpointInfo, err := s.awsIotSvc.DescribeEndpoint(&awsIot.DescribeEndpointInput{EndpointType: aws.String("iot:Data")})
	if err != nil {
		log.Error("could not describe AWS Iot Endpoint:", err)
		return &cProvderApi.GetConfigurationOutput{}, err
	}

	endpointAddress = *endpointInfo.EndpointAddress

	awsCAs := make([]cProvderApi.CAConfiguration, 0)
	listCAsResponse, err := s.awsIotSvc.ListCACertificates(&awsIot.ListCACertificatesInput{AscendingOrder: aws.Bool(true), PageSize: aws.Int64(int64(30))})
	if err != nil {
		log.Error("could not list AWS Iot CA Certificates: ", err)
		return &cProvderApi.GetConfigurationOutput{}, err
	}

	for _, ca := range listCAsResponse.Certificates {
		tagsResponse, err := s.awsIotSvc.ListTagsForResource(&awsIot.ListTagsForResourceInput{
			ResourceArn: ca.CertificateArn,
		})

		if err != nil {
			log.Error(fmt.Sprintf("could not list AWS Iot tags for certifcate [%s]: ", *ca.CertificateArn), err)
			continue
		}

		nameTagIdx := slices.IndexFunc(tagsResponse.Tags, func(tag *awsIot.Tag) bool {
			return *tag.Key == "lamassuCAName"
		})

		if nameTagIdx == -1 {
			log.Warn(fmt.Sprintf("Failed to find lamassuCAName field tag for CA [AWS ID]= %s", *ca.CertificateId))
			continue
		}

		nameTag := tagsResponse.Tags[nameTagIdx]
		caDescription, err := s.awsIotSvc.DescribeCACertificate(&awsIot.DescribeCACertificateInput{
			CertificateId: ca.CertificateId,
		})

		if err != nil {
			log.Error(fmt.Sprintf("could not describe AWS Iot CA [%s]: ", *ca.CertificateId), err)
			continue
		}

		if caDescription.RegistrationConfig != nil {
			provisioningTemplateName := caDescription.RegistrationConfig.TemplateName
			pTemplate, err := s.awsIotSvc.DescribeProvisioningTemplate(&awsIot.DescribeProvisioningTemplateInput{
				TemplateName: provisioningTemplateName,
			})

			if err != nil {
				log.Error("could not describe provisioning template:", err)
				continue
			}

			caTemplate := *pTemplate.TemplateBody
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
				log.Warn("Failed to get policy response: ", err)
				caConfig := AWSCAConfiguration{
					Name:         *nameTag.Value,
					ARN:          *ca.CertificateArn,
					ID:           *ca.CertificateId,
					Status:       *ca.Status,
					CreationDate: *ca.CreationDate,
					PolicyName:   policyName,
					PolicyStatus: "Inconsistent",
				}
				awsCAs = append(awsCAs, cProvderApi.CAConfiguration{
					CAName:        *nameTag.Value,
					Configuration: caConfig,
				})
			} else {
				caConfig := AWSCAConfiguration{
					Name:           *nameTag.Value,
					ARN:            *ca.CertificateArn,
					ID:             *ca.CertificateId,
					Status:         *ca.Status,
					CreationDate:   *ca.CreationDate,
					PolicyName:     policyName,
					PolicyDocument: *policyResponse.PolicyDocument,
					PolicyStatus:   "Active",
				}
				awsCAs = append(awsCAs, cProvderApi.CAConfiguration{
					CAName:        *nameTag.Value,
					Configuration: caConfig,
				})
			}
		} else {
			caConfig := AWSCAConfiguration{
				Name:         *nameTag.Value,
				ARN:          *ca.CertificateArn,
				ID:           *ca.CertificateId,
				Status:       *ca.Status,
				CreationDate: *ca.CreationDate,
				PolicyStatus: "NoPolicy",
			}
			awsCAs = append(awsCAs, cProvderApi.CAConfiguration{
				CAName:        *nameTag.Value,
				Configuration: caConfig,
			})
		}

	}

	return &cProvderApi.GetConfigurationOutput{
		Configuration: AWSConfiguration{
			AwsAccountID:    s.accountID,
			IotCoreEndpoint: endpointAddress,
		},
		CAsConfiguration: awsCAs,
	}, nil

}

func (s *awsService) GetDeviceConfiguration(ctx context.Context, input *cProvderApi.GetDeviceConfigurationInput) (*cProvderApi.GetDeviceConfigurationOutput, error) {
	searchResult, err := s.awsIotSvc.SearchIndex(&awsIot.SearchIndexInput{QueryString: aws.String("thingName:" + input.DeviceID)})

	if err != nil && strings.Contains(err.Error(), "Index AWS_Things does not exist") {
		_, err = s.awsIotSvc.UpdateIndexingConfiguration(&awsIot.UpdateIndexingConfigurationInput{
			ThingIndexingConfiguration: &awsIot.ThingIndexingConfiguration{
				ThingIndexingMode:             aws.String("REGISTRY_AND_SHADOW"),
				ThingConnectivityIndexingMode: aws.String("STATUS"),
				DeviceDefenderIndexingMode:    aws.String("OFF"),
				NamedShadowIndexingMode:       aws.String("ON"),
				Filter: &awsIot.IndexingFilter{
					NamedShadowNames: aws.StringSlice([]string{"lamassu-identity"}),
				},
				ManagedFields: []*awsIot.Field{
					{
						Name: aws.String("connectivity.version"),
						Type: aws.String("Number"),
					},
					{
						Name: aws.String("connectivity.timestamp"),
						Type: aws.String("Number"),
					},
					{
						Name: aws.String("connectivity.connected"),
						Type: aws.String("Boolean"),
					},
					{
						Name: aws.String("thingId"),
						Type: aws.String("String"),
					},
					{
						Name: aws.String("thingName"),
						Type: aws.String("String"),
					},
					{
						Name: aws.String("registry.thingTypeName"),
						Type: aws.String("String"),
					},
					{
						Name: aws.String("registry.thingGroupNames"),
						Type: aws.String("String"),
					},
				},
			},
		})
		if err != nil {
			log.Error("could not update aws index configuration: ", err)
		}

		searchResult, err = s.awsIotSvc.SearchIndex(&awsIot.SearchIndexInput{QueryString: aws.String("thingName:" + input.DeviceID)})
		if err != nil {
			log.Error("could not use aws iot search index: ", err)
			return &cProvderApi.GetDeviceConfigurationOutput{}, err
		}
	}

	if len(searchResult.Things) == 1 {
		thingResult := searchResult.Things[0]
		var thing AWSThingConfig

		if *thingResult.Connectivity.Connected {
			thing.LastConnection = int(*thingResult.Connectivity.Timestamp)
		}

		principalResponse, err := s.awsIotSvc.ListThingPrincipals(&awsIot.ListThingPrincipalsInput{
			ThingName:  thingResult.ThingName,
			MaxResults: aws.Int64(int64(25)),
		})

		if err != nil {
			log.Error("could not list iot thing principals: ", err)
			return &cProvderApi.GetDeviceConfigurationOutput{}, err
		}

		for _, principal := range principalResponse.Principals {
			splitiedPrincipal := strings.Split(*principal, ":")
			certificateID := strings.Replace(splitiedPrincipal[len(splitiedPrincipal)-1], "cert/", "", 1)
			certificateResponse, err := s.awsIotSvc.DescribeCertificate(&awsIot.DescribeCertificateInput{CertificateId: &certificateID})
			if err != nil {
				log.Error("could not describe iot certificate: ", err)
				return &cProvderApi.GetDeviceConfigurationOutput{}, err
			}

			block, _ := pem.Decode([]byte(*certificateResponse.CertificateDescription.CertificatePem))

			crt, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return &cProvderApi.GetDeviceConfigurationOutput{}, err
			}

			thingCrt := AWSThingCertificate{
				ARN:          *certificateResponse.CertificateDescription.CertificateArn,
				ID:           *certificateResponse.CertificateDescription.CertificateId,
				SerialNumber: utils.InsertNth(utils.ToHexInt(crt.SerialNumber), 2),
				Status:       *certificateResponse.CertificateDescription.Status,
				UpdateDate:   *certificateResponse.CertificateDescription.LastModifiedDate,
				CaName:       crt.Issuer.CommonName,
			}
			thing.Certificates = append(thing.Certificates, thingCrt)
		}

		return &cProvderApi.GetDeviceConfigurationOutput{
			Configuration: thing,
		}, err
	}
	return &cProvderApi.GetDeviceConfigurationOutput{}, errors.New("Device not found")
}

func (s *awsService) HandleUpdateCAStatus(ctx context.Context, input *api.HandleUpdateCAStatusInput) error {
	log.Info(fmt.Sprintf("invalidating config cache due to CA update. caName:%s caSerialNumber:%s caID:%s status:%s", input.CaName, input.CaSerialNumber, input.CaID, input.Status))
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
func (s *awsService) HandleCloudEvents(ctx context.Context, event cloudevents.Event) error {
	body, _ := json.Marshal(event)
	msgBody := string(body)
	// switch event.Type() {
	// case "io.lamassuiot.dms.update":
	// 	var data dmsApi.DeviceManufacturingServiceSerialized
	// 	err := json.Unmarshal(event.Data(), &data)
	// 	_, err = s.UpdateDMSCaCerts(ctx, &cProvderApi.UpdateDMSCaCertsInput{
	// 		DeviceManufacturingService: data.Deserialize(),
	// 	})
	// 	return err
	// case "io.lamassuiot.dms.update-authorizedcas":
	// 	var data dmsApi.DeviceManufacturingServiceSerialized
	// 	err := json.Unmarshal(event.Data(), &data)
	// 	_, err = s.UpdateDMSCaCerts(ctx, &cProvderApi.UpdateDMSCaCertsInput{
	// 		DeviceManufacturingService: data.Deserialize(),
	// 	})
	// 	return err
	// }
	s.sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody: &msgBody,
		QueueUrl:    &s.sqsOutboundURL,
	})
	return nil
}

func (s *awsService) HandleUpdateCertificateStatus(ctx context.Context, input *api.HandleUpdateCertificateStatusInput) error {
	if input.Status == "REVOKED" {
		_, err := s.lamassuCAClient.UpdateCertificateStatus(serviceV3.UpdateCertificateStatusInput{
			SerialNumber: input.SerialNumber,
			NewStatus:    models.StatusRevoked,
		})
		if err != nil {
			switch err.(type) {
			case *lamassuCAClient.AlreadyRevokedError:
				return nil
			default:
				return err
			}

		}
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------
// ------------------------------------------------- Utils Functions -------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------

func (s *awsService) getAWSCAByName(caName string) *awsIot.CACertificate {
	listCAsResponse, err := s.awsIotSvc.ListCACertificates(&awsIot.ListCACertificatesInput{
		AscendingOrder: aws.Bool(true),
		PageSize:       aws.Int64(int64(50)),
	})
	if err != nil {
		return nil
	}

	nameTagIdx := slices.IndexFunc(listCAsResponse.Certificates, func(ca *awsIot.CACertificate) bool {
		describeOut, err := s.awsIotSvc.DescribeCACertificate(&awsIot.DescribeCACertificateInput{
			CertificateId: ca.CertificateId,
		})
		if err != nil {
			return false
		}

		derCertificate, _ := pem.Decode([]byte(*describeOut.CertificateDescription.CertificatePem))
		certificate, err := x509.ParseCertificate(derCertificate.Bytes)
		if err != nil {
			return false
		}

		return certificate.Subject.CommonName == caName
	})

	if nameTagIdx == -1 {
		return nil
	}

	return listCAsResponse.Certificates[nameTagIdx]
}
