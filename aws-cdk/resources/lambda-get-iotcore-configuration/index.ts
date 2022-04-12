import { Iot, SQS } from "aws-sdk"
import { CloudEvent } from "cloudevents"

const iot = new Iot()
const sqs = new SQS()

export const handler = async (event: any) => {
  const requestedCloudEvent = new CloudEvent<any>(event)

  const endpointInfo = await iot.describeEndpoint({ endpointType: "iot:Data" }).promise()
  // endpointInfo.endpointAddress

  const caListEventData = []
  try {
    const listCAsResponse = await iot.listCACertificates({ ascendingOrder: true, pageSize: 30 }).promise()
    if (listCAsResponse.certificates !== undefined) {
      for (let i = 0; i < listCAsResponse.certificates.length; i++) {
        const ca = listCAsResponse.certificates[i]
        if (ca.certificateId !== undefined && ca.certificateArn) {
          console.log(ca.certificateId)
          try {
            const tagsResponse = await iot.listTagsForResource({ resourceArn: ca.certificateArn }).promise()
            console.log(tagsResponse)
            if (tagsResponse.tags !== undefined) {
              const nameTag = tagsResponse.tags.find(tag => tag.Key === "lamassuCAName")
              if (nameTag !== undefined) {
                const caDescription = await iot.describeCACertificate({ certificateId: ca.certificateId }).promise()
                console.log(caDescription.registrationConfig)
                if (caDescription.registrationConfig !== undefined && caDescription.registrationConfig.templateBody) {
                  const caTemplate = JSON.parse(caDescription.registrationConfig.templateBody)
                  const policyName = caTemplate.Resources.policy.Properties.PolicyName
                  if (policyName !== undefined) {
                    const policyResponse = await iot.getPolicy({ policyName: policyName }).promise()
                    if (policyResponse.policyDocument) {
                      caListEventData.push({ name: nameTag.Value, arn: ca.certificateArn, id: ca.certificateId, status: ca.status, creation_date: ca.creationDate, policy_name: policyName, policy_status: "Active", policy_document: policyResponse.policyDocument })
                    }
                  } else {
                    caListEventData.push({ name: nameTag.Value, arn: ca.certificateArn, id: ca.certificateId, status: ca.status, creation_date: ca.creationDate, policy_status: "Inconsistent" })
                  }
                } else {
                  caListEventData.push({ name: nameTag.Value, arn: ca.certificateArn, id: ca.certificateId, status: ca.status, creation_date: ca.creationDate, policy_status: "NoPolicy" })
                }
              }
            }
          } catch (err) {
            console.log("error while geting information for CA [" + ca.certificateId + "]", err)
          }
        }
      }
    }
  } catch (err) {
    console.log("error while lisiting CA certificates", err)
  }

  const cloudEvent = new CloudEvent({
    type: "io.lamassu.iotcore.config.response",
    id: requestedCloudEvent.id,
    source: "aws/lambda",
    time: new Date().toString(),
    specversion: "1.0",
    data: {
      iot_core_endpoint: endpointInfo.endpointAddress,
      account_id: process.env.AWS_ACCOUNT_ID,
      registered_cas: caListEventData
    }
  })
  try {
    console.log("Cloud Event", cloudEvent)
    const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL!, MessageBody: cloudEvent.toString() }).promise()
    console.log(sqsResponse)
  } catch (err) {
    console.log("error while sending SQS messgae", err)
  }
}
