import { Iot } from "aws-sdk"
import { CloudEvent } from "cloudevents"

const iot = new Iot()

export const handler = async (event: any) => {
  console.log(event)

  const requestedCloudEvent = new CloudEvent<any>(event)

  try {
    const response = await iot.registerCACertificate({
      caCertificate: Buffer.from(requestedCloudEvent.data.ca_cert, "base64").toString(),
      verificationCertificate: Buffer.from(requestedCloudEvent.data.verification_cert, "base64").toString(),
      setAsActive: false,
      allowAutoRegistration: true,
      tags: [{
        Key: "lamassuCAName",
        Value: requestedCloudEvent.data.ca_name
      },
      {
        Key: "serialNumber",
        Value: requestedCloudEvent.data.serial_number
      }]
    }).promise()

    console.log(response)
  } catch (err) {
    console.log("error while registerting IoT Core CA", err)
  }
}
