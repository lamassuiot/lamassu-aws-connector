import { Iot, SQS } from "aws-sdk"
import { CloudEvent } from "cloudevents"

const iot = new Iot()
const sqs = new SQS()

export const handler = async (event: any) => {
  console.log(event)

  const requestedCloudEvent = new CloudEvent<any>(event)

  const registrationCodeResponse = await iot.getRegistrationCode({}).promise()
  const registrationCodeChallenge = registrationCodeResponse.registrationCode
  console.log("Registration code:", registrationCodeChallenge)

  const cloudEvent = new CloudEvent({
    type: "io.lamassu.iotcore.ca.registration.reg-code",
    id: requestedCloudEvent.id,
    source: "aws/lambda",
    time: new Date().toString(),
    specversion: "1.0",
    data: {
      ca_name: requestedCloudEvent.data.ca_name,
      ca_cert: requestedCloudEvent.data.ca_cert,
      serial_number: requestedCloudEvent.data.serial_number,
      registration_code: registrationCodeChallenge
    }
  })
  try {
    const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL!, MessageBody: cloudEvent.toString() }).promise()
    console.log(sqsResponse)
  } catch (err) {
    console.log("error while sending SQS messgae", err)
  }
}
