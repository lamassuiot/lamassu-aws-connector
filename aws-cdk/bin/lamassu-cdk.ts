#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { LamassuCdkStack } from '../lib/cdk-stack';

const app = new cdk.App();
const env = { account: process.env.AWS_ACCOUNT_ID, region: process.env.AWS_DEFAULT_REGION };

new LamassuCdkStack(app, 'LamassuCdkStack', {env : env});