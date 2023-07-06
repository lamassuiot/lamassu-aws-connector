import { Iot, SQS } from "aws-sdk"
import { CloudEvent } from "cloudevents"
import { Certificate } from "@fidm/x509"

const iot = new Iot()
const sqs = new SQS()

export const handler = async (event: any) => {
  console.log("Incoming event from Cloud Trail", event)
  const requestParameters = event.detail.requestParameters
  const newStatus: string = requestParameters.newStatus
  const certificateId: string = requestParameters.certificateId

  try {
    const describeCAResponse = await iot.describeCACertificate({ certificateId: certificateId }).promise()
    const caCert = Certificate.fromPEM(Buffer.from(describeCAResponse.certificateDescription!.certificatePem!, "utf8"))
    const cloudEvent = new CloudEvent({
      type: "io.lamassu.iotcore.ca.status.update",
      id: "",
      source: "aws/cloud-trail",
      time: new Date().toString(),
      specversion: "1.0",
      data: {
        ca_id: certificateId,
        ca_name: caCert.subject.commonName,
        ca_serial_number: chunk(caCert.serialNumber, 2).join("-"),
        status: newStatus
      }
    })
    try {
      const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL!, MessageBody: cloudEvent.toString() }).promise()
      console.log(sqsResponse)
    } catch (err) {
      console.log("error while sending SQS messgae", err)
    }
  } catch (err) {
    console.log("error while describing CA", err)
  }
}

function chunk(str: string, n: number) {
  const ret = []
  let i
  let len

  for (i = 0, len = str.length; i < len; i += n) {
    ret.push(str.substr(i, n))
  }

  return ret
}
