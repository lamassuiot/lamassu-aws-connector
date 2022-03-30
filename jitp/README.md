# JITP (Just In Time Provisioning)

Code to test JITP in AWS.

## Prerequisites to execute jitp_demo.sh

- Install jq
```bash
apt install jq
```
- Configure AWS CLI: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `REGION` must be configured.

## Test

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

