import { Iot } from "aws-sdk"
import { CloudEvent } from "cloudevents"
import { Certificate } from "@fidm/x509"

const iot = new Iot()

export const handler = async (event: any) => {
  console.log(event)

  const requestedCloudEvent = new CloudEvent<any>(event)

  const deviceID = requestedCloudEvent.data.device_id
  const deviceCert = Buffer.from(requestedCloudEvent.data.device_cert, "base64").toString("utf-8")
  const caCert = Buffer.from(requestedCloudEvent.data.ca_cert, "base64").toString("utf-8")
  const status = requestedCloudEvent.data.status
  const serialNumber = requestedCloudEvent.data.serial_number

  if (status === "ACTIVE" || status === "INACTIVE" || status === "REVOKED") {
    const searchResponse = await iot.searchIndex({ queryString: `thingName:${deviceID}` }).promise()
    if (searchResponse.things!.length === 0) {
      console.log("No results with device ID (" + deviceID + "). Registering device")
      await iot.createThing({ thingName: deviceID }).promise()
      const registerCertificateResponse = await iot.registerCertificate({ certificatePem: deviceCert, caCertificatePem: caCert, status: status }).promise()
      await iot.attachThingPrincipal({ thingName: deviceID, principal: registerCertificateResponse.certificateArn! }).promise()
    } else if (searchResponse.things!.length !== 1) {
      console.log("Inconsistent thing repo: More than one result for the same device ID (" + deviceID + ")")
    } else {
      const thing = searchResponse.things![0]
      const principalResponse = await iot.listThingPrincipals({ thingName: thing.thingName!, maxResults: 25 }).promise()
      const principals = principalResponse.principals

      for (const principal of principals!) {
        const splitiedPrincipal = principal.split(":")
        const certificateID = splitiedPrincipal[splitiedPrincipal.length - 1].replace("cert/", "")
        const certificateResponse = await iot.describeCertificate({ certificateId: certificateID }).promise()
        const certPem = certificateResponse.certificateDescription!.certificatePem!
        const cert = Certificate.fromPEM(Buffer.from(certPem, "utf8"))

        if (chunk(cert.serialNumber, 2).join("-") === serialNumber) {
          await iot.updateCertificate({ newStatus: status, certificateId: certificateID }).promise()
          return
        }
      }

      console.log("device does not have the certificate yet. Registering manually")
      const registerCertificateResponse = await iot.registerCertificate({ certificatePem: deviceCert, caCertificatePem: caCert, status: status }).promise()
      await iot.attachThingPrincipal({ thingName: deviceID, principal: registerCertificateResponse.certificateArn! }).promise()
    }
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
