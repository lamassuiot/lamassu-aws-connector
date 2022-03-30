#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { LamassuCdkStack } from '../lib/cdk-stack';

const app = new cdk.App();
const envIkerlan  = { account: '345876576284', region: 'eu-west-1' };

new LamassuCdkStack(app, 'LamassuCdkStack', {env : envIkerlan});