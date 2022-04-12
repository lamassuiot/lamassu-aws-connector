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
        console.log("Cloud Event", cloudEvent);
        const sqsResponse = await sqs.sendMessage({ QueueUrl: process.env.SQS_RESPONSE_QUEUE_URL, MessageBody: cloudEvent.toString() }).promise();
        console.log(sqsResponse);
    }
    catch (err) {
        console.log("error while sending SQS messgae", err);
    }
};
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBa0M7QUFDbEMsNkNBQXdDO0FBRXhDLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFDckIsTUFBTSxHQUFHLEdBQUcsSUFBSSxhQUFHLEVBQUUsQ0FBQTtBQUVSLFFBQUEsT0FBTyxHQUFHLEtBQUssRUFBRSxLQUFVLEVBQUUsRUFBRTtJQUMxQyxNQUFNLG1CQUFtQixHQUFHLElBQUksd0JBQVUsQ0FBTSxLQUFLLENBQUMsQ0FBQTtJQUV0RCxNQUFNLFlBQVksR0FBRyxNQUFNLEdBQUcsQ0FBQyxnQkFBZ0IsQ0FBQyxFQUFFLFlBQVksRUFBRSxVQUFVLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO0lBQ3ZGLCtCQUErQjtJQUUvQixNQUFNLGVBQWUsR0FBRyxFQUFFLENBQUE7SUFDMUIsSUFBSTtRQUNGLE1BQU0sZUFBZSxHQUFHLE1BQU0sR0FBRyxDQUFDLGtCQUFrQixDQUFDLEVBQUUsY0FBYyxFQUFFLElBQUksRUFBRSxRQUFRLEVBQUUsRUFBRSxFQUFFLENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTtRQUN0RyxJQUFJLGVBQWUsQ0FBQyxZQUFZLEtBQUssU0FBUyxFQUFFO1lBQzlDLEtBQUssSUFBSSxDQUFDLEdBQUcsQ0FBQyxFQUFFLENBQUMsR0FBRyxlQUFlLENBQUMsWUFBWSxDQUFDLE1BQU0sRUFBRSxDQUFDLEVBQUUsRUFBRTtnQkFDNUQsTUFBTSxFQUFFLEdBQUcsZUFBZSxDQUFDLFlBQVksQ0FBQyxDQUFDLENBQUMsQ0FBQTtnQkFDMUMsSUFBSSxFQUFFLENBQUMsYUFBYSxLQUFLLFNBQVMsSUFBSSxFQUFFLENBQUMsY0FBYyxFQUFFO29CQUN2RCxPQUFPLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxhQUFhLENBQUMsQ0FBQTtvQkFDN0IsSUFBSTt3QkFDRixNQUFNLFlBQVksR0FBRyxNQUFNLEdBQUcsQ0FBQyxtQkFBbUIsQ0FBQyxFQUFFLFdBQVcsRUFBRSxFQUFFLENBQUMsY0FBYyxFQUFFLENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTt3QkFDaEcsT0FBTyxDQUFDLEdBQUcsQ0FBQyxZQUFZLENBQUMsQ0FBQTt3QkFDekIsSUFBSSxZQUFZLENBQUMsSUFBSSxLQUFLLFNBQVMsRUFBRTs0QkFDbkMsTUFBTSxPQUFPLEdBQUcsWUFBWSxDQUFDLElBQUksQ0FBQyxJQUFJLENBQUMsR0FBRyxDQUFDLEVBQUUsQ0FBQyxHQUFHLENBQUMsR0FBRyxLQUFLLGVBQWUsQ0FBQyxDQUFBOzRCQUMxRSxJQUFJLE9BQU8sS0FBSyxTQUFTLEVBQUU7Z0NBQ3pCLE1BQU0sYUFBYSxHQUFHLE1BQU0sR0FBRyxDQUFDLHFCQUFxQixDQUFDLEVBQUUsYUFBYSxFQUFFLEVBQUUsQ0FBQyxhQUFhLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO2dDQUNwRyxPQUFPLENBQUMsR0FBRyxDQUFDLGFBQWEsQ0FBQyxrQkFBa0IsQ0FBQyxDQUFBO2dDQUM3QyxJQUFJLGFBQWEsQ0FBQyxrQkFBa0IsS0FBSyxTQUFTLElBQUksYUFBYSxDQUFDLGtCQUFrQixDQUFDLFlBQVksRUFBRTtvQ0FDbkcsTUFBTSxVQUFVLEdBQUcsSUFBSSxDQUFDLEtBQUssQ0FBQyxhQUFhLENBQUMsa0JBQWtCLENBQUMsWUFBWSxDQUFDLENBQUE7b0NBQzVFLE1BQU0sVUFBVSxHQUFHLFVBQVUsQ0FBQyxTQUFTLENBQUMsTUFBTSxDQUFDLFVBQVUsQ0FBQyxVQUFVLENBQUE7b0NBQ3BFLElBQUksVUFBVSxLQUFLLFNBQVMsRUFBRTt3Q0FDNUIsTUFBTSxjQUFjLEdBQUcsTUFBTSxHQUFHLENBQUMsU0FBUyxDQUFDLEVBQUUsVUFBVSxFQUFFLFVBQVUsRUFBRSxDQUFDLENBQUMsT0FBTyxFQUFFLENBQUE7d0NBQ2hGLElBQUksY0FBYyxDQUFDLGNBQWMsRUFBRTs0Q0FDakMsZUFBZSxDQUFDLElBQUksQ0FBQyxFQUFFLElBQUksRUFBRSxPQUFPLENBQUMsS0FBSyxFQUFFLEdBQUcsRUFBRSxFQUFFLENBQUMsY0FBYyxFQUFFLEVBQUUsRUFBRSxFQUFFLENBQUMsYUFBYSxFQUFFLE1BQU0sRUFBRSxFQUFFLENBQUMsTUFBTSxFQUFFLGFBQWEsRUFBRSxFQUFFLENBQUMsWUFBWSxFQUFFLFdBQVcsRUFBRSxVQUFVLEVBQUUsYUFBYSxFQUFFLFFBQVEsRUFBRSxlQUFlLEVBQUUsY0FBYyxDQUFDLGNBQWMsRUFBRSxDQUFDLENBQUE7eUNBQ2pQO3FDQUNGO3lDQUFNO3dDQUNMLGVBQWUsQ0FBQyxJQUFJLENBQUMsRUFBRSxJQUFJLEVBQUUsT0FBTyxDQUFDLEtBQUssRUFBRSxHQUFHLEVBQUUsRUFBRSxDQUFDLGNBQWMsRUFBRSxFQUFFLEVBQUUsRUFBRSxDQUFDLGFBQWEsRUFBRSxNQUFNLEVBQUUsRUFBRSxDQUFDLE1BQU0sRUFBRSxhQUFhLEVBQUUsRUFBRSxDQUFDLFlBQVksRUFBRSxhQUFhLEVBQUUsY0FBYyxFQUFFLENBQUMsQ0FBQTtxQ0FDOUs7aUNBQ0Y7cUNBQU07b0NBQ0wsZUFBZSxDQUFDLElBQUksQ0FBQyxFQUFFLElBQUksRUFBRSxPQUFPLENBQUMsS0FBSyxFQUFFLEdBQUcsRUFBRSxFQUFFLENBQUMsY0FBYyxFQUFFLEVBQUUsRUFBRSxFQUFFLENBQUMsYUFBYSxFQUFFLE1BQU0sRUFBRSxFQUFFLENBQUMsTUFBTSxFQUFFLGFBQWEsRUFBRSxFQUFFLENBQUMsWUFBWSxFQUFFLGFBQWEsRUFBRSxVQUFVLEVBQUUsQ0FBQyxDQUFBO2lDQUMxSzs2QkFDRjt5QkFDRjtxQkFDRjtvQkFBQyxPQUFPLEdBQUcsRUFBRTt3QkFDWixPQUFPLENBQUMsR0FBRyxDQUFDLHlDQUF5QyxHQUFHLEVBQUUsQ0FBQyxhQUFhLEdBQUcsR0FBRyxFQUFFLEdBQUcsQ0FBQyxDQUFBO3FCQUNyRjtpQkFDRjthQUNGO1NBQ0Y7S0FDRjtJQUFDLE9BQU8sR0FBRyxFQUFFO1FBQ1osT0FBTyxDQUFDLEdBQUcsQ0FBQyxzQ0FBc0MsRUFBRSxHQUFHLENBQUMsQ0FBQTtLQUN6RDtJQUVELE1BQU0sVUFBVSxHQUFHLElBQUksd0JBQVUsQ0FBQztRQUNoQyxJQUFJLEVBQUUsb0NBQW9DO1FBQzFDLEVBQUUsRUFBRSxtQkFBbUIsQ0FBQyxFQUFFO1FBQzFCLE1BQU0sRUFBRSxZQUFZO1FBQ3BCLElBQUksRUFBRSxJQUFJLElBQUksRUFBRSxDQUFDLFFBQVEsRUFBRTtRQUMzQixXQUFXLEVBQUUsS0FBSztRQUNsQixJQUFJLEVBQUU7WUFDSixpQkFBaUIsRUFBRSxZQUFZLENBQUMsZUFBZTtZQUMvQyxVQUFVLEVBQUUsT0FBTyxDQUFDLEdBQUcsQ0FBQyxjQUFjO1lBQ3RDLGNBQWMsRUFBRSxlQUFlO1NBQ2hDO0tBQ0YsQ0FBQyxDQUFBO0lBQ0YsSUFBSTtRQUNGLE9BQU8sQ0FBQyxHQUFHLENBQUMsYUFBYSxFQUFFLFVBQVUsQ0FBQyxDQUFBO1FBQ3RDLE1BQU0sV0FBVyxHQUFHLE1BQU0sR0FBRyxDQUFDLFdBQVcsQ0FBQyxFQUFFLFFBQVEsRUFBRSxPQUFPLENBQUMsR0FBRyxDQUFDLHNCQUF1QixFQUFFLFdBQVcsRUFBRSxVQUFVLENBQUMsUUFBUSxFQUFFLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO1FBQzFJLE9BQU8sQ0FBQyxHQUFHLENBQUMsV0FBVyxDQUFDLENBQUE7S0FDekI7SUFBQyxPQUFPLEdBQUcsRUFBRTtRQUNaLE9BQU8sQ0FBQyxHQUFHLENBQUMsaUNBQWlDLEVBQUUsR0FBRyxDQUFDLENBQUE7S0FDcEQ7QUFDSCxDQUFDLENBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgeyBJb3QsIFNRUyB9IGZyb20gXCJhd3Mtc2RrXCJcbmltcG9ydCB7IENsb3VkRXZlbnQgfSBmcm9tIFwiY2xvdWRldmVudHNcIlxuXG5jb25zdCBpb3QgPSBuZXcgSW90KClcbmNvbnN0IHNxcyA9IG5ldyBTUVMoKVxuXG5leHBvcnQgY29uc3QgaGFuZGxlciA9IGFzeW5jIChldmVudDogYW55KSA9PiB7XG4gIGNvbnN0IHJlcXVlc3RlZENsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudDxhbnk+KGV2ZW50KVxuXG4gIGNvbnN0IGVuZHBvaW50SW5mbyA9IGF3YWl0IGlvdC5kZXNjcmliZUVuZHBvaW50KHsgZW5kcG9pbnRUeXBlOiBcImlvdDpEYXRhXCIgfSkucHJvbWlzZSgpXG4gIC8vIGVuZHBvaW50SW5mby5lbmRwb2ludEFkZHJlc3NcblxuICBjb25zdCBjYUxpc3RFdmVudERhdGEgPSBbXVxuICB0cnkge1xuICAgIGNvbnN0IGxpc3RDQXNSZXNwb25zZSA9IGF3YWl0IGlvdC5saXN0Q0FDZXJ0aWZpY2F0ZXMoeyBhc2NlbmRpbmdPcmRlcjogdHJ1ZSwgcGFnZVNpemU6IDMwIH0pLnByb21pc2UoKVxuICAgIGlmIChsaXN0Q0FzUmVzcG9uc2UuY2VydGlmaWNhdGVzICE9PSB1bmRlZmluZWQpIHtcbiAgICAgIGZvciAobGV0IGkgPSAwOyBpIDwgbGlzdENBc1Jlc3BvbnNlLmNlcnRpZmljYXRlcy5sZW5ndGg7IGkrKykge1xuICAgICAgICBjb25zdCBjYSA9IGxpc3RDQXNSZXNwb25zZS5jZXJ0aWZpY2F0ZXNbaV1cbiAgICAgICAgaWYgKGNhLmNlcnRpZmljYXRlSWQgIT09IHVuZGVmaW5lZCAmJiBjYS5jZXJ0aWZpY2F0ZUFybikge1xuICAgICAgICAgIGNvbnNvbGUubG9nKGNhLmNlcnRpZmljYXRlSWQpXG4gICAgICAgICAgdHJ5IHtcbiAgICAgICAgICAgIGNvbnN0IHRhZ3NSZXNwb25zZSA9IGF3YWl0IGlvdC5saXN0VGFnc0ZvclJlc291cmNlKHsgcmVzb3VyY2VBcm46IGNhLmNlcnRpZmljYXRlQXJuIH0pLnByb21pc2UoKVxuICAgICAgICAgICAgY29uc29sZS5sb2codGFnc1Jlc3BvbnNlKVxuICAgICAgICAgICAgaWYgKHRhZ3NSZXNwb25zZS50YWdzICE9PSB1bmRlZmluZWQpIHtcbiAgICAgICAgICAgICAgY29uc3QgbmFtZVRhZyA9IHRhZ3NSZXNwb25zZS50YWdzLmZpbmQodGFnID0+IHRhZy5LZXkgPT09IFwibGFtYXNzdUNBTmFtZVwiKVxuICAgICAgICAgICAgICBpZiAobmFtZVRhZyAhPT0gdW5kZWZpbmVkKSB7XG4gICAgICAgICAgICAgICAgY29uc3QgY2FEZXNjcmlwdGlvbiA9IGF3YWl0IGlvdC5kZXNjcmliZUNBQ2VydGlmaWNhdGUoeyBjZXJ0aWZpY2F0ZUlkOiBjYS5jZXJ0aWZpY2F0ZUlkIH0pLnByb21pc2UoKVxuICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nKGNhRGVzY3JpcHRpb24ucmVnaXN0cmF0aW9uQ29uZmlnKVxuICAgICAgICAgICAgICAgIGlmIChjYURlc2NyaXB0aW9uLnJlZ2lzdHJhdGlvbkNvbmZpZyAhPT0gdW5kZWZpbmVkICYmIGNhRGVzY3JpcHRpb24ucmVnaXN0cmF0aW9uQ29uZmlnLnRlbXBsYXRlQm9keSkge1xuICAgICAgICAgICAgICAgICAgY29uc3QgY2FUZW1wbGF0ZSA9IEpTT04ucGFyc2UoY2FEZXNjcmlwdGlvbi5yZWdpc3RyYXRpb25Db25maWcudGVtcGxhdGVCb2R5KVxuICAgICAgICAgICAgICAgICAgY29uc3QgcG9saWN5TmFtZSA9IGNhVGVtcGxhdGUuUmVzb3VyY2VzLnBvbGljeS5Qcm9wZXJ0aWVzLlBvbGljeU5hbWVcbiAgICAgICAgICAgICAgICAgIGlmIChwb2xpY3lOYW1lICE9PSB1bmRlZmluZWQpIHtcbiAgICAgICAgICAgICAgICAgICAgY29uc3QgcG9saWN5UmVzcG9uc2UgPSBhd2FpdCBpb3QuZ2V0UG9saWN5KHsgcG9saWN5TmFtZTogcG9saWN5TmFtZSB9KS5wcm9taXNlKClcbiAgICAgICAgICAgICAgICAgICAgaWYgKHBvbGljeVJlc3BvbnNlLnBvbGljeURvY3VtZW50KSB7XG4gICAgICAgICAgICAgICAgICAgICAgY2FMaXN0RXZlbnREYXRhLnB1c2goeyBuYW1lOiBuYW1lVGFnLlZhbHVlLCBhcm46IGNhLmNlcnRpZmljYXRlQXJuLCBpZDogY2EuY2VydGlmaWNhdGVJZCwgc3RhdHVzOiBjYS5zdGF0dXMsIGNyZWF0aW9uX2RhdGU6IGNhLmNyZWF0aW9uRGF0ZSwgcG9saWN5X25hbWU6IHBvbGljeU5hbWUsIHBvbGljeV9zdGF0dXM6IFwiQWN0aXZlXCIsIHBvbGljeV9kb2N1bWVudDogcG9saWN5UmVzcG9uc2UucG9saWN5RG9jdW1lbnQgfSlcbiAgICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICAgICAgICAgICAgY2FMaXN0RXZlbnREYXRhLnB1c2goeyBuYW1lOiBuYW1lVGFnLlZhbHVlLCBhcm46IGNhLmNlcnRpZmljYXRlQXJuLCBpZDogY2EuY2VydGlmaWNhdGVJZCwgc3RhdHVzOiBjYS5zdGF0dXMsIGNyZWF0aW9uX2RhdGU6IGNhLmNyZWF0aW9uRGF0ZSwgcG9saWN5X3N0YXR1czogXCJJbmNvbnNpc3RlbnRcIiB9KVxuICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgICAgICAgICBjYUxpc3RFdmVudERhdGEucHVzaCh7IG5hbWU6IG5hbWVUYWcuVmFsdWUsIGFybjogY2EuY2VydGlmaWNhdGVBcm4sIGlkOiBjYS5jZXJ0aWZpY2F0ZUlkLCBzdGF0dXM6IGNhLnN0YXR1cywgY3JlYXRpb25fZGF0ZTogY2EuY3JlYXRpb25EYXRlLCBwb2xpY3lfc3RhdHVzOiBcIk5vUG9saWN5XCIgfSlcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH1cbiAgICAgICAgICB9IGNhdGNoIChlcnIpIHtcbiAgICAgICAgICAgIGNvbnNvbGUubG9nKFwiZXJyb3Igd2hpbGUgZ2V0aW5nIGluZm9ybWF0aW9uIGZvciBDQSBbXCIgKyBjYS5jZXJ0aWZpY2F0ZUlkICsgXCJdXCIsIGVycilcbiAgICAgICAgICB9XG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9XG4gIH0gY2F0Y2ggKGVycikge1xuICAgIGNvbnNvbGUubG9nKFwiZXJyb3Igd2hpbGUgbGlzaXRpbmcgQ0EgY2VydGlmaWNhdGVzXCIsIGVycilcbiAgfVxuXG4gIGNvbnN0IGNsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudCh7XG4gICAgdHlwZTogXCJpby5sYW1hc3N1LmlvdGNvcmUuY29uZmlnLnJlc3BvbnNlXCIsXG4gICAgaWQ6IHJlcXVlc3RlZENsb3VkRXZlbnQuaWQsXG4gICAgc291cmNlOiBcImF3cy9sYW1iZGFcIixcbiAgICB0aW1lOiBuZXcgRGF0ZSgpLnRvU3RyaW5nKCksXG4gICAgc3BlY3ZlcnNpb246IFwiMS4wXCIsXG4gICAgZGF0YToge1xuICAgICAgaW90X2NvcmVfZW5kcG9pbnQ6IGVuZHBvaW50SW5mby5lbmRwb2ludEFkZHJlc3MsXG4gICAgICBhY2NvdW50X2lkOiBwcm9jZXNzLmVudi5BV1NfQUNDT1VOVF9JRCxcbiAgICAgIHJlZ2lzdGVyZWRfY2FzOiBjYUxpc3RFdmVudERhdGFcbiAgICB9XG4gIH0pXG4gIHRyeSB7XG4gICAgY29uc29sZS5sb2coXCJDbG91ZCBFdmVudFwiLCBjbG91ZEV2ZW50KVxuICAgIGNvbnN0IHNxc1Jlc3BvbnNlID0gYXdhaXQgc3FzLnNlbmRNZXNzYWdlKHsgUXVldWVVcmw6IHByb2Nlc3MuZW52LlNRU19SRVNQT05TRV9RVUVVRV9VUkwhLCBNZXNzYWdlQm9keTogY2xvdWRFdmVudC50b1N0cmluZygpIH0pLnByb21pc2UoKVxuICAgIGNvbnNvbGUubG9nKHNxc1Jlc3BvbnNlKVxuICB9IGNhdGNoIChlcnIpIHtcbiAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIHNlbmRpbmcgU1FTIG1lc3NnYWVcIiwgZXJyKVxuICB9XG59XG4iXX0=