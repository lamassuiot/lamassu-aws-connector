"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handler = void 0;
const aws_sdk_1 = require("aws-sdk");
const cloudevents_1 = require("cloudevents");
const iot = new aws_sdk_1.Iot();
const returnIoTCoreCATemplate = (policyName) => {
    const templateBody = {
        Parameters: {
            "AWS::IoT::Certificate::CommonName": {
                Type: "String"
            },
            "AWS::IoT::Certificate::Country": {
                Type: "String"
            },
            "AWS::IoT::Certificate::Id": {
                Type: "String"
            },
            "AWS::IoT::Certificate::SerialNumber": {
                Type: "String"
            }
        },
        Resources: {
            thing: {
                Type: "AWS::IoT::Thing",
                Properties: {
                    ThingName: {
                        Ref: "AWS::IoT::Certificate::CommonName"
                    },
                    AttributePayload: {}
                }
            },
            certificate: {
                Type: "AWS::IoT::Certificate",
                Properties: {
                    CertificateId: {
                        Ref: "AWS::IoT::Certificate::Id"
                    },
                    Status: "ACTIVE"
                }
            },
            policy: {
                Type: "AWS::IoT::Policy",
                Properties: {
                    PolicyName: policyName
                }
            }
        }
    };
    return templateBody;
};
exports.handler = async (event) => {
    console.log(event);
    const requestedCloudEvent = new cloudevents_1.CloudEvent(event);
    try {
        const listCAsResponse = await iot.listCACertificates({ ascendingOrder: true, pageSize: 30 }).promise();
        console.log(listCAsResponse.certificates);
        if (listCAsResponse.certificates) {
            for (let i = 0; i < listCAsResponse.certificates.length; i++) {
                const ca = listCAsResponse.certificates[i];
                if (ca.certificateArn !== undefined && ca.certificateId !== undefined) {
                    try {
                        const tagsResponse = await iot.listTagsForResource({ resourceArn: ca.certificateArn }).promise();
                        if (tagsResponse.tags) {
                            const nameTag = tagsResponse.tags.find(tag => tag.Key === "lamassuCAName");
                            if (nameTag && nameTag.Value === requestedCloudEvent.data.ca_name) {
                                const policyName = "lamassulambdapolicy_" + requestedCloudEvent.data.ca_name + "_" + Date.now();
                                try {
                                    await iot.createPolicy({
                                        policyDocument: requestedCloudEvent.data.policy,
                                        policyName: policyName,
                                        tags: [{
                                                Key: "serialNumber",
                                                Value: requestedCloudEvent.data.serial_number
                                            }]
                                    }).promise();
                                    try {
                                        await iot.updateCACertificate({
                                            certificateId: ca.certificateId,
                                            newAutoRegistrationStatus: "ENABLE",
                                            newStatus: "ACTIVE",
                                            registrationConfig: {
                                                roleArn: `arn:aws:iam::${process.env.AWS_ACCOUNT_ID}:role/JITPRole`,
                                                templateBody: JSON.stringify(returnIoTCoreCATemplate(policyName))
                                            },
                                            removeAutoRegistration: false
                                        }).promise();
                                    }
                                    catch (err) {
                                        console.log("error while updating IoT Core CA [" + ca.certificateId + "] from IoT Core", err);
                                    }
                                }
                                catch (err) {
                                    console.log("error while creating IoT Core Policy for CA [" + ca.certificateId + "] from IoT Core", err);
                                }
                            }
                        }
                    }
                    catch (err) {
                        console.log("error while listing tags for CA [" + ca.certificateId + "] from IoT Core", err);
                    }
                }
                else {
                    console.log("b");
                    throw new Error("CA certificate ID is null");
                }
            }
        }
    }
    catch (err) {
        console.log("error while listing CAs from IoT Core", err);
    }
};
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBNkI7QUFDN0IsNkNBQXdDO0FBRXhDLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFckIsTUFBTSx1QkFBdUIsR0FBRyxDQUFDLFVBQWtCLEVBQUUsRUFBRTtJQUNyRCxNQUFNLFlBQVksR0FBRztRQUNuQixVQUFVLEVBQUU7WUFDVixtQ0FBbUMsRUFBRTtnQkFDbkMsSUFBSSxFQUFFLFFBQVE7YUFDZjtZQUNELGdDQUFnQyxFQUFFO2dCQUNoQyxJQUFJLEVBQUUsUUFBUTthQUNmO1lBQ0QsMkJBQTJCLEVBQUU7Z0JBQzNCLElBQUksRUFBRSxRQUFRO2FBQ2Y7WUFDRCxxQ0FBcUMsRUFBRTtnQkFDckMsSUFBSSxFQUFFLFFBQVE7YUFDZjtTQUNGO1FBQ0QsU0FBUyxFQUFFO1lBQ1QsS0FBSyxFQUFFO2dCQUNMLElBQUksRUFBRSxpQkFBaUI7Z0JBQ3ZCLFVBQVUsRUFBRTtvQkFDVixTQUFTLEVBQUU7d0JBQ1QsR0FBRyxFQUFFLG1DQUFtQztxQkFDekM7b0JBQ0QsZ0JBQWdCLEVBQUUsRUFBRTtpQkFDckI7YUFDRjtZQUNELFdBQVcsRUFBRTtnQkFDWCxJQUFJLEVBQUUsdUJBQXVCO2dCQUM3QixVQUFVLEVBQUU7b0JBQ1YsYUFBYSxFQUFFO3dCQUNiLEdBQUcsRUFBRSwyQkFBMkI7cUJBQ2pDO29CQUNELE1BQU0sRUFBRSxRQUFRO2lCQUNqQjthQUNGO1lBQ0QsTUFBTSxFQUFFO2dCQUNOLElBQUksRUFBRSxrQkFBa0I7Z0JBQ3hCLFVBQVUsRUFBRTtvQkFDVixVQUFVLEVBQUUsVUFBVTtpQkFDdkI7YUFDRjtTQUNGO0tBQ0YsQ0FBQTtJQUNELE9BQU8sWUFBWSxDQUFBO0FBQ3JCLENBQUMsQ0FBQTtBQUVZLFFBQUEsT0FBTyxHQUFHLEtBQUssRUFBRSxLQUFVLEVBQUUsRUFBRTtJQUMxQyxP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TyxDQUFDLEdBQUcsQ0FBQyxtQ0FBbUMsR0FBRyxFQUFFLENBQUMsYUFBYSxHQUFHLGlCQUFpQixFQUFFLEdBQUcsQ0FBQyxDQUFBO3FCQUM3RjtpQkFDRjtxQkFBTTtvQkFDTCxPQUFPLENBQUMsR0FBRyxDQUFDLEdBQUcsQ0FBQyxDQUFBO29CQUVoQixNQUFNLElBQUksS0FBSyxDQUFDLDJCQUEyQixDQUFDLENBQUE7aUJBQzdDO2FBQ0Y7U0FDRjtLQUNGO0lBQUMsT0FBTyxHQUFHLEVBQUU7UUFDWixPQUFPLENBQUMsR0FBRyxDQUFDLHVDQUF1QyxFQUFFLEdBQUcsQ0FBQyxDQUFBO0tBQzFEO0FBQ0gsQ0FBQyxDQUFBIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgSW90IH0gZnJvbSBcImF3cy1zZGtcIlxuaW1wb3J0IHsgQ2xvdWRFdmVudCB9IGZyb20gXCJjbG91ZGV2ZW50c1wiXG5cbmNvbnN0IGlvdCA9IG5ldyBJb3QoKVxuXG5jb25zdCByZXR1cm5Jb1RDb3JlQ0FUZW1wbGF0ZSA9IChwb2xpY3lOYW1lOiBzdHJpbmcpID0+IHtcbiAgY29uc3QgdGVtcGxhdGVCb2R5ID0ge1xuICAgIFBhcmFtZXRlcnM6IHtcbiAgICAgIFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlOjpDb21tb25OYW1lXCI6IHtcbiAgICAgICAgVHlwZTogXCJTdHJpbmdcIlxuICAgICAgfSxcbiAgICAgIFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlOjpDb3VudHJ5XCI6IHtcbiAgICAgICAgVHlwZTogXCJTdHJpbmdcIlxuICAgICAgfSxcbiAgICAgIFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlOjpJZFwiOiB7XG4gICAgICAgIFR5cGU6IFwiU3RyaW5nXCJcbiAgICAgIH0sXG4gICAgICBcIkFXUzo6SW9UOjpDZXJ0aWZpY2F0ZTo6U2VyaWFsTnVtYmVyXCI6IHtcbiAgICAgICAgVHlwZTogXCJTdHJpbmdcIlxuICAgICAgfVxuICAgIH0sXG4gICAgUmVzb3VyY2VzOiB7XG4gICAgICB0aGluZzoge1xuICAgICAgICBUeXBlOiBcIkFXUzo6SW9UOjpUaGluZ1wiLFxuICAgICAgICBQcm9wZXJ0aWVzOiB7XG4gICAgICAgICAgVGhpbmdOYW1lOiB7XG4gICAgICAgICAgICBSZWY6IFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlOjpDb21tb25OYW1lXCJcbiAgICAgICAgICB9LFxuICAgICAgICAgIEF0dHJpYnV0ZVBheWxvYWQ6IHt9XG4gICAgICAgIH1cbiAgICAgIH0sXG4gICAgICBjZXJ0aWZpY2F0ZToge1xuICAgICAgICBUeXBlOiBcIkFXUzo6SW9UOjpDZXJ0aWZpY2F0ZVwiLFxuICAgICAgICBQcm9wZXJ0aWVzOiB7XG4gICAgICAgICAgQ2VydGlmaWNhdGVJZDoge1xuICAgICAgICAgICAgUmVmOiBcIkFXUzo6SW9UOjpDZXJ0aWZpY2F0ZTo6SWRcIlxuICAgICAgICAgIH0sXG4gICAgICAgICAgU3RhdHVzOiBcIkFDVElWRVwiXG4gICAgICAgIH1cbiAgICAgIH0sXG4gICAgICBwb2xpY3k6IHtcbiAgICAgICAgVHlwZTogXCJBV1M6OklvVDo6UG9saWN5XCIsXG4gICAgICAgIFByb3BlcnRpZXM6IHtcbiAgICAgICAgICBQb2xpY3lOYW1lOiBwb2xpY3lOYW1lXG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9XG4gIH1cbiAgcmV0dXJuIHRlbXBsYXRlQm9keVxufVxuXG5leHBvcnQgY29uc3QgaGFuZGxlciA9IGFzeW5jIChldmVudDogYW55KSA9PiB7XG4gIGNvbnNvbGUubG9nKGV2ZW50KVxuXG4gIGNvbnN0IHJlcXVlc3RlZENsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudDxhbnk+KGV2ZW50KVxuICB0cnkge1xuICAgIGNvbnN0IGxpc3RDQXNSZXNwb25zZSA9IGF3YWl0IGlvdC5saXN0Q0FDZXJ0aWZpY2F0ZXMoeyBhc2NlbmRpbmdPcmRlcjogdHJ1ZSwgcGFnZVNpemU6IDMwIH0pLnByb21pc2UoKVxuICAgIGNvbnNvbGUubG9nKGxpc3RDQXNSZXNwb25zZS5jZXJ0aWZpY2F0ZXMpXG4gICAgaWYgKGxpc3RDQXNSZXNwb25zZS5jZXJ0aWZpY2F0ZXMpIHtcbiAgICAgIGZvciAobGV0IGkgPSAwOyBpIDwgbGlzdENBc1Jlc3BvbnNlLmNlcnRpZmljYXRlcy5sZW5ndGg7IGkrKykge1xuICAgICAgICBjb25zdCBjYSA9IGxpc3RDQXNSZXNwb25zZS5jZXJ0aWZpY2F0ZXNbaV1cbiAgICAgICAgaWYgKGNhLmNlcnRpZmljYXRlQXJuICE9PSB1bmRlZmluZWQgJiYgY2EuY2VydGlmaWNhdGVJZCAhPT0gdW5kZWZpbmVkKSB7XG4gICAgICAgICAgdHJ5IHtcbiAgICAgICAgICAgIGNvbnN0IHRhZ3NSZXNwb25zZSA9IGF3YWl0IGlvdC5saXN0VGFnc0ZvclJlc291cmNlKHsgcmVzb3VyY2VBcm46IGNhLmNlcnRpZmljYXRlQXJuIH0pLnByb21pc2UoKVxuICAgICAgICAgICAgaWYgKHRhZ3NSZXNwb25zZS50YWdzKSB7XG4gICAgICAgICAgICAgIGNvbnN0IG5hbWVUYWcgPSB0YWdzUmVzcG9uc2UudGFncy5maW5kKHRhZyA9PiB0YWcuS2V5ID09PSBcImxhbWFzc3VDQU5hbWVcIilcbiAgICAgICAgICAgICAgaWYgKG5hbWVUYWcgJiYgbmFtZVRhZy5WYWx1ZSA9PT0gcmVxdWVzdGVkQ2xvdWRFdmVudC5kYXRhLmNhX25hbWUpIHtcbiAgICAgICAgICAgICAgICBjb25zdCBwb2xpY3lOYW1lID0gXCJsYW1hc3N1bGFtYmRhcG9saWN5X1wiICsgcmVxdWVzdGVkQ2xvdWRFdmVudC5kYXRhLmNhX25hbWUgKyBcIl9cIiArIERhdGUubm93KClcbiAgICAgICAgICAgICAgICB0cnkge1xuICAgICAgICAgICAgICAgICAgYXdhaXQgaW90LmNyZWF0ZVBvbGljeSh7XG4gICAgICAgICAgICAgICAgICAgIHBvbGljeURvY3VtZW50OiByZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEucG9saWN5LFxuICAgICAgICAgICAgICAgICAgICBwb2xpY3lOYW1lOiBwb2xpY3lOYW1lLFxuICAgICAgICAgICAgICAgICAgICB0YWdzOiBbe1xuICAgICAgICAgICAgICAgICAgICAgIEtleTogXCJzZXJpYWxOdW1iZXJcIixcbiAgICAgICAgICAgICAgICAgICAgICBWYWx1ZTogcmVxdWVzdGVkQ2xvdWRFdmVudC5kYXRhLnNlcmlhbF9udW1iZXJcbiAgICAgICAgICAgICAgICAgICAgfV1cbiAgICAgICAgICAgICAgICAgIH0pLnByb21pc2UoKVxuXG4gICAgICAgICAgICAgICAgICB0cnkge1xuICAgICAgICAgICAgICAgICAgICBhd2FpdCBpb3QudXBkYXRlQ0FDZXJ0aWZpY2F0ZSh7XG4gICAgICAgICAgICAgICAgICAgICAgY2VydGlmaWNhdGVJZDogY2EuY2VydGlmaWNhdGVJZCxcbiAgICAgICAgICAgICAgICAgICAgICBuZXdBdXRvUmVnaXN0cmF0aW9uU3RhdHVzOiBcIkVOQUJMRVwiLFxuICAgICAgICAgICAgICAgICAgICAgIG5ld1N0YXR1czogXCJBQ1RJVkVcIixcbiAgICAgICAgICAgICAgICAgICAgICByZWdpc3RyYXRpb25Db25maWc6IHtcbiAgICAgICAgICAgICAgICAgICAgICAgIHJvbGVBcm46IGBhcm46YXdzOmlhbTo6JHtwcm9jZXNzLmVudi5BV1NfQUNDT1VOVF9JRH06cm9sZS9KSVRQUm9sZWAsXG4gICAgICAgICAgICAgICAgICAgICAgICB0ZW1wbGF0ZUJvZHk6IEpTT04uc3RyaW5naWZ5KHJldHVybklvVENvcmVDQVRlbXBsYXRlKHBvbGljeU5hbWUpKVxuICAgICAgICAgICAgICAgICAgICAgIH0sXG4gICAgICAgICAgICAgICAgICAgICAgcmVtb3ZlQXV0b1JlZ2lzdHJhdGlvbjogZmFsc2VcbiAgICAgICAgICAgICAgICAgICAgfSkucHJvbWlzZSgpXG4gICAgICAgICAgICAgICAgICB9IGNhdGNoIChlcnIpIHtcbiAgICAgICAgICAgICAgICAgICAgY29uc29sZS5sb2coXCJlcnJvciB3aGlsZSB1cGRhdGluZyBJb1QgQ29yZSBDQSBbXCIgKyBjYS5jZXJ0aWZpY2F0ZUlkICsgXCJdIGZyb20gSW9UIENvcmVcIiwgZXJyKVxuICAgICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICAgIH0gY2F0Y2ggKGVycikge1xuICAgICAgICAgICAgICAgICAgY29uc29sZS5sb2coXCJlcnJvciB3aGlsZSBjcmVhdGluZyBJb1QgQ29yZSBQb2xpY3kgZm9yIENBIFtcIiArIGNhLmNlcnRpZmljYXRlSWQgKyBcIl0gZnJvbSBJb1QgQ29yZVwiLCBlcnIpXG4gICAgICAgICAgICAgICAgfVxuICAgICAgICAgICAgICB9XG4gICAgICAgICAgICB9XG4gICAgICAgICAgfSBjYXRjaCAoZXJyKSB7XG4gICAgICAgICAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIGxpc3RpbmcgdGFncyBmb3IgQ0EgW1wiICsgY2EuY2VydGlmaWNhdGVJZCArIFwiXSBmcm9tIElvVCBDb3JlXCIsIGVycilcbiAgICAgICAgICB9XG4gICAgICAgIH0gZWxzZSB7XG4gICAgICAgICAgY29uc29sZS5sb2coXCJiXCIpXG5cbiAgICAgICAgICB0aHJvdyBuZXcgRXJyb3IoXCJDQSBjZXJ0aWZpY2F0ZSBJRCBpcyBudWxsXCIpXG4gICAgICAgIH1cbiAgICAgIH1cbiAgICB9XG4gIH0gY2F0Y2ggKGVycikge1xuICAgIGNvbnNvbGUubG9nKFwiZXJyb3Igd2hpbGUgbGlzdGluZyBDQXMgZnJvbSBJb1QgQ29yZVwiLCBlcnIpXG4gIH1cbn1cbiJdfQ==