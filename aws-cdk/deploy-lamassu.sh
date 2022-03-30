stackname=LamassuCdkStackHsaiz

cdk deploy $stackname --outputs-file outputs.json

aws lambda update-event-source-mapping --uuid $(cat outputs.json | jq -r .LamassuCdkStack.getRegistrationCodeEventSource) --filter-criteria file://filter_criterias/reg_code.json
aws lambda update-event-source-mapping --uuid $(cat outputs.json | jq -r .LamassuCdkStack.createCaEventSource) --filter-criteria file://filter_criterias/create_ca.json
aws lambda update-event-source-mapping --uuid $(cat outputs.json | jq -r .LamassuCdkStack.bindCaAwsPolicyEventSource) --filter-criteria file://filter_criterias/bind_ca_aws_policy.json