"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.handler = void 0;
const aws_sdk_1 = require("aws-sdk");
const cloudevents_1 = require("cloudevents");
const x509_1 = require("@fidm/x509");
const iot = new aws_sdk_1.Iot();
const sqs = new aws_sdk_1.SQS();
exports.handler = async (event) => {
    console.log(event);
    const requestParameters = event.detail.requestParameters;
    const newStatus = requestParameters.newStatus;
    const certificateId = requestParameters.certificateId;
    const describeCAResponse = await iot.describeCACertificate({ certificateId: certificateId }).promise();
    const caCert = x509_1.Certificate.fromPEM(Buffer.from(describeCAResponse.certificateDescription.certificatePem, "utf8"));
    const cloudEvent = new cloudevents_1.CloudEvent({
        type: "io.lamassu.iotcore.ca.status.update",
        id: "",
        source: "aws/cloud-trail",
        time: new Date().toString(),
        specversion: "1.0",
        data: {
            ca_id: certificateId,
            ca_name: caCert.subject.commonName,
            ca_serial_number: chunk(caCert.serialNumber, 2).join("-"),
            status: newStatus
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
function chunk(str, n) {
    const ret = [];
    let i;
    let len;
    for (i = 0, len = str.length; i < len; i += n) {
        ret.push(str.substr(i, n));
    }
    return ret;
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBa0M7QUFDbEMsNkNBQXdDO0FBQ3hDLHFDQUF3QztBQUV4QyxNQUFNLEdBQUcsR0FBRyxJQUFJLGFBQUcsRUFBRSxDQUFBO0FBQ3JCLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFUixRQUFBLE9BQU8sR0FBRyxLQUFLLEVBQUUsS0FBVSxFQUFFLEVBQUU7SUFDMUMsT0FBTyxDQUFDLEdBQUcsQ0FBQyxLQUFLLENBQUMsQ0FBQTtJQUNsQixNQUFNLGlCQUFpQixHQUFHLEtBQUssQ0FBQyxNQUFNLENBQUMsaUJBQWlCLENBQUE7SUFDeEQsTUFBTSxTQUFTLEdBQVcsaUJBQWlCLENBQUMsU0FBUyxDQUFBO0lBQ3JELE1BQU0sYUFBYSxHQUFXLGlCQUFpQixDQUFDLGFBQWEsQ0FBQTtJQUU3RCxNQUFNLGtCQUFrQixHQUFHLE1BQU0sR0FBRyxDQUFDLHFCQUFxQixDQUFDLEVBQUUsYUFBYSxFQUFFLGFBQWEsRUFBRSxDQUFDLENBQUMsT0FBTyxFQUFFLENBQUE7SUFFdEcsTUFBTSxNQUFNLEdBQUcsa0JBQVcsQ0FBQyxPQUFPLENBQUMsTUFBTSxDQUFDLElBQUksQ0FBQyxrQkFBa0IsQ0FBQyxzQkFBdUIsQ0FBQyxjQUFlLEVBQUUsTUFBTSxDQUFDLENBQUMsQ0FBQTtJQUVuSCxNQUFNLFVBQVUsR0FBRyxJQUFJLHdCQUFVLENBQUM7UUFDaEMsSUFBSSxFQUFFLHFDQUFxQztRQUMzQyxFQUFFLEVBQUUsRUFBRTtRQUNOLE1BQU0sRUFBRSxpQkFBaUI7UUFDekIsSUFBSSxFQUFFLElBQUksSUFBSSxFQUFFLENBQUMsUUFBUSxFQUFFO1FBQzNCLFdBQVcsRUFBRSxLQUFLO1FBQ2xCLElBQUksRUFBRTtZQUNKLEtBQUssRUFBRSxhQUFhO1lBQ3BCLE9BQU8sRUFBRSxNQUFNLENBQUMsT0FBTyxDQUFDLFVBQVU7WUFDbEMsZ0JBQWdCLEVBQUUsS0FBSyxDQUFDLE1BQU0sQ0FBQyxZQUFZLEVBQUUsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLEdBQUcsQ0FBQztZQUN6RCxNQUFNLEVBQUUsU0FBUztTQUNsQjtLQUNGLENBQUMsQ0FBQTtJQUNGLElBQUk7UUFDRixNQUFNLFdBQVcsR0FBRyxNQUFNLEdBQUcsQ0FBQyxXQUFXLENBQUMsRUFBRSxRQUFRLEVBQUUsT0FBTyxDQUFDLEdBQUcsQ0FBQyxzQkFBdUIsRUFBRSxXQUFXLEVBQUUsVUFBVSxDQUFDLFFBQVEsRUFBRSxFQUFFLENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTtRQUMxSSxPQUFPLENBQUMsR0FBRyxDQUFDLFdBQVcsQ0FBQyxDQUFBO0tBQ3pCO0lBQUMsT0FBTyxHQUFHLEVBQUU7UUFDWixPQUFPLENBQUMsR0FBRyxDQUFDLGlDQUFpQyxFQUFFLEdBQUcsQ0FBQyxDQUFBO0tBQ3BEO0FBQ0gsQ0FBQyxDQUFBO0FBRUQsU0FBUyxLQUFLLENBQUUsR0FBVyxFQUFFLENBQVM7SUFDcEMsTUFBTSxHQUFHLEdBQUcsRUFBRSxDQUFBO0lBQ2QsSUFBSSxDQUFDLENBQUE7SUFDTCxJQUFJLEdBQUcsQ0FBQTtJQUVQLEtBQUssQ0FBQyxHQUFHLENBQUMsRUFBRSxHQUFHLEdBQUcsR0FBRyxDQUFDLE1BQU0sRUFBRSxDQUFDLEdBQUcsR0FBRyxFQUFFLENBQUMsSUFBSSxDQUFDLEVBQUU7UUFDN0MsR0FBRyxDQUFDLElBQUksQ0FBQyxHQUFHLENBQUMsTUFBTSxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxDQUFBO0tBQzNCO0lBRUQsT0FBTyxHQUFHLENBQUE7QUFDWixDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgSW90LCBTUVMgfSBmcm9tIFwiYXdzLXNka1wiXG5pbXBvcnQgeyBDbG91ZEV2ZW50IH0gZnJvbSBcImNsb3VkZXZlbnRzXCJcbmltcG9ydCB7IENlcnRpZmljYXRlIH0gZnJvbSBcIkBmaWRtL3g1MDlcIlxuXG5jb25zdCBpb3QgPSBuZXcgSW90KClcbmNvbnN0IHNxcyA9IG5ldyBTUVMoKVxuXG5leHBvcnQgY29uc3QgaGFuZGxlciA9IGFzeW5jIChldmVudDogYW55KSA9PiB7XG4gIGNvbnNvbGUubG9nKGV2ZW50KVxuICBjb25zdCByZXF1ZXN0UGFyYW1ldGVycyA9IGV2ZW50LmRldGFpbC5yZXF1ZXN0UGFyYW1ldGVyc1xuICBjb25zdCBuZXdTdGF0dXM6IHN0cmluZyA9IHJlcXVlc3RQYXJhbWV0ZXJzLm5ld1N0YXR1c1xuICBjb25zdCBjZXJ0aWZpY2F0ZUlkOiBzdHJpbmcgPSByZXF1ZXN0UGFyYW1ldGVycy5jZXJ0aWZpY2F0ZUlkXG5cbiAgY29uc3QgZGVzY3JpYmVDQVJlc3BvbnNlID0gYXdhaXQgaW90LmRlc2NyaWJlQ0FDZXJ0aWZpY2F0ZSh7IGNlcnRpZmljYXRlSWQ6IGNlcnRpZmljYXRlSWQgfSkucHJvbWlzZSgpXG5cbiAgY29uc3QgY2FDZXJ0ID0gQ2VydGlmaWNhdGUuZnJvbVBFTShCdWZmZXIuZnJvbShkZXNjcmliZUNBUmVzcG9uc2UuY2VydGlmaWNhdGVEZXNjcmlwdGlvbiEuY2VydGlmaWNhdGVQZW0hLCBcInV0ZjhcIikpXG5cbiAgY29uc3QgY2xvdWRFdmVudCA9IG5ldyBDbG91ZEV2ZW50KHtcbiAgICB0eXBlOiBcImlvLmxhbWFzc3UuaW90Y29yZS5jYS5zdGF0dXMudXBkYXRlXCIsXG4gICAgaWQ6IFwiXCIsXG4gICAgc291cmNlOiBcImF3cy9jbG91ZC10cmFpbFwiLFxuICAgIHRpbWU6IG5ldyBEYXRlKCkudG9TdHJpbmcoKSxcbiAgICBzcGVjdmVyc2lvbjogXCIxLjBcIixcbiAgICBkYXRhOiB7XG4gICAgICBjYV9pZDogY2VydGlmaWNhdGVJZCxcbiAgICAgIGNhX25hbWU6IGNhQ2VydC5zdWJqZWN0LmNvbW1vbk5hbWUsXG4gICAgICBjYV9zZXJpYWxfbnVtYmVyOiBjaHVuayhjYUNlcnQuc2VyaWFsTnVtYmVyLCAyKS5qb2luKFwiLVwiKSxcbiAgICAgIHN0YXR1czogbmV3U3RhdHVzXG4gICAgfVxuICB9KVxuICB0cnkge1xuICAgIGNvbnN0IHNxc1Jlc3BvbnNlID0gYXdhaXQgc3FzLnNlbmRNZXNzYWdlKHsgUXVldWVVcmw6IHByb2Nlc3MuZW52LlNRU19SRVNQT05TRV9RVUVVRV9VUkwhLCBNZXNzYWdlQm9keTogY2xvdWRFdmVudC50b1N0cmluZygpIH0pLnByb21pc2UoKVxuICAgIGNvbnNvbGUubG9nKHNxc1Jlc3BvbnNlKVxuICB9IGNhdGNoIChlcnIpIHtcbiAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIHNlbmRpbmcgU1FTIG1lc3NnYWVcIiwgZXJyKVxuICB9XG59XG5cbmZ1bmN0aW9uIGNodW5rIChzdHI6IHN0cmluZywgbjogbnVtYmVyKSB7XG4gIGNvbnN0IHJldCA9IFtdXG4gIGxldCBpXG4gIGxldCBsZW5cblxuICBmb3IgKGkgPSAwLCBsZW4gPSBzdHIubGVuZ3RoOyBpIDwgbGVuOyBpICs9IG4pIHtcbiAgICByZXQucHVzaChzdHIuc3Vic3RyKGksIG4pKVxuICB9XG5cbiAgcmV0dXJuIHJldFxufVxuIl19