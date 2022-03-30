import { Iot, SQS } from "aws-sdk"
import { CloudEvent } from "cloudevents"
import { Certificate } from "@fidm/x509"

const iot = new Iot()
const sqs = new SQS()

export const handler = async (event: any) => {
  console.log(event)
  const requestParameters = event.detail.requestParameters
  const newStatus: string = requestParameters.newStatus
  const certificateId: string = requestParameters.certificateId

  const describeCertResponse = await iot.describeCertificate({ certificateId: certificateId }).promise()

  console.log("caCertificateID" + describeCertResponse.certificateDescription!.caCertificateId)
  console.log("certificateId" + certificateId)

  const describeCAResponse = await iot.describeCACertificate({ certificateId: describeCertResponse.certificateDescription!.caCertificateId! }).promise()

  const caCert = Certificate.fromPEM(Buffer.from(describeCAResponse.certificateDescription!.certificatePem!, "utf8"))
  const cert = Certificate.fromPEM(Buffer.from(describeCertResponse.certificateDescription!.certificatePem!, "utf8"))

  const deviceID = cert.subject.commonName
  console.log("deviceID: [" + deviceID + "] newStatus:[" + newStatus + "]")

  const cloudEvent = new CloudEvent({
    type: "io.lamassu.iotcore.cert.updatestatus",
    id: "",
    source: "cloud-trail",
    time: new Date().toString(),
    specversion: "1.0",
    data: {
      caID: describeCertResponse.certificateDescription!.caCertificateId,
      caName: caCert.subject.commonName,
      caSerialNumber: caCert.serialNumber,
      certificateID: certificateId,
      deviceID: deviceID
    }
  })
  try {
    const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL!, MessageBody: cloudEvent.toString() }).promise()
    console.log(sqsResponse)
  } catch (err) {
    console.log("error while sending SQS messgae", err)
  }
}
