"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handler = void 0;
const aws_sdk_1 = require("aws-sdk");
const cloudevents_1 = require("cloudevents");
const iot = new aws_sdk_1.Iot();
exports.handler = async (event) => {
    console.log(event);
    const requestedCloudEvent = new cloudevents_1.CloudEvent(event);
    try {
        const response = await iot.registerCACertificate({
            caCertificate: Buffer.from(requestedCloudEvent.data.ca_cert, "base64").toString(),
            verificationCertificate: Buffer.from(requestedCloudEvent.data.verification_cert, "base64").toString(),
            setAsActive: false,
            allowAutoRegistration: true,
            tags: [{
                    Key: "lamassuCAName",
                    Value: requestedCloudEvent.data.ca_name
                },
                {
                    Key: "serialNumber",
                    Value: requestedCloudEvent.data.serial_number
                }]
        }).promise();
        console.log(response);
    }
    catch (err) {
        console.log("error while registerting IoT Core CA", err);
    }
};
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBNkI7QUFDN0IsNkNBQXdDO0FBRXhDLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFUixRQUFBLE9BQU8sR0FBRyxLQUFLLEVBQUUsS0FBVSxFQUFFLEVBQUU7SUFDMUMsT0FBTyxDQUFDLEdBQUcsQ0FBQyxLQUFLLENBQUMsQ0FBQTtJQUVsQixNQUFNLG1CQUFtQixHQUFHLElBQUksd0JBQVUsQ0FBTSxLQUFLLENBQUMsQ0FBQTtJQUV0RCxJQUFJO1FBQ0YsTUFBTSxRQUFRLEdBQUcsTUFBTSxHQUFHLENBQUMscUJBQXFCLENBQUM7WUFDL0MsYUFBYSxFQUFFLE1BQU0sQ0FBQyxJQUFJLENBQUMsbUJBQW1CLENBQUMsSUFBSSxDQUFDLE9BQU8sRUFBRSxRQUFRLENBQUMsQ0FBQyxRQUFRLEVBQUU7WUFDakYsdUJBQXVCLEVBQUUsTUFBTSxDQUFDLElBQUksQ0FBQyxtQkFBbUIsQ0FBQyxJQUFJLENBQUMsaUJBQWlCLEVBQUUsUUFBUSxDQUFDLENBQUMsUUFBUSxFQUFFO1lBQ3JHLFdBQVcsRUFBRSxLQUFLO1lBQ2xCLHFCQUFxQixFQUFFLElBQUk7WUFDM0IsSUFBSSxFQUFFLENBQUM7b0JBQ0wsR0FBRyxFQUFFLGVBQWU7b0JBQ3BCLEtBQUssRUFBRSxtQkFBbUIsQ0FBQyxJQUFJLENBQUMsT0FBTztpQkFDeEM7Z0JBQ0Q7b0JBQ0UsR0FBRyxFQUFFLGNBQWM7b0JBQ25CLEtBQUssRUFBRSxtQkFBbUIsQ0FBQyxJQUFJLENBQUMsYUFBYTtpQkFDOUMsQ0FBQztTQUNILENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTtRQUVaLE9BQU8sQ0FBQyxHQUFHLENBQUMsUUFBUSxDQUFDLENBQUE7S0FDdEI7SUFBQyxPQUFPLEdBQUcsRUFBRTtRQUNaLE9BQU8sQ0FBQyxHQUFHLENBQUMsc0NBQXNDLEVBQUUsR0FBRyxDQUFDLENBQUE7S0FDekQ7QUFDSCxDQUFDLENBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQgeyBJb3QgfSBmcm9tIFwiYXdzLXNka1wiXG5pbXBvcnQgeyBDbG91ZEV2ZW50IH0gZnJvbSBcImNsb3VkZXZlbnRzXCJcblxuY29uc3QgaW90ID0gbmV3IElvdCgpXG5cbmV4cG9ydCBjb25zdCBoYW5kbGVyID0gYXN5bmMgKGV2ZW50OiBhbnkpID0+IHtcbiAgY29uc29sZS5sb2coZXZlbnQpXG5cbiAgY29uc3QgcmVxdWVzdGVkQ2xvdWRFdmVudCA9IG5ldyBDbG91ZEV2ZW50PGFueT4oZXZlbnQpXG5cbiAgdHJ5IHtcbiAgICBjb25zdCByZXNwb25zZSA9IGF3YWl0IGlvdC5yZWdpc3RlckNBQ2VydGlmaWNhdGUoe1xuICAgICAgY2FDZXJ0aWZpY2F0ZTogQnVmZmVyLmZyb20ocmVxdWVzdGVkQ2xvdWRFdmVudC5kYXRhLmNhX2NlcnQsIFwiYmFzZTY0XCIpLnRvU3RyaW5nKCksXG4gICAgICB2ZXJpZmljYXRpb25DZXJ0aWZpY2F0ZTogQnVmZmVyLmZyb20ocmVxdWVzdGVkQ2xvdWRFdmVudC5kYXRhLnZlcmlmaWNhdGlvbl9jZXJ0LCBcImJhc2U2NFwiKS50b1N0cmluZygpLFxuICAgICAgc2V0QXNBY3RpdmU6IGZhbHNlLFxuICAgICAgYWxsb3dBdXRvUmVnaXN0cmF0aW9uOiB0cnVlLFxuICAgICAgdGFnczogW3tcbiAgICAgICAgS2V5OiBcImxhbWFzc3VDQU5hbWVcIixcbiAgICAgICAgVmFsdWU6IHJlcXVlc3RlZENsb3VkRXZlbnQuZGF0YS5jYV9uYW1lXG4gICAgICB9LFxuICAgICAge1xuICAgICAgICBLZXk6IFwic2VyaWFsTnVtYmVyXCIsXG4gICAgICAgIFZhbHVlOiByZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEuc2VyaWFsX251bWJlclxuICAgICAgfV1cbiAgICB9KS5wcm9taXNlKClcblxuICAgIGNvbnNvbGUubG9nKHJlc3BvbnNlKVxuICB9IGNhdGNoIChlcnIpIHtcbiAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIHJlZ2lzdGVydGluZyBJb1QgQ29yZSBDQVwiLCBlcnIpXG4gIH1cbn1cbiJdfQ==