# AWS Cloud Development Kit

![AWS CDK logo](docs/img/cdk_logo.png)

[< :house:](.)
### Table of contents
* [Overview](#Overview)
* [Prerequisites](#Prerequisites)
* [Usage](#Usage)
* [References](#References)
* [Authors](#Authors)
* [License](#License)  

## Overview

The AWS Cloud Development Kit (AWS CDK) is an open-source software development framework to define your cloud application resources using familiar programming languages.

### Directory layout

`/bin/lamassu-cdk.ts` app entry point. It creates objects of classes defined in /lib.

`/lib` it contains all the stacks for our project, were AWS resources are defined.

`cdk.json` file tells the CDK Toolkit how to execute your app.

## Prerequisites

Install nodejs 14.X, npm and jq.

```
sudo apt install jq nodejs npm
```

Install aws-cdk with npm:
```
npm install -g aws-cdk@1.x
```

Verify cdk is installed:
```
cdk --version
```

Install AWS cli and configure credentials:

1. Install AWS CLI: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html
2. Configure AWS CLI:
```
aws configure
```
Provide your AWS access key ID, secret access key, and default region when prompted. You can also configure `.aws/credentials` file in your home directory instead.


## Usage


> **NOTE**: As some AWS features used in this project are still not implemented in AWS CDK, some of the steps are done via CLI.


Go to cdk project home directory:

```
cd lamassu-cdk
```

**Option 1**: Execute deploy script:

```
bash deploy-lamassu.sh
```

**Option 2**: Execute it step by step:

```
npm i

npm i --prefix resources/sqsFilterLambda 

cdk deploy --outputs-file outputs.json
aws lambda update-event-source-mapping --uuid $(cat outputs.json | jq -r .LamassuCdkStack.EventSourceMappingUUID) --filter-criteria file://filter-criteria.json
```

## Remove Lamassu resources

```
cdk destroy LamassuCdkStack
```

### Useful commands

 * `npm run build`   compile typescript to js
 * `npm run watch`   watch for changes and compile
 * `npm run test`    perform the jest unit tests
 * `cdk deploy`      deploy this stack to your default AWS account/region
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template

## References

* AWS Docs on CDK: [AWS CDK](https://docs.aws.amazon.com/cdk/v2/guide/getting_started.html)
* API reference for AWS CDK: [API reference](https://docs.aws.amazon.com/cdk/api/v1/docs/aws-construct-library.html)
## Authors

- Haritz Saiz @hsaiz – [hsaiz@ikerlan.es](mailto:hsaiz@ikerlan.es)
- Markel Orallo @morallo – [morallo@ikerlan.es](mailto:morallo@ikerlan.es)

## License


[< :house:](.)





