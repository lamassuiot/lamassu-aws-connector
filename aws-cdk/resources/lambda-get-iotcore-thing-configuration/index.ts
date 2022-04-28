import { Iot, SQS } from "aws-sdk"
import { CloudEvent } from "cloudevents"
import { Certificate } from "@fidm/x509"

const iot = new Iot()
const sqs = new SQS()

export const handler = async (event: any) => {
  const requestedCloudEvent = new CloudEvent<any>(event)
  const deviceID = requestedCloudEvent.data.device_id
  console.log(deviceID)

  let thingResult: undefined | any
  const searchResponse = await iot.searchIndex({ queryString: `thingName:${deviceID}` }).promise()
  if (searchResponse.things!.length === 0) {
    console.log("No results with device ID (" + deviceID + ")")
    thingResult = {
      device_id: deviceID,
      status: 404
    }
  } else if (searchResponse.things!.length !== 1) {
    console.log("Inconsistent thing repo: More than one result for the same device ID (" + deviceID + ")")
    thingResult = {
      device_id: deviceID,
      status: 409
    }
  } else {
    const thing = searchResponse.things![0]
    const principalResponse = await iot.listThingPrincipals({ thingName: thing.thingName!, maxResults: 25 }).promise()
    const certificates = []
    const principals = principalResponse.principals
    for (const principal of principals!) {
      const splitiedPrincipal = principal.split(":")
      const certificateID = splitiedPrincipal[splitiedPrincipal.length - 1].replace("cert/", "")
      const certificateResponse = await iot.describeCertificate({ certificateId: certificateID }).promise()
      const certPem = certificateResponse.certificateDescription!.certificatePem!

      const cert = Certificate.fromPEM(Buffer.from(certPem, "utf8"))

      certificates.push({
        serial_number: chunk(cert.serialNumber, 2).join("-"),
        status: certificateResponse.certificateDescription!.status,
        arn: certificateResponse.certificateDescription!.certificateArn,
        id: certificateResponse.certificateDescription!.certificateId,
        update_date: certificateResponse.certificateDescription!.lastModifiedDate,
        ca_name: cert.issuer.commonName
      })
    }

    thingResult = {
      device_id: thing.thingName,
      status: 200,
      config: {
        aws_id: thing.thingId,
        last_connection: thing.connectivity?.timestamp,
        certificates: certificates
      }
    }
  }

  const cloudEvent = new CloudEvent({
    type: "io.lamassu.iotcore.thing.config.response",
    id: requestedCloudEvent.id,
    source: "aws/lambda",
    time: new Date().toString(),
    specversion: "1.0",
    data: thingResult
  })
  try {
    console.log("Cloud Event", cloudEvent)
    const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL!, MessageBody: cloudEvent.toString() }).promise()
    console.log(sqsResponse)
  } catch (err) {
    console.log("error while sending SQS messgae", err)
  }
}

function chunk (str: string, n: number) {
  const ret = []
  let i
  let len

  for (i = 0, len = str.length; i < len; i += n) {
    ret.push(str.substr(i, n))
  }

  return ret
}
