import { Iot } from "aws-sdk"
import { CloudEvent } from "cloudevents"

const iot = new Iot()

export const handler = async (event: any) => {
  console.log(event)

  const requestedCloudEvent = new CloudEvent<any>(event)
  const newRequestedStatus = requestedCloudEvent.data.new_status

  if (newRequestedStatus === "ACTIVE" || newRequestedStatus === "INACTIVE") {
    await iot.updateCACertificate({ newStatus: newRequestedStatus, certificateId: requestedCloudEvent.data.certificate_id }).promise()
  }
}
