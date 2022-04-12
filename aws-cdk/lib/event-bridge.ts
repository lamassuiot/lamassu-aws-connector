import * as lambdaNodeJS from "@aws-cdk/aws-lambda-nodejs"
import * as cdk from "@aws-cdk/core"
import * as lambda from "@aws-cdk/aws-lambda"
import { IQueue } from "@aws-cdk/aws-sqs"
import * as targets from "@aws-cdk/aws-events-targets"
import * as s3 from "@aws-cdk/aws-s3"
import * as cloudtrail from "@aws-cdk/aws-cloudtrail"
import * as path from "path"
import { Duration } from "@aws-cdk/core"
import * as iam from "@aws-cdk/aws-iam"

export interface ILamassuEventBridge {
  outboundSQSQueue: IQueue,
}

export class LamassuEventBridge extends cdk.Construct {
  constructor (scope: cdk.Construct, id: string, config: ILamassuEventBridge) {
    super(scope, id)

    const cloudTrailLogsBbucket = new s3.Bucket(this, "CloudTrailLogs", {
      lifecycleRules: [
        {
          expiration: Duration.days(30),
          id: "Expire object created with a lifespan of 30 days"
        }
      ]
    })

    const trail = new cloudtrail.Trail(this, "CloudTrail", {
      bucket: cloudTrailLogsBbucket
    })

    const monitorIotCoreCertUpdate = new lambdaNodeJS.NodejsFunction(this, "IotCoreUpdateCertificateStatus", {
      entry: path.join(__dirname, "../resources/lambda-notify-iotcore-update-certificate/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      },
      environment: {
        SQS_RESPONSE_QUEUE_URL: config.outboundSQSQueue.queueUrl
      }
    })

    monitorIotCoreCertUpdate.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:DescribeCertificate",
        "iot:DescribeCACertificate",
        "sqs:*"
      ],
      resources: ["*"]
    }))

    const iotCertUpdateEvent = cloudtrail.Trail.onEvent(this, "IotCoreCertRevokeMonitor", {
      target: new targets.LambdaFunction(monitorIotCoreCertUpdate)
    })

    iotCertUpdateEvent.addEventPattern({
      source: ["aws.iot"],
      detail: {
        eventSource: ["iot.amazonaws.com"],
        eventName: ["UpdateCertificate"]
      }
    })

    const monitorIotCoreCACertUpdate = new lambdaNodeJS.NodejsFunction(this, "IotCoreUpdatCACertificateStatus", {
      entry: path.join(__dirname, "../resources/lambda-notify-iotcore-update-ca-certificate/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      },
      environment: {
        SQS_RESPONSE_QUEUE_URL: config.outboundSQSQueue.queueUrl
      }
    })

    monitorIotCoreCACertUpdate.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:DescribeCACertificate",
        "sqs:*"
      ],
      resources: ["*"]
    }))

    const iotCACertUpdateEvent = cloudtrail.Trail.onEvent(this, "IotCoreCACertRevokeMonitor", {
      target: new targets.LambdaFunction(monitorIotCoreCACertUpdate)
    })

    iotCACertUpdateEvent.addEventPattern({
      source: ["aws.iot"],
      detail: {
        eventSource: ["iot.amazonaws.com"],
        eventName: ["UpdateCACertificate"]
      }
    })
  }
}
