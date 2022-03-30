import { SQSEvent, Context } from "aws-lambda"
import { Lambda } from "aws-sdk"
import { CloudEvent } from "cloudevents"

const lambda = new Lambda()

export const handler = async (event: SQSEvent, context: Context) => {
  try {
    for (const record of event.Records) {
      console.log("Message Body -->  ", record.body)
      const cloudEvent = new CloudEvent(JSON.parse(record.body))

      switch (cloudEvent.type) {
        case "io.lamassu.iotcore.ca.registration.init":
          await execLambda(process.env.LAMBDA_CA_REGISTRATION_INIT!, cloudEvent)
          break
        case "io.lamassu.iotcore.ca.registration.signed_challenge":
          await execLambda(process.env.LAMBDA_IMPORT_IOTCORE_CA!, cloudEvent)
          break
        case "io.lamassu.iotcore.ca.policy.attach":
          await execLambda(process.env.LAMBDA_ATTACH_IOTCORE_CA_POLICY!, cloudEvent)
          break
        case "io.lamassu.iotcore.config.request":
          await execLambda(process.env.LAMBDA_GET_IOTCORE_CONFIG!, cloudEvent)
          break

        default:
          break
      }
    }
  } catch (error) {
    console.log(error)
  }
}

const execLambda = async (functionName: string | undefined, event: CloudEvent) => {
  if (typeof functionName === "string") {
    return lambda.invoke({
      FunctionName: functionName,
      Payload: JSON.stringify(event),
      LogType: "None" // pass params
    }).promise()
  } else {
    throw Error("missing lambda name")
  }
}
