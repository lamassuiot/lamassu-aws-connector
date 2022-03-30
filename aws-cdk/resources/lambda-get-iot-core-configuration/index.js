"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handler = void 0;
const aws_sdk_1 = require("aws-sdk");
const cloudevents_1 = require("cloudevents");
const iot = new aws_sdk_1.Iot();
const sqs = new aws_sdk_1.SQS();
exports.handler = async (event) => {
    const requestedCloudEvent = new cloudevents_1.CloudEvent(event);
    const endpointInfo = await iot.describeEndpoint({ endpointType: "iot:Data" }).promise();
    // endpointInfo.endpointAddress
    const caListEventData = [];
    try {
        const listCAsResponse = await iot.listCACertificates({ ascendingOrder: true, pageSize: 30 }).promise();
        if (listCAsResponse.certificates !== undefined) {
            for (let i = 0; i < listCAsResponse.certificates.length; i++) {
                const ca = listCAsResponse.certificates[i];
                if (ca.certificateId !== undefined && ca.certificateArn) {
                    console.log(ca.certificateId);
                    try {
                        const tagsResponse = await iot.listTagsForResource({ resourceArn: ca.certificateArn }).promise();
                        console.log(tagsResponse);
                        if (tagsResponse.tags !== undefined) {
                            const nameTag = tagsResponse.tags.find(tag => tag.Key === "lamassuCAName");
                            if (nameTag !== undefined) {
                                const caDescription = await iot.describeCACertificate({ certificateId: ca.certificateId }).promise();
                                console.log(caDescription.registrationConfig);
                                if (caDescription.registrationConfig !== undefined && caDescription.registrationConfig.templateBody) {
                                    const caTemplate = JSON.parse(caDescription.registrationConfig.templateBody);
                                    const policyName = caTemplate.Resources.policy.Properties.PolicyName;
                                    if (policyName !== undefined) {
                                        const policyResponse = await iot.getPolicy({ policyName: policyName }).promise();
                                        if (policyResponse.policyDocument) {
                                            caListEventData.push({ name: nameTag.Value, arn: ca.certificateArn, id: ca.certificateId, status: ca.status, creation_date: ca.creationDate, policy_name: policyName, policy_status: "Active", policy_document: policyResponse.policyDocument });
                                        }
                                    }
                                    else {
                                        caListEventData.push({ name: nameTag.Value, arn: ca.certificateArn, id: ca.certificateId, status: ca.status, creation_date: ca.creationDate, policy_status: "Inconsistent" });
                                    }
                                }
                                else {
                                    caListEventData.push({ name: nameTag.Value, arn: ca.certificateArn, id: ca.certificateId, status: ca.status, creation_date: ca.creationDate, policy_status: "NoPolicy" });
                                }
                            }
                        }
                    }
                    catch (err) {
                        console.log("error while geting information for CA [" + ca.certificateId + "]", err);
                    }
                }
            }
        }
    }
    catch (err) {
        console.log("error while lisiting CA certificates", err);
    }
    console.log(caListEventData);
    const cloudEvent = new cloudevents_1.CloudEvent({
        type: "io.lamassu.iotcore.config.response",
        id: requestedCloudEvent.id,
        source: "aws/lambda",
        time: new Date().toString(),
        specversion: "1.0",
        data: {
            iot_core_endpoint: endpointInfo.endpointAddress,
            account_id: process.env.AWS_ACCOUNT_ID,
            registered_cas: caListEventData
        }
    });
    try {
        const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL, MessageBody: cloudEvent.toString() }).promise();
        console.log(sqsResponse);
    }
    catch (err) {
        console.log("error while sending SQS messgae", err);
    }
};
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBa0M7QUFDbEMsNkNBQXdDO0FBRXhDLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFDckIsTUFBTSxHQUFHLEdBQUcsSUFBSSxhQUFHLEVBQUUsQ0FBQTtBQUVSLFFBQUEsT0FBTyxHQUFHLEtBQUssRUFBRSxLQUFVLEVBQUUsRUFBRTtJQUMxQyxNQUFNLG1CQUFtQixHQUFHLElBQUksd0JBQVUsQ0FBTSxLQUFLLENBQUMsQ0FBQTtJQUV0RCxNQUFNLFlBQVksR0FBRyxNQUFNLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyxFQUFFLFlBQVksRUFBRSxVQUFVLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO0lBQ3ZGLCtCQUErQjtJQUUvQixNQUFNLGVBQWUsR0FBRyxFQUFFLENBQUE7SUFDMUIsSUFBSTtRQUNGLE1BQU0sZUFBZSxHQUFHLE1BQU0sR0FBRyxDQUFDLGtCQUFrQixDQUFDLEVBQUUsY0FBYyxFQUFFLElBQUksRUFBRSxRQUFRLEVBQUUsRUFBRSxFQUFFLENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTtRQUN0RyxJQUFJLGVBQWUsQ0FBQyxZQUFZLEtBQUssU0FBUyxFQUFFO1lBQzlDLEtBQUssSUFBSSxDQUFDLEdBQUcsQ0FBQyxFQUFFLENBQUMsR0FBRyxlQUFlLENBQUMsWUFBWSxDQUFDLE1BQU0sRUFBRSxDQUFDLEVBQUUsRUFBRTtnQkFDNUQsTUFBTSxFQUFFLEdBQUcsZUFBZSxDQUFDLFlBQVksQ0FBQyxDQUFDLENBQUMsQ0FBQTtnQkFDMUMsSUFBSSxFQUFFLENBQUMsYUFBYSxLQUFLLFNBQVMsSUFBSSxFQUFFLENBQUMsY0FBYyxFQUFFO29CQUN2RCxPQUFPLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxhQUFhLENBQUMsQ0FBQTtvQkFDN0IsSUFBSTt3QkFDRixNQUFNLFlB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RUFBRSxHQUFHLENBQUMsQ0FBQTtLQUN6RDtJQUVELE9BQU8sQ0FBQyxHQUFHLENBQUMsZUFBZSxDQUFDLENBQUE7SUFFNUIsTUFBTSxVQUFVLEdBQUcsSUFBSSx3QkFBVSxDQUFDO1FBQ2hDLElBQUksRUFBRSxvQ0FBb0M7UUFDMUMsRUFBRSxFQUFFLG1CQUFtQixDQUFDLEVBQUU7UUFDMUIsTUFBTSxFQUFFLFlBQVk7UUFDcEIsSUFBSSxFQUFFLElBQUksSUFBSSxFQUFFLENBQUMsUUFBUSxFQUFFO1FBQzNCLFdBQVcsRUFBRSxLQUFLO1FBQ2xCLElBQUksRUFBRTtZQUNKLGlCQUFpQixFQUFFLFlBQVksQ0FBQyxlQUFlO1lBQy9DLFVBQVUsRUFBRSxPQUFPLENBQUMsR0FBRyxDQUFDLGNBQWM7WUFDdEMsY0FBYyxFQUFFLGVBQWU7U0FDaEM7S0FDRixDQUFDLENBQUE7SUFDRixJQUFJO1FBQ0YsTUFBTSxXQUFXLEdBQUcsTUFBTSxHQUFHLENBQUMsV0FBVyxDQUFDLEVBQUUsUUFBUSxFQUFFLE9BQU8sQ0FBQyxHQUFHLENBQUMsc0JBQXVCLEVBQUUsV0FBVyxFQUFFLFVBQVUsQ0FBQyxRQUFRLEVBQUUsRUFBRSxDQUFDLENBQUMsT0FBTyxFQUFFLENBQUE7UUFDMUksT0FBTyxDQUFDLEdBQUcsQ0FBQyxXQUFXLENBQUMsQ0FBQTtLQUN6QjtJQUFDLE9BQU8sR0FBRyxFQUFFO1FBQ1osT0FBTyxDQUFDLEdBQUcsQ0FBQyxpQ0FBaUMsRUFBRSxHQUFHLENBQUMsQ0FBQTtLQUNwRDtBQUNILENBQUMsQ0FBQSIsInNvdXJjZXNDb250ZW50IjpbImltcG9ydCB7IElvdCwgU1FTIH0gZnJvbSBcImF3cy1zZGtcIlxuaW1wb3J0IHsgQ2xvdWRFdmVudCB9IGZyb20gXCJjbG91ZGV2ZW50c1wiXG5cbmNvbnN0IGlvdCA9IG5ldyBJb3QoKVxuY29uc3Qgc3FzID0gbmV3IFNRUygpXG5cbmV4cG9ydCBjb25zdCBoYW5kbGVyID0gYXN5bmMgKGV2ZW50OiBhbnkpID0+IHtcbiAgY29uc3QgcmVxdWVzdGVkQ2xvdWRFdmVudCA9IG5ldyBDbG91ZEV2ZW50PGFueT4oZXZlbnQpXG5cbiAgY29uc3QgZW5kcG9pbnRJbmZvID0gYXdhaXQgaW90LmRlc2NyaWJlRW5kcG9pbnQoeyBlbmRwb2ludFR5cGU6IFwiaW90OkRhdGFcIiB9KS5wcm9taXNlKClcbiAgLy8gZW5kcG9pbnRJbmZvLmVuZHBvaW50QWRkcmVzc1xuXG4gIGNvbnN0IGNhTGlzdEV2ZW50RGF0YSA9IFtdXG4gIHRyeSB7XG4gICAgY29uc3QgbGlzdENBc1Jlc3BvbnNlID0gYXdhaXQgaW90Lmxpc3RDQUNlcnRpZmljYXRlcyh7IGFzY2VuZGluZ09yZGVyOiB0cnVlLCBwYWdlU2l6ZTogMzAgfSkucHJvbWlzZSgpXG4gICAgaWYgKGxpc3RDQXNSZXNwb25zZS5jZXJ0aWZpY2F0ZXMgIT09IHVuZGVmaW5lZCkge1xuICAgICAgZm9yIChsZXQgaSA9IDA7IGkgPCBsaXN0Q0FzUmVzcG9uc2UuY2VydGlmaWNhdGVzLmxlbmd0aDsgaSsrKSB7XG4gICAgICAgIGNvbnN0IGNhID0gbGlzdENBc1Jlc3BvbnNlLmNlcnRpZmljYXRlc1tpXVxuICAgICAgICBpZiAoY2EuY2VydGlmaWNhdGVJZCAhPT0gdW5kZWZpbmVkICYmIGNhLmNlcnRpZmljYXRlQXJuKSB7XG4gICAgICAgICAgY29uc29sZS5sb2coY2EuY2VydGlmaWNhdGVJZClcbiAgICAgICAgICB0cnkge1xuICAgICAgICAgICAgY29uc3QgdGFnc1Jlc3BvbnNlID0gYXdhaXQgaW90Lmxpc3RUYWdzRm9yUmVzb3VyY2UoeyByZXNvdXJjZUFybjogY2EuY2VydGlmaWNhdGVBcm4gfSkucHJvbWlzZSgpXG4gICAgICAgICAgICBjb25zb2xlLmxvZyh0YWdzUmVzcG9uc2UpXG4gICAgICAgICAgICBpZiAodGFnc1Jlc3BvbnNlLnRhZ3MgIT09IHVuZGVmaW5lZCkge1xuICAgICAgICAgICAgICBjb25zdCBuYW1lVGFnID0gdGFnc1Jlc3BvbnNlLnRhZ3MuZmluZCh0YWcgPT4gdGFnLktleSA9PT0gXCJsYW1hc3N1Q0FOYW1lXCIpXG4gICAgICAgICAgICAgIGlmIChuYW1lVGFnICE9PSB1bmRlZmluZWQpIHtcbiAgICAgICAgICAgICAgICBjb25zdCBjYURlc2NyaXB0aW9uID0gYXdhaXQgaW90LmRlc2NyaWJlQ0FDZXJ0aWZpY2F0ZSh7IGNlcnRpZmljYXRlSWQ6IGNhLmNlcnRpZmljYXRlSWQgfSkucHJvbWlzZSgpXG4gICAgICAgICAgICAgICAgY29uc29sZS5sb2coY2FEZXNjcmlwdGlvbi5yZWdpc3RyYXRpb25Db25maWcpXG4gICAgICAgICAgICAgICAgaWYgKGNhRGVzY3JpcHRpb24ucmVnaXN0cmF0aW9uQ29uZmlnICE9PSB1bmRlZmluZWQgJiYgY2FEZXNjcmlwdGlvbi5yZWdpc3RyYXRpb25Db25maWcudGVtcGxhdGVCb2R5KSB7XG4gICAgICAgICAgICAgICAgICBjb25zdCBjYVRlbXBsYXRlID0gSlNPTi5wYXJzZShjYURlc2NyaXB0aW9uLnJlZ2lzdHJhdGlvbkNvbmZpZy50ZW1wbGF0ZUJvZHkpXG4gICAgICAgICAgICAgICAgICBjb25zdCBwb2xpY3lOYW1lID0gY2FUZW1wbGF0ZS5SZXNvdXJjZXMucG9saWN5LlByb3BlcnRpZXMuUG9saWN5TmFtZVxuICAgICAgICAgICAgICAgICAgaWYgKHBvbGljeU5hbWUgIT09IHVuZGVmaW5lZCkge1xuICAgICAgICAgICAgICAgICAgICBjb25zdCBwb2xpY3lSZXNwb25zZSA9IGF3YWl0IGlvdC5nZXRQb2xpY3koeyBwb2xpY3lOYW1lOiBwb2xpY3lOYW1lIH0pLnByb21pc2UoKVxuICAgICAgICAgICAgICAgICAgICBpZiAocG9saWN5UmVzcG9uc2UucG9saWN5RG9jdW1lbnQpIHtcbiAgICAgICAgICAgICAgICAgICAgICBjYUxpc3RFdmVudERhdGEucHVzaCh7IG5hbWU6IG5hbWVUYWcuVmFsdWUsIGFybjogY2EuY2VydGlmaWNhdGVBcm4sIGlkOiBjYS5jZXJ0aWZpY2F0ZUlkLCBzdGF0dXM6IGNhLnN0YXR1cywgY3JlYXRpb25fZGF0ZTogY2EuY3JlYXRpb25EYXRlLCBwb2xpY3lfbmFtZTogcG9saWN5TmFtZSwgcG9saWN5X3N0YXR1czogXCJBY3RpdmVcIiwgcG9saWN5X2RvY3VtZW50OiBwb2xpY3lSZXNwb25zZS5wb2xpY3lEb2N1bWVudCB9KVxuICAgICAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgICB9IGVsc2Uge1xuICAgICAgICAgICAgICAgICAgICBjYUxpc3RFdmVudERhdGEucHVzaCh7IG5hbWU6IG5hbWVUYWcuVmFsdWUsIGFybjogY2EuY2VydGlmaWNhdGVBcm4sIGlkOiBjYS5jZXJ0aWZpY2F0ZUlkLCBzdGF0dXM6IGNhLnN0YXR1cywgY3JlYXRpb25fZGF0ZTogY2EuY3JlYXRpb25EYXRlLCBwb2xpY3lfc3RhdHVzOiBcIkluY29uc2lzdGVudFwiIH0pXG4gICAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgICAgICAgIGNhTGlzdEV2ZW50RGF0YS5wdXNoKHsgbmFtZTogbmFtZVRhZy5WYWx1ZSwgYXJuOiBjYS5jZXJ0aWZpY2F0ZUFybiwgaWQ6IGNhLmNlcnRpZmljYXRlSWQsIHN0YXR1czogY2Euc3RhdHVzLCBjcmVhdGlvbl9kYXRlOiBjYS5jcmVhdGlvbkRhdGUsIHBvbGljeV9zdGF0dXM6IFwiTm9Qb2xpY3lcIiB9KVxuICAgICAgICAgICAgICAgIH1cbiAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgfVxuICAgICAgICAgIH0gY2F0Y2ggKGVycikge1xuICAgICAgICAgICAgY29uc29sZS5sb2coXCJlcnJvciB3aGlsZSBnZXRpbmcgaW5mb3JtYXRpb24gZm9yIENBIFtcIiArIGNhLmNlcnRpZmljYXRlSWQgKyBcIl1cIiwgZXJyKVxuICAgICAgICAgIH1cbiAgICAgICAgfVxuICAgICAgfVxuICAgIH1cbiAgfSBjYXRjaCAoZXJyKSB7XG4gICAgY29uc29sZS5sb2coXCJlcnJvciB3aGlsZSBsaXNpdGluZyBDQSBjZXJ0aWZpY2F0ZXNcIiwgZXJyKVxuICB9XG5cbiAgY29uc29sZS5sb2coY2FMaXN0RXZlbnREYXRhKVxuXG4gIGNvbnN0IGNsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudCh7XG4gICAgdHlwZTogXCJpby5sYW1hc3N1LmlvdGNvcmUuY29uZmlnLnJlc3BvbnNlXCIsXG4gICAgaWQ6IHJlcXVlc3RlZENsb3VkRXZlbnQuaWQsXG4gICAgc291cmNlOiBcImF3cy9sYW1iZGFcIixcbiAgICB0aW1lOiBuZXcgRGF0ZSgpLnRvU3RyaW5nKCksXG4gICAgc3BlY3ZlcnNpb246IFwiMS4wXCIsXG4gICAgZGF0YToge1xuICAgICAgaW90X2NvcmVfZW5kcG9pbnQ6IGVuZHBvaW50SW5mby5lbmRwb2ludEFkZHJlc3MsXG4gICAgICBhY2NvdW50X2lkOiBwcm9jZXNzLmVudi5BV1NfQUNDT1VOVF9JRCxcbiAgICAgIHJlZ2lzdGVyZWRfY2FzOiBjYUxpc3RFdmVudERhdGFcbiAgICB9XG4gIH0pXG4gIHRyeSB7XG4gICAgY29uc3Qgc3FzUmVzcG9uc2UgPSBhd2FpdCBzcXMuc2VuZE1lc3NhZ2UoeyBRdWV1ZVVybDogcHJvY2Vzcy5lbnYuU1FTX1JFU1BPTlNFX1FVRVVFX1VSTCEsIE1lc3NhZ2VCb2R5OiBjbG91ZEV2ZW50LnRvU3RyaW5nKCkgfSkucHJvbWlzZSgpXG4gICAgY29uc29sZS5sb2coc3FzUmVzcG9uc2UpXG4gIH0gY2F0Y2ggKGVycikge1xuICAgIGNvbnNvbGUubG9nKFwiZXJyb3Igd2hpbGUgc2VuZGluZyBTUVMgbWVzc2dhZVwiLCBlcnIpXG4gIH1cbn1cbiJdfQ==