region="eu-west-1"

read -p "Introduce CA name: " caName
read -p "Introduce demo device name: " deviceName
read -p "Introduce AWS Policy Name (Must be unique in AWS): " policyName
echo "AWS region is set to: $region"

# read -p "Continue?(y/n)" aux

# if [[ "$aux" == "n" ]]; then
#   exit 0
# fi

mkdir test/certs

echo -e "\033[0;32mGenerating device CA private key\033[0m"
openssl genrsa -out test/certs/deviceRootCA.key 2048 
echo ""

echo -e "\033[0;32mCreating CSR for device root CA\033[0m"
openssl req -new -sha256 -key test/certs/deviceRootCA.key -nodes -out test/certs/deviceRootCA.csr -config openssl/deviceRootCA_openssl.conf -subj "/C=ES/O=$caName" 
echo ""

echo -e "\033[0;32mCreating CA certificate\033[0m"
openssl x509 -req -days 3650 -extfile openssl/deviceRootCA_openssl.conf -extensions v3_ca -in test/certs/deviceRootCA.csr -signkey test/certs/deviceRootCA.key -out test/certs/deviceRootCA.pem 
echo ""


echo -e "\033[0;32m####################################################################\033[0m"
echo -e "\033[0;32mGetting reg code from AWS $region region"
reg_code=$(aws iot get-registration-code --region $region | jq -r .registrationCode) 
echo -e "\033[0;32mREGISTRATION CODE: $reg_code"

echo -e "\033[0;32m####################################################################\033[0m"
echo ""
echo -e "\033[0;32mGeneration private key for CA verification certificate\033[0m"
openssl genrsa -out test/certs/verificationCert.key 2048 

echo -e "\033[0;32mGenerating CSR for CA verification certificate\033[0m"
openssl req -new -key test/certs/verificationCert.key -out test/certs/verificationCert.csr -subj "/C=ES/ST=Gipuzkoa/L=Arrasate/O=$caName/OU=Lamassu/CN=$reg_code" 
echo ""

echo -e "\033[0;32mGenerating CA verification certificate\033[0m"
openssl x509 -req -in test/certs/verificationCert.csr -CA test/certs/deviceRootCA.pem -CAkey test/certs/deviceRootCA.key -CAcreateserial -out test/certs/verificationCert.crt -days 500 -sha256 
echo ""

cat templates/jitp_template_pname.json | sed "s/myauxpolicy/$policyName/g" > templates/jitp_temp_aux.json

echo -e "\033[0;32mRegister policy in AWS\033[0m"
aws iot create-policy --policy-name $policyName --policy-document file://templates/policy.json


echo -e "\033[0;32mRegister device Root CA in AWS\033[0m"
aws iot register-ca-certificate --ca-certificate file://test/certs/deviceRootCA.pem --verification-cert file://test/certs/verificationCert.crt --set-as-active --allow-auto-registration --registration-config file://templates/jitp_temp_aux.json --region $region
echo ""
rm -rf templates/jitp_temp_aux.json

echo -e "\033[0;32mGenerate device private key\033[0m"
openssl genrsa -out test/certs/deviceCert.key 2048 
echo ""

echo -e "\033[0;32mGenerating CSR for device\033[0m"
openssl req -new -key test/certs/deviceCert.key -out test/certs/deviceCert.csr -subj "/C=ES/ST=Gipuzkoa/L=Arrasate/O=$caName/OU=Lamassu/CN=$deviceName" 
echo ""

echo -e "\033[0;32mGenerating device certificate\033[0m"
openssl x509 -req -in test/certs/deviceCert.csr -CA test/certs/deviceRootCA.pem -CAkey test/certs/deviceRootCA.key -CAcreateserial -out test/certs/deviceCert.crt -days 365 -sha256 
echo ""

cat test/certs/deviceCert.crt test/certs/deviceRootCA.pem > test/certs/deviceCertAndCACert.crt

echo -e "\033[0;32mDownloading AWS Root CA certificate\033[0m"
wget https://www.amazontrust.com/repository/AmazonRootCA1.pem -O test/certs/awsRootCA.pem
echo ""

