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
                    ThingGroups: ["LAMASSU"],
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
                                console.log("creating policy: " + policyName);
                                try {
                                    await iot.createPolicy({
                                        policyDocument: JSON.stringify(JSON.parse(requestedCloudEvent.data.policy)),
                                        policyName: policyName,
                                        tags: [{
                                                Key: "serialNumber",
                                                Value: requestedCloudEvent.data.serial_number
                                            }]
                                    }).promise();
                                    console.log("updating CA certificate with ID" + ca.certificateId);
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
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBNkI7QUFDN0IsNkNBQXdDO0FBRXhDLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFckIsTUFBTSx1QkFBdUIsR0FBRyxDQUFDLFVBQWtCLEVBQUUsRUFBRTtJQUNyRCxNQUFNLFlBQVksR0FBRztRQUNuQixVQUFVLEVBQUU7WUFDVixtQ0FBbUMsRUFBRTtnQkFDbkMsSUFBSSxFQUFFLFFBQVE7YUFDZjtZQUNELGdDQUFnQyxFQUFFO2dCQUNoQyxJQUFJLEVBQUUsUUFBUTthQUNmO1lBQ0QsMkJBQTJCLEVBQUU7Z0JBQzNCLElBQUksRUFBRSxRQUFRO2FBQ2Y7WUFDRCxxQ0FBcUMsRUFBRTtnQkFDckMsSUFBSSxFQUFFLFFBQVE7YUFDZjtTQUNGO1FBQ0QsU0FBUyxFQUFFO1lBQ1QsS0FBSyxFQUFFO2dCQUNMLElBQUksRUFBRSxpQkFBaUI7Z0JBQ3ZCLFVBQVUsRUFBRTtvQkFDVixTQUFTLEVBQUU7d0JBQ1QsR0FBRyxFQUFFLG1DQUFtQztxQkFDekM7b0JBQ0QsV0FBVyxFQUFFLENBQUMsU0FBUyxDQUFDO29CQUN4QixnQkFBZ0IsRUFBRSxFQUFFO2lCQUNyQjthQUNGO1lBQ0QsV0FBVyxFQUFFO2dCQUNYLElBQUksRUFBRSx1QkFBdUI7Z0JBQzdCLFVBQVUsRUFBRTtvQkFDVixhQUFhLEVBQUU7d0JBQ2IsR0FBRyxFQUFFLDJCQUEyQjtxQkFDakM7b0JBQ0QsTUFBTSxFQUFFLFFBQVE7aUJBQ2pCO2FBQ0Y7WUFDRCxNQUFNLEVBQUU7Z0JBQ04sSUFBSSxFQUFFLGtCQUFrQjtnQkFDeEIsVUFBVSxFQUFFO29CQUNWLFVBQVUsRUFBRSxVQUFVO2lCQUN2QjthQUNGO1NBQ0Y7S0FDRixDQUFBO0lBQ0QsT0FBTyxZQUFZLENBQUE7QUFDckIsQ0FBQyxDQUFBO0FBRVksUUFBQSxPQUFPLEdBQUcs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QkFBaUIsRUFBRSxHQUFHLENBQUMsQ0FBQTtxQ0FDOUY7aUNBQ0Y7Z0NBQUMsT0FBTyxHQUFHLEVBQUU7b0NBQ1osT0FBTyxDQUFDLEdBQUcsQ0FBQywrQ0FBK0MsR0FBRyxFQUFFLENBQUMsYUFBYSxHQUFHLGlCQUFpQixFQUFFLEdBQUcsQ0FBQyxDQUFBO2lDQUN6Rzs2QkFDRjt5QkFDRjtxQkFDRjtvQkFBQyxPQUFPLEdBQUcsRUFBRTt3QkFDWixPQUFPLENBQUMsR0FBRyxDQUFDLG1DQUFtQyxHQUFHLEVBQUUsQ0FBQyxhQUFhLEdBQUcsaUJBQWlCLEVBQUUsR0FBRyxDQUFDLENBQUE7cUJBQzdGO2lCQUNGO3FCQUFNO29CQUNMLE9BQU8sQ0FBQyxHQUFHLENBQUMsR0FBRyxDQUFDLENBQUE7b0JBRWhCLE1BQU0sSUFBSSxLQUFLLENBQUMsMkJBQTJCLENBQUMsQ0FBQTtpQkFDN0M7YUFDRjtTQUNGO0tBQ0Y7SUFBQyxPQUFPLEdBQUcsRUFBRTtRQUNaLE9BQU8sQ0FBQyxHQUFHLENBQUMsdUNBQXVDLEVBQUUsR0FBRyxDQUFDLENBQUE7S0FDMUQ7QUFDSCxDQUFDLENBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgeyBJb3QgfSBmcm9tIFwiYXdzLXNka1wiXG5pbXBvcnQgeyBDbG91ZEV2ZW50IH0gZnJvbSBcImNsb3VkZXZlbnRzXCJcblxuY29uc3QgaW90ID0gbmV3IElvdCgpXG5cbmNvbnN0IHJldHVybklvVENvcmVDQVRlbXBsYXRlID0gKHBvbGljeU5hbWU6IHN0cmluZykgPT4ge1xuICBjb25zdCB0ZW1wbGF0ZUJvZHkgPSB7XG4gICAgUGFyYW1ldGVyczoge1xuICAgICAgXCJBV1M6OklvVDo6Q2VydGlmaWNhdGU6OkNvbW1vbk5hbWVcIjoge1xuICAgICAgICBUeXBlOiBcIlN0cmluZ1wiXG4gICAgICB9LFxuICAgICAgXCJBV1M6OklvVDo6Q2VydGlmaWNhdGU6OkNvdW50cnlcIjoge1xuICAgICAgICBUeXBlOiBcIlN0cmluZ1wiXG4gICAgICB9LFxuICAgICAgXCJBV1M6OklvVDo6Q2VydGlmaWNhdGU6OklkXCI6IHtcbiAgICAgICAgVHlwZTogXCJTdHJpbmdcIlxuICAgICAgfSxcbiAgICAgIFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlOjpTZXJpYWxOdW1iZXJcIjoge1xuICAgICAgICBUeXBlOiBcIlN0cmluZ1wiXG4gICAgICB9XG4gICAgfSxcbiAgICBSZXNvdXJjZXM6IHtcbiAgICAgIHRoaW5nOiB7XG4gICAgICAgIFR5cGU6IFwiQVdTOjpJb1Q6OlRoaW5nXCIsXG4gICAgICAgIFByb3BlcnRpZXM6IHtcbiAgICAgICAgICBUaGluZ05hbWU6IHtcbiAgICAgICAgICAgIFJlZjogXCJBV1M6OklvVDo6Q2VydGlmaWNhdGU6OkNvbW1vbk5hbWVcIlxuICAgICAgICAgIH0sXG4gICAgICAgICAgVGhpbmdHcm91cHM6IFtcIkxBTUFTU1VcIl0sXG4gICAgICAgICAgQXR0cmlidXRlUGF5bG9hZDoge31cbiAgICAgICAgfVxuICAgICAgfSxcbiAgICAgIGNlcnRpZmljYXRlOiB7XG4gICAgICAgIFR5cGU6IFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlXCIsXG4gICAgICAgIFByb3BlcnRpZXM6IHtcbiAgICAgICAgICBDZXJ0aWZpY2F0ZUlkOiB7XG4gICAgICAgICAgICBSZWY6IFwiQVdTOjpJb1Q6OkNlcnRpZmljYXRlOjpJZFwiXG4gICAgICAgICAgfSxcbiAgICAgICAgICBTdGF0dXM6IFwiQUNUSVZFXCJcbiAgICAgICAgfVxuICAgICAgfSxcbiAgICAgIHBvbGljeToge1xuICAgICAgICBUeXBlOiBcIkFXUzo6SW9UOjpQb2xpY3lcIixcbiAgICAgICAgUHJvcGVydGllczoge1xuICAgICAgICAgIFBvbGljeU5hbWU6IHBvbGljeU5hbWVcbiAgICAgICAgfVxuICAgICAgfVxuICAgIH1cbiAgfVxuICByZXR1cm4gdGVtcGxhdGVCb2R5XG59XG5cbmV4cG9ydCBjb25zdCBoYW5kbGVyID0gYXN5bmMgKGV2ZW50OiBhbnkpID0+IHtcbiAgY29uc29sZS5sb2coZXZlbnQpXG4gIGNvbnN0IHJlcXVlc3RlZENsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudDxhbnk+KGV2ZW50KVxuXG4gIHRyeSB7XG4gICAgY29uc3QgbGlzdENBc1Jlc3BvbnNlID0gYXdhaXQgaW90Lmxpc3RDQUNlcnRpZmljYXRlcyh7IGFzY2VuZGluZ09yZGVyOiB0cnVlLCBwYWdlU2l6ZTogMzAgfSkucHJvbWlzZSgpXG4gICAgY29uc29sZS5sb2cobGlzdENBc1Jlc3BvbnNlLmNlcnRpZmljYXRlcylcbiAgICBpZiAobGlzdENBc1Jlc3BvbnNlLmNlcnRpZmljYXRlcykge1xuICAgICAgZm9yIChsZXQgaSA9IDA7IGkgPCBsaXN0Q0FzUmVzcG9uc2UuY2VydGlmaWNhdGVzLmxlbmd0aDsgaSsrKSB7XG4gICAgICAgIGNvbnN0IGNhID0gbGlzdENBc1Jlc3BvbnNlLmNlcnRpZmljYXRlc1tpXVxuICAgICAgICBpZiAoY2EuY2VydGlmaWNhdGVBcm4gIT09IHVuZGVmaW5lZCAmJiBjYS5jZXJ0aWZpY2F0ZUlkICE9PSB1bmRlZmluZWQpIHtcbiAgICAgICAgICB0cnkge1xuICAgICAgICAgICAgY29uc3QgdGFnc1Jlc3BvbnNlID0gYXdhaXQgaW90Lmxpc3RUYWdzRm9yUmVzb3VyY2UoeyByZXNvdXJjZUFybjogY2EuY2VydGlmaWNhdGVBcm4gfSkucHJvbWlzZSgpXG4gICAgICAgICAgICBpZiAodGFnc1Jlc3BvbnNlLnRhZ3MpIHtcbiAgICAgICAgICAgICAgY29uc3QgbmFtZVRhZyA9IHRhZ3NSZXNwb25zZS50YWdzLmZpbmQodGFnID0+IHRhZy5LZXkgPT09IFwibGFtYXNzdUNBTmFtZVwiKVxuICAgICAgICAgICAgICBpZiAobmFtZVRhZyAmJiBuYW1lVGFnLlZhbHVlID09PSByZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEuY2FfbmFtZSkge1xuICAgICAgICAgICAgICAgIGNvbnN0IHBvbGljeU5hbWUgPSBcImxhbWFzc3VsYW1iZGFwb2xpY3lfXCIgKyByZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEuY2FfbmFtZSArIFwiX1wiICsgRGF0ZS5ub3coKVxuICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nKFwiY3JlYXRpbmcgcG9saWN5OiBcIiArIHBvbGljeU5hbWUpXG4gICAgICAgICAgICAgICAgdHJ5IHtcbiAgICAgICAgICAgICAgICAgIGF3YWl0IGlvdC5jcmVhdGVQb2xpY3koe1xuICAgICAgICAgICAgICAgICAgICBwb2xpY3lEb2N1bWVudDogSlNPTi5zdHJpbmdpZnkoSlNPTi5wYXJzZShyZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEucG9saWN5KSksXG4gICAgICAgICAgICAgICAgICAgIHBvbGljeU5hbWU6IHBvbGljeU5hbWUsXG4gICAgICAgICAgICAgICAgICAgIHRhZ3M6IFt7XG4gICAgICAgICAgICAgICAgICAgICAgS2V5OiBcInNlcmlhbE51bWJlclwiLFxuICAgICAgICAgICAgICAgICAgICAgIFZhbHVlOiByZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEuc2VyaWFsX251bWJlclxuICAgICAgICAgICAgICAgICAgICB9XVxuICAgICAgICAgICAgICAgICAgfSkucHJvbWlzZSgpXG5cbiAgICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nKFwidXBkYXRpbmcgQ0EgY2VydGlmaWNhdGUgd2l0aCBJRFwiICsgY2EuY2VydGlmaWNhdGVJZClcbiAgICAgICAgICAgICAgICAgIHRyeSB7XG4gICAgICAgICAgICAgICAgICAgIGF3YWl0IGlvdC51cGRhdGVDQUNlcnRpZmljYXRlKHtcbiAgICAgICAgICAgICAgICAgICAgICBjZXJ0aWZpY2F0ZUlkOiBjYS5jZXJ0aWZpY2F0ZUlkLFxuICAgICAgICAgICAgICAgICAgICAgIG5ld0F1dG9SZWdpc3RyYXRpb25TdGF0dXM6IFwiRU5BQkxFXCIsXG4gICAgICAgICAgICAgICAgICAgICAgbmV3U3RhdHVzOiBcIkFDVElWRVwiLFxuICAgICAgICAgICAgICAgICAgICAgIHJlZ2lzdHJhdGlvbkNvbmZpZzoge1xuICAgICAgICAgICAgICAgICAgICAgICAgcm9sZUFybjogYGFybjphd3M6aWFtOjoke3Byb2Nlc3MuZW52LkFXU19BQ0NPVU5UX0lEfTpyb2xlL0pJVFBSb2xlYCxcbiAgICAgICAgICAgICAgICAgICAgICAgIHRlbXBsYXRlQm9keTogSlNPTi5zdHJpbmdpZnkocmV0dXJuSW9UQ29yZUNBVGVtcGxhdGUocG9saWN5TmFtZSkpXG4gICAgICAgICAgICAgICAgICAgICAgfSxcbiAgICAgICAgICAgICAgICAgICAgICByZW1vdmVBdXRvUmVnaXN0cmF0aW9uOiBmYWxzZVxuICAgICAgICAgICAgICAgICAgICB9KS5wcm9taXNlKClcbiAgICAgICAgICAgICAgICAgIH0gY2F0Y2ggKGVycikge1xuICAgICAgICAgICAgICAgICAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIHVwZGF0aW5nIElvVCBDb3JlIENBIFtcIiArIGNhLmNlcnRpZmljYXRlSWQgKyBcIl0gZnJvbSBJb1QgQ29yZVwiLCBlcnIpXG4gICAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgICAgfSBjYXRjaCAoZXJyKSB7XG4gICAgICAgICAgICAgICAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIGNyZWF0aW5nIElvVCBDb3JlIFBvbGljeSBmb3IgQ0EgW1wiICsgY2EuY2VydGlmaWNhdGVJZCArIFwiXSBmcm9tIElvVCBDb3JlXCIsIGVycilcbiAgICAgICAgICAgICAgICB9XG4gICAgICAgICAgICAgIH1cbiAgICAgICAgICAgIH1cbiAgICAgICAgICB9IGNhdGNoIChlcnIpIHtcbiAgICAgICAgICAgIGNvbnNvbGUubG9nKFwiZXJyb3Igd2hpbGUgbGlzdGluZyB0YWdzIGZvciBDQSBbXCIgKyBjYS5jZXJ0aWZpY2F0ZUlkICsgXCJdIGZyb20gSW9UIENvcmVcIiwgZXJyKVxuICAgICAgICAgIH1cbiAgICAgICAgfSBlbHNlIHtcbiAgICAgICAgICBjb25zb2xlLmxvZyhcImJcIilcblxuICAgICAgICAgIHRocm93IG5ldyBFcnJvcihcIkNBIGNlcnRpZmljYXRlIElEIGlzIG51bGxcIilcbiAgICAgICAgfVxuICAgICAgfVxuICAgIH1cbiAgfSBjYXRjaCAoZXJyKSB7XG4gICAgY29uc29sZS5sb2coXCJlcnJvciB3aGlsZSBsaXN0aW5nIENBcyBmcm9tIElvVCBDb3JlXCIsIGVycilcbiAgfVxufVxuIl19