import * as lambdaNodeJS from "@aws-cdk/aws-lambda-nodejs"
import * as cdk from "@aws-cdk/core"
import { SqsEventSource } from "@aws-cdk/aws-lambda-event-sources"
import * as iam from "@aws-cdk/aws-iam"
import * as path from "path"
import * as lambda from "@aws-cdk/aws-lambda"
import { IQueue } from "@aws-cdk/aws-sqs"

export interface ILamassuSQSEventHandler {
  awsAccountID: string,
  inboundSQSQueue: IQueue,
  outboundSQSQueue: IQueue,
}

export class LamassuSQSEventHandler extends cdk.Construct {
  constructor (scope: cdk.Construct, id: string, config: ILamassuSQSEventHandler) {
    super(scope, id)

    // Get Registration code lambda
    const getRegistrationCodeLambda = new lambdaNodeJS.NodejsFunction(this, "GenerateCARegCode", {
      entry: path.join(__dirname, "../resources/lambda-generate-iotcore-ca-registrationcode/index.ts"),
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

    getRegistrationCodeLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:GetRegistrationCode"
      ],
      resources: ["*"]
    }))

    getRegistrationCodeLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "sqs:*"
      ],
      resources: [config.outboundSQSQueue.queueArn]
    }))

    // Import CA lambda
    const importIoTCoreCALambda = new lambdaNodeJS.NodejsFunction(this, "ImportIoTCoreCA", {
      entry: path.join(__dirname, "../resources/lambda-import-iotcore-ca/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      }
    })

    importIoTCoreCALambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:RegisterCACertificate",
        "iot:TagResource"
      ],
      resources: ["*"]
    }))

    // Attach IoT Core Policy code lambda
    const attachIoTCorePolicyToCALambda = new lambdaNodeJS.NodejsFunction(this, "AttachIoTCorePolicy", {
      entry: path.join(__dirname, "../resources/lambda-attach-iotcore-policy-to-ca/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      },
      environment: {
        AWS_ACCOUNT_ID: config.awsAccountID
      }
    })

    attachIoTCorePolicyToCALambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:ListCACertificates",
        "iot:ListTagsForResource",
        "iot:CreatePolicy",
        "iot:TagResource",
        "iam:PassRole",
        "iot:UpdateCACertificate"
      ],
      resources: ["*"]
    }))

    const getIoTCoreConfigurationLambda = new lambdaNodeJS.NodejsFunction(this, "GetIoTCoreConfig", {
      entry: path.join(__dirname, "../resources/lambda-get-iotcore-configuration/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      },
      environment: {
        SQS_RESPONSE_QUEUE_URL: config.outboundSQSQueue.queueUrl,
        AWS_ACCOUNT_ID: config.awsAccountID
      }
    })

    getIoTCoreConfigurationLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:ListCACertificates",
        "iot:ListTagsForResource",
        "iot:DescribeEndpoint",
        "iot:DescribeCACertificate",
        "iot:GetPolicy",
        "iot:SearchIndex",
        "sqs:*"
      ],
      resources: ["*"]
    }))

    const getIoTCoreThingConfigurationLambda = new lambdaNodeJS.NodejsFunction(this, "GetIoTCoreThingConfig", {
      entry: path.join(__dirname, "../resources/lambda-get-iotcore-thing-configuration/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents",
          "@fidm/x509"
        ]
      },
      environment: {
        SQS_RESPONSE_QUEUE_URL: config.outboundSQSQueue.queueUrl
      }
    })

    getIoTCoreThingConfigurationLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:SearchIndex",
        "iot:ListThingPrincipals",
        "iot:DescribeCertificate",
        "sqs:*"
      ],
      resources: ["*"]
    }))

    const updateIotCoreCertificateStatusLambda = new lambdaNodeJS.NodejsFunction(this, "UpdateIoTCoreCertStatus", {
      entry: path.join(__dirname, "../resources/lambda-iotcore-update-certificate-status/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      }
    })

    updateIotCoreCertificateStatusLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "iot:UpdateCertificate"
      ],
      resources: ["*"]
    }))

    // Get Registration code lambda
    const sqsHandlerLambda = new lambdaNodeJS.NodejsFunction(this, "InbounSQSHandler", {
      entry: path.join(__dirname, "../resources/lambda-sqs-filter/index.ts"),
      runtime: lambda.Runtime.NODEJS_14_X,
      handler: "handler",
      bundling: {
        nodeModules: [
          "cloudevents"
        ]
      },
      environment: {
        LAMBDA_CA_REGISTRATION_INIT: getRegistrationCodeLambda.functionArn,
        LAMBDA_IMPORT_IOTCORE_CA: importIoTCoreCALambda.functionArn,
        LAMBDA_ATTACH_IOTCORE_CA_POLICY: attachIoTCorePolicyToCALambda.functionArn,
        LAMBDA_GET_IOTCORE_THING_CONFIG: getIoTCoreThingConfigurationLambda.functionArn,
        LAMBDA_GET_IOTCORE_CONFIG: getIoTCoreConfigurationLambda.functionArn,
        LAMBDA_UPDATE_IOTCORE_CERT_STATUS: updateIotCoreCertificateStatusLambda.functionArn
      }
    })

    sqsHandlerLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "sqs:*"
      ],
      resources: [config.inboundSQSQueue.queueArn]
    }))

    sqsHandlerLambda.addToRolePolicy(new iam.PolicyStatement({
      actions: [
        "lambda:InvokeFunction"
      ],
      resources: [
        getRegistrationCodeLambda.functionArn,
        importIoTCoreCALambda.functionArn,
        attachIoTCorePolicyToCALambda.functionArn,
        getIoTCoreThingConfigurationLambda.functionArn,
        getIoTCoreConfigurationLambda.functionArn,
        updateIotCoreCertificateStatusLambda.functionArn
      ]
    }))

    const sqsHandlerEventSource: SqsEventSource = new SqsEventSource(config.inboundSQSQueue)
    sqsHandlerLambda.addEventSource(sqsHandlerEventSource)
  }
}
