
mosquitto_pub --cafile ../certs/awsRootCA.pem --cert ../certs/deviceCertAndCACert.crt --key ../certs/deviceCert.key -h a3hczhtwc7h4es-ats.iot.eu-west-1.amazonaws.com -p 8883 -q 1 -t markeldemodev -i markeldemodev -m "hello" -d 

