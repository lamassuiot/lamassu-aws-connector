import * as cdk from "@aws-cdk/core"
import * as sqs from "@aws-cdk/aws-sqs"
import { LamassuEventBridge } from "./event-bridge"
// import {LamassuEventBridge} from "./event-bridge"
// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class LamassuCdkStack extends cdk.Stack {
  constructor (scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    // Lamassu response queue
    const responseQueue = new sqs.Queue(this, "lamassuResponse", {
      queueName: "lamassuResponse"
    })


    const lamassuEventBridgeInfra = new LamassuEventBridge(this, "MonitorInfra", {
      outboundSQSQueue: responseQueue
    })

    new cdk.CfnOutput(this, "Response Queue ARN", {
      value: responseQueue.queueArn,
      description: "Lamassu SQS response queue ARN"
    })
  }
}
