# JITP (Just In Time Provisioning)

[< :house:](.)

### Table of contents
* [Introduction](#Overview)
* [Prerequisites](#Prerequisites)
* [Usage](#Usage)
* [References](#References)
* [Authors](#Authors)
* [License](##License)  

## Overview

You can have your devices provisioned when they first attempt to connect to AWS IoT with ***just-in-time provisioning (JITP)***. To provision the device, you must enable automatic registration and associate a provisioning template with the CA certificate used to sign the device certificate.

When a device attempts to connect to AWS IoT by using a certificate signed by a registered CA certificate, AWS IoT loads the template from the CA certificate and uses it to register the thing. The JITP workflow first registers a certificate with a status value of PENDING_ACTIVATION. When the device provisioning flow is complete, the status of the certificate is changed to ACTIVE.

## Prerequisites

To perform the tests in we need the following prerequisites:

- Install jq and mosquitto-clients
```bash
apt install jq
apt install mosquitto-clients
```
- Configure AWS CLI: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `REGION` must be configured. [Configure AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)

## Flow

1. First we need to create a certificate authority and upload it to AWS.
2. Attach a JITP template to the CA in AWS IoT Core. 
3. Enable auto-registration for that CA in AWS
4. Issue a certificate with that CA.
5. Load the certificate in the device
6. Connect to the IoT Core Enpoint using that certificate.

> ***NOTE**** : For more information about each step, visit [Detailed flow](docs/jitp/detailed_jitp.md)

The device will automatically register in AWS IoT Core with the corresponding certificate and the certificate would be attached to the policy stated in the JITP template. Now the device can interact with AWS.

  > :warning: ***NOTE**** : Device fails to connect to AWS on first attempt. 


## Usage

Steps to reproduce:

 1. Create a CA certificate in AWS.
 ```bash
 bash jitp_demo.sh
 ```
 Fill the following inputs and check that your CA has been correctly created in AWS IoT Core.
 ```bash
Introduce CA name: 
Introduce demo device name:
Introduce AWS Policy Name (Must be unique in AWS):
 ```
 This script covers steps 1 to 4 in JITP flow.

 2. As a result of `jitp_demo.sh` script some test certificates have been created under `test/certs` folder. You can execute `test/go/main.go` or `test/bash/mosquitto.sh` to test connection to AWS IoT Core.
 > :warning: **If you are using mosquitto.sh**: you need to change -i flag to be equal to the certificate common name and -t flag (topic) must be "**common_name**/*" for the device to connect to AWS. For example, if CN=demodev:
 ```bash
mosquitto_pub --cafile ../certs/awsRootCA.pem  \
              --cert ../certs/deviceCertAndCACert.crt \
              --key ../certs/deviceCert.key \
              -h a3hczhtwc7h4es-ats.iot.eu-west-1.amazonaws.com \
              -p 8883 -q 1 \
              -t demodev -i demodev -m "hello" -d 
 ```

  > :warning: **If you are using go main.go**: you need to change device Id and topic. Lines 103 and 118.
  ```go
  opts.SetClientID("demodev").SetTLSConfig(tlsconfig)
  ...
  c.Publish("demodev/hello", 0, false, text)
  ```

## References

* ***AWS JITP example***: [JITP demo example](https://aws.amazon.com/es/blogs/iot/setting-up-just-in-time-provisioning-with-aws-iot-core/)
* ***JITP official documentation***: [JITP documentation](https://docs.aws.amazon.com/iot/latest/developerguide/jit-provisioning.html)

## Authors

- Haritz Saiz @hsaiz – [hsaiz@ikerlan.es](mailto:hsaiz@ikerlan.es)
- Markel Orallo @morallo – [morallo@ikerlan.es](mailto:morallo@ikerlan.es)

## License


[< :house:](.)