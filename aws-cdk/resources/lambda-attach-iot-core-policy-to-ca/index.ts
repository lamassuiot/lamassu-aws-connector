import { Iot } from "aws-sdk"
import { CloudEvent } from "cloudevents"

const iot = new Iot()

const returnIoTCoreCATemplate = (policyName: string) => {
  const templateBody = {
    Parameters: {
      "AWS::IoT::Certificate::CommonName": {
        Type: "String"
      },
      "AWS::IoT::Certificate::Country": {
        Type: "String"
      },
      "AWS::IoT::Certificate::Id": {
        Type: "String"
      },
      "AWS::IoT::Certificate::SerialNumber": {
        Type: "String"
      }
    },
    Resources: {
      thing: {
        Type: "AWS::IoT::Thing",
        Properties: {
          ThingName: {
            Ref: "AWS::IoT::Certificate::CommonName"
          },
          AttributePayload: {}
        }
      },
      certificate: {
        Type: "AWS::IoT::Certificate",
        Properties: {
          CertificateId: {
            Ref: "AWS::IoT::Certificate::Id"
          },
          Status: "ACTIVE"
        }
      },
      policy: {
        Type: "AWS::IoT::Policy",
        Properties: {
          PolicyName: policyName
        }
      }
    }
  }
  return templateBody
}

export const handler = async (event: any) => {
  console.log(event)

  const requestedCloudEvent = new CloudEvent<any>(event)
  try {
    const listCAsResponse = await iot.listCACertificates({ ascendingOrder: true, pageSize: 30 }).promise()
    console.log(listCAsResponse.certificates)
    if (listCAsResponse.certificates) {
      for (let i = 0; i < listCAsResponse.certificates.length; i++) {
        const ca = listCAsResponse.certificates[i]
        if (ca.certificateArn !== undefined && ca.certificateId !== undefined) {
          try {
            const tagsResponse = await iot.listTagsForResource({ resourceArn: ca.certificateArn }).promise()
            if (tagsResponse.tags) {
              const nameTag = tagsResponse.tags.find(tag => tag.Key === "lamassuCAName")
              if (nameTag && nameTag.Value === requestedCloudEvent.data.ca_name) {
                const policyName = "lamassulambdapolicy_" + requestedCloudEvent.data.ca_name + "_" + Date.now()
                try {
                  await iot.createPolicy({
                    policyDocument: requestedCloudEvent.data.policy,
                    policyName: policyName,
                    tags: [{
                      Key: "serialNumber",
                      Value: requestedCloudEvent.data.serial_number
                    }]
                  }).promise()

                  try {
                    await iot.updateCACertificate({
                      certificateId: ca.certificateId,
                      newAutoRegistrationStatus: "ENABLE",
                      newStatus: "ACTIVE",
                      registrationConfig: {
                        roleArn: `arn:aws:iam::${process.env.AWS_ACCOUNT_ID}:role/JITPRole`,
                        templateBody: JSON.stringify(returnIoTCoreCATemplate(policyName))
                      },
                      removeAutoRegistration: false
                    }).promise()
                  } catch (err) {
                    console.log("error while updating IoT Core CA [" + ca.certificateId + "] from IoT Core", err)
                  }
                } catch (err) {
                  console.log("error while creating IoT Core Policy for CA [" + ca.certificateId + "] from IoT Core", err)
                }
              }
            }
          } catch (err) {
            console.log("error while listing tags for CA [" + ca.certificateId + "] from IoT Core", err)
          }
        } else {
          console.log("b")

          throw new Error("CA certificate ID is null")
        }
      }
    }
  } catch (err) {
    console.log("error while listing CAs from IoT Core", err)
  }
}
