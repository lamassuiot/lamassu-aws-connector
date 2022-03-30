import * as cdk from "@aws-cdk/core"
import * as sqs from "@aws-cdk/aws-sqs"
import { LamassuSQSEventHandler } from "./sqs-cloudevents-handlers"
import { LamassuEventBridge } from "./event-bridge"
// import {LamassuEventBridge} from "./event-bridge"
// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class LamassuCdkStack extends cdk.Stack {
  constructor (scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props)

    // Lamassu command queue
    const commandQueue: sqs.Queue = new sqs.Queue(this, "lamassuCommand", {
      queueName: "lamassuCommand"
    })

    // Lamassu response queue
    const responseQueue = new sqs.Queue(this, "lamassuResponse", {
      queueName: "lamassuResponse"
    })

    const eventHandlers = new LamassuSQSEventHandler(this, "SqsEventHandler", {
      awsAccountID: this.account,
      inboundSQSQueue: commandQueue,
      outboundSQSQueue: responseQueue
    })

    const lamassuEventBridgeInfra = new LamassuEventBridge(this, "MonitorInfra", {
      outboundSQSQueue: responseQueue
    })

    // CDK stack outputs
    new cdk.CfnOutput(this, "Command Queue ARN", {
      value: commandQueue.queueArn,
      description: "Lamassu SQS command queue ARN"
    })

    new cdk.CfnOutput(this, "Response Queue ARN", {
      value: responseQueue.queueArn,
      description: "Lamassu SQS response queue ARN"
    })
  }
}
