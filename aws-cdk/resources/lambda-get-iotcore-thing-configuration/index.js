"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handler = void 0;
const aws_sdk_1 = require("aws-sdk");
const cloudevents_1 = require("cloudevents");
const x509_1 = require("@fidm/x509");
const iot = new aws_sdk_1.Iot();
const sqs = new aws_sdk_1.SQS();
exports.handler = async (event) => {
    var _a;
    const requestedCloudEvent = new cloudevents_1.CloudEvent(event);
    const deviceID = requestedCloudEvent.data.device_id;
    console.log(deviceID);
    const things = [];
    const searchResponse = await iot.searchIndex({ queryString: `thingGroup:${deviceID}` }).promise();
    console.log(searchResponse);
    for (const thing of searchResponse.things) {
        const principalResponse = await iot.listThingPrincipals({ thingName: thing.thingName, maxResults: 25 }).promise();
        const certificates = [];
        const principals = principalResponse.principals;
        for (const principal of principals) {
            const splitiedPrincipal = principal.split(":");
            const certificateID = splitiedPrincipal[splitiedPrincipal.length - 1].replace("cert/", "");
            const certificateResponse = await iot.describeCertificate({ certificateId: certificateID }).promise();
            const certPem = certificateResponse.certificateDescription.certificatePem;
            const caCert = x509_1.Certificate.fromPEM(Buffer.from(certPem, "utf8"));
            certificates.push({
                serial_number: chunk(caCert.serialNumber, 2).join("-"),
                status: certificateResponse.certificateDescription.status,
                arn: certificateResponse.certificateDescription.certificateArn,
                id: certificateResponse.certificateDescription.certificateId,
                update_date: certificateResponse.certificateDescription.lastModifiedDate
            });
        }
        things.push({
            aws_id: thing.thingId,
            device_id: thing.thingName,
            last_connection: (_a = thing.connectivity) === null || _a === void 0 ? void 0 : _a.timestamp,
            certificates: certificates
        });
    }
    const cloudEvent = new cloudevents_1.CloudEvent({
        type: "io.lamassu.iotcore.thing.config.response",
        id: requestedCloudEvent.id,
        source: "aws/lambda",
        time: new Date().toString(),
        specversion: "1.0",
        data: things
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
function chunk(str, n) {
    const ret = [];
    let i;
    let len;
    for (i = 0, len = str.length; i < len; i += n) {
        ret.push(str.substr(i, n));
    }
    return ret;
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBa0M7QUFDbEMsNkNBQXdDO0FBQ3hDLHFDQUF3QztBQUV4QyxNQUFNLEdBQUcsR0FBRyxJQUFJLGFBQUcsRUFBRSxDQUFBO0FBQ3JCLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFUixRQUFBLE9BQU8sR0FBRyxLQUFLLEVBQUUsS0FBVSxFQUFFLEVBQUU7O0lBQzFDLE1BQU0sbUJBQW1CLEdBQUcsSUFBSSx3QkFBVSxDQUFNLEtBQUssQ0FBQyxDQUFBO0lBQ3RELE1BQU0sUUFBUSxHQUFHLG1CQUFtQixDQUFDLElBQUksQ0FBQyxTQUFTLENBQUE7SUFDbkQsT0FBTyxDQUFDLEdBQUcsQ0FBQyxRQUFRLENBQUMsQ0FBQTtJQUVyQixNQUFNLE1BQU0sR0FBZSxFQUFFLENBQUE7SUFDN0IsTUFBTSxjQUFjLEdBQUcsTUFBTSxHQUFHLENBQUMsV0FBVyxDQUFDLEVBQUUsV0FBVyxFQUFFLGNBQWMsUUFBUSxFQUFFLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO0lBQ2pHLE9BQU8sQ0FBQyxHQUFHLENBQUMsY0FBYyxDQUFDLENBQUE7SUFFM0IsS0FBSyxNQUFNLEtBQUssSUFBSSxjQUFjLENBQUMsTUFBTyxFQUFFO1FBQzFDLE1BQU0saUJBQWlCLEdBQUcsTUFBTSxHQUFHLENBQUMsbUJBQW1CLENBQUMsRUFBRSxTQUFTLEVBQUUsS0FBSyxDQUFDLFNBQVUsRUFBRSxVQUFVLEVBQUUsRUFBRSxFQUFFLENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTtRQUNsSCxNQUFNLFlBQVksR0FBRyxFQUFFLENBQUE7UUFDdkIsTUFBTSxVQUFVLEdBQUcsaUJBQWlCLENBQUMsVUFBVSxDQUFBO1FBQy9DLEtBQUssTUFBTSxTQUFTLElBQUksVUFBVyxFQUFFO1lBQ25DLE1BQU0saUJBQWlCLEdBQUcsU0FB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IFwiYXdzLXNka1wiXG5pbXBvcnQgeyBDbG91ZEV2ZW50IH0gZnJvbSBcImNsb3VkZXZlbnRzXCJcbmltcG9ydCB7IENlcnRpZmljYXRlIH0gZnJvbSBcIkBmaWRtL3g1MDlcIlxuXG5jb25zdCBpb3QgPSBuZXcgSW90KClcbmNvbnN0IHNxcyA9IG5ldyBTUVMoKVxuXG5leHBvcnQgY29uc3QgaGFuZGxlciA9IGFzeW5jIChldmVudDogYW55KSA9PiB7XG4gIGNvbnN0IHJlcXVlc3RlZENsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudDxhbnk+KGV2ZW50KVxuICBjb25zdCBkZXZpY2VJRCA9IHJlcXVlc3RlZENsb3VkRXZlbnQuZGF0YS5kZXZpY2VfaWRcbiAgY29uc29sZS5sb2coZGV2aWNlSUQpXG5cbiAgY29uc3QgdGhpbmdzOiBBcnJheTxhbnk+ID0gW11cbiAgY29uc3Qgc2VhcmNoUmVzcG9uc2UgPSBhd2FpdCBpb3Quc2VhcmNoSW5kZXgoeyBxdWVyeVN0cmluZzogYHRoaW5nR3JvdXA6JHtkZXZpY2VJRH1gIH0pLnByb21pc2UoKVxuICBjb25zb2xlLmxvZyhzZWFyY2hSZXNwb25zZSlcblxuICBmb3IgKGNvbnN0IHRoaW5nIG9mIHNlYXJjaFJlc3BvbnNlLnRoaW5ncyEpIHtcbiAgICBjb25zdCBwcmluY2lwYWxSZXNwb25zZSA9IGF3YWl0IGlvdC5saXN0VGhpbmdQcmluY2lwYWxzKHsgdGhpbmdOYW1lOiB0aGluZy50aGluZ05hbWUhLCBtYXhSZXN1bHRzOiAyNSB9KS5wcm9taXNlKClcbiAgICBjb25zdCBjZXJ0aWZpY2F0ZXMgPSBbXVxuICAgIGNvbnN0IHByaW5jaXBhbHMgPSBwcmluY2lwYWxSZXNwb25zZS5wcmluY2lwYWxzXG4gICAgZm9yIChjb25zdCBwcmluY2lwYWwgb2YgcHJpbmNpcGFscyEpIHtcbiAgICAgIGNvbnN0IHNwbGl0aWVkUHJpbmNpcGFsID0gcHJpbmNpcGFsLnNwbGl0KFwiOlwiKVxuICAgICAgY29uc3QgY2VydGlmaWNhdGVJRCA9IHNwbGl0aWVkUHJpbmNpcGFsW3NwbGl0aWVkUHJpbmNpcGFsLmxlbmd0aCAtIDFdLnJlcGxhY2UoXCJjZXJ0L1wiLCBcIlwiKVxuICAgICAgY29uc3QgY2VydGlmaWNhdGVSZXNwb25zZSA9IGF3YWl0IGlvdC5kZXNjcmliZUNlcnRpZmljYXRlKHsgY2VydGlmaWNhdGVJZDogY2VydGlmaWNhdGVJRCB9KS5wcm9taXNlKClcbiAgICAgIGNvbnN0IGNlcnRQZW0gPSBjZXJ0aWZpY2F0ZVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNlcnRpZmljYXRlUGVtIVxuXG4gICAgICBjb25zdCBjYUNlcnQgPSBDZXJ0aWZpY2F0ZS5mcm9tUEVNKEJ1ZmZlci5mcm9tKGNlcnRQZW0sIFwidXRmOFwiKSlcblxuICAgICAgY2VydGlmaWNhdGVzLnB1c2goe1xuICAgICAgICBzZXJpYWxfbnVtYmVyOiBjaHVuayhjYUNlcnQuc2VyaWFsTnVtYmVyLCAyKS5qb2luKFwiLVwiKSxcbiAgICAgICAgc3RhdHVzOiBjZXJ0aWZpY2F0ZVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLnN0YXR1cyxcbiAgICAgICAgYXJuOiBjZXJ0aWZpY2F0ZVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNlcnRpZmljYXRlQXJuLFxuICAgICAgICBpZDogY2VydGlmaWNhdGVSZXNwb25zZS5jZXJ0aWZpY2F0ZURlc2NyaXB0aW9uIS5jZXJ0aWZpY2F0ZUlkLFxuICAgICAgICB1cGRhdGVfZGF0ZTogY2VydGlmaWNhdGVSZXNwb25zZS5jZXJ0aWZpY2F0ZURlc2NyaXB0aW9uIS5sYXN0TW9kaWZpZWREYXRlXG4gICAgICB9KVxuICAgIH1cblxuICAgIHRoaW5ncy5wdXNoKFxuICAgICAge1xuICAgICAgICBhd3NfaWQ6IHRoaW5nLnRoaW5nSWQsXG4gICAgICAgIGRldmljZV9pZDogdGhpbmcudGhpbmdOYW1lLFxuICAgICAgICBsYXN0X2Nvbm5lY3Rpb246IHRoaW5nLmNvbm5lY3Rpdml0eT8udGltZXN0YW1wLFxuICAgICAgICBjZXJ0aWZpY2F0ZXM6IGNlcnRpZmljYXRlc1xuICAgICAgfVxuICAgIClcbiAgfVxuXG4gIGNvbnN0IGNsb3VkRXZlbnQgPSBuZXcgQ2xvdWRFdmVudCh7XG4gICAgdHlwZTogXCJpby5sYW1hc3N1LmlvdGNvcmUudGhpbmcuY29uZmlnLnJlc3BvbnNlXCIsXG4gICAgaWQ6IHJlcXVlc3RlZENsb3VkRXZlbnQuaWQsXG4gICAgc291cmNlOiBcImF3cy9sYW1iZGFcIixcbiAgICB0aW1lOiBuZXcgRGF0ZSgpLnRvU3RyaW5nKCksXG4gICAgc3BlY3ZlcnNpb246IFwiMS4wXCIsXG4gICAgZGF0YTogdGhpbmdzXG4gIH0pXG4gIHRyeSB7XG4gICAgY29uc29sZS5sb2coXCJDbG91ZCBFdmVudFwiLCBjbG91ZEV2ZW50KVxuICAgIGNvbnN0IHNxc1Jlc3BvbnNlID0gYXdhaXQgc3FzLnNlbmRNZXNzYWdlKHsgUXVldWVVcmw6IHByb2Nlc3MuZW52LlNRU19SRVNQT05TRV9RVUVVRV9VUkwhLCBNZXNzYWdlQm9keTogY2xvdWRFdmVudC50b1N0cmluZygpIH0pLnByb21pc2UoKVxuICAgIGNvbnNvbGUubG9nKHNxc1Jlc3BvbnNlKVxuICB9IGNhdGNoIChlcnIpIHtcbiAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIHNlbmRpbmcgU1FTIG1lc3NnYWVcIiwgZXJyKVxuICB9XG59XG5cbmZ1bmN0aW9uIGNodW5rIChzdHI6IHN0cmluZywgbjogbnVtYmVyKSB7XG4gIGNvbnN0IHJldCA9IFtdXG4gIGxldCBpXG4gIGxldCBsZW5cblxuICBmb3IgKGkgPSAwLCBsZW4gPSBzdHIubGVuZ3RoOyBpIDwgbGVuOyBpICs9IG4pIHtcbiAgICByZXQucHVzaChzdHIuc3Vic3RyKGksIG4pKVxuICB9XG5cbiAgcmV0dXJuIHJldFxufVxuIl19