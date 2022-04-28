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
    const describeCertResponse = await iot.describeCertificate({ certificateId: certificateId }).promise();
    console.log("caCertificateID" + describeCertResponse.certificateDescription.caCertificateId);
    console.log("certificateId" + certificateId);
    const describeCAResponse = await iot.describeCACertificate({ certificateId: describeCertResponse.certificateDescription.caCertificateId }).promise();
    const caCert = x509_1.Certificate.fromPEM(Buffer.from(describeCAResponse.certificateDescription.certificatePem, "utf8"));
    const cert = x509_1.Certificate.fromPEM(Buffer.from(describeCertResponse.certificateDescription.certificatePem, "utf8"));
    const deviceID = cert.subject.commonName;
    console.log("deviceID: [" + deviceID + "] newStatus:[" + newStatus + "]");
    const cloudEvent = new cloudevents_1.CloudEvent({
        type: "io.lamassu.iotcore.cert.status.update",
        id: "",
        source: "aws/cloud-trail",
        time: new Date().toString(),
        specversion: "1.0",
        data: {
            ca_id: describeCertResponse.certificateDescription.caCertificateId,
            ca_name: caCert.subject.commonName,
            ca_serial_number: chunk(caCert.serialNumber, 2).join("-"),
            certificate_id: certificateId,
            serial_number: chunk(cert.serialNumber, 2).join("-"),
            device_id: deviceID,
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
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBa0M7QUFDbEMsNkNBQXdDO0FBQ3hDLHFDQUF3QztBQUV4QyxNQUFNLEdBQUcsR0FBRyxJQUFJLGFBQUcsRUFBRSxDQUFBO0FBQ3JCLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFUixRQUFBLE9BQU8sR0FBRyxLQUFLLEVBQUUsS0FBVSxFQUFFLEVBQUU7SUFDMUMsT0FBTyxDQUFDLEdBQUcsQ0FBQyxLQUFLLENBQUMsQ0FBQTtJQUNsQixNQUFNLGlCQUFpQixHQUFHLEtBQUssQ0FBQyxNQUFNLENBQUMsaUJBQWlCLENBQUE7SUFDeEQsTUFBTSxTQUFTLEdBQVcsaUJBQWlCLENBQUMsU0FBUyxDQUFBO0lBQ3JELE1BQU0sYUFBYSxHQUFXLGlCQUFpQixDQUFDLGFBQWEsQ0FBQTtJQUU3RCxNQUFNLG9CQUFvQixHQUFHLE1BQU0sR0FBRyxDQUFDLG1CQUFtQixDQUFDLEVBQUUsYUFBYSxFQUFFLGFBQWEsRUFBRSxDQUFDLENBQUMsT0FBTyxFQUFFLENBQUE7SUFFdEcsT0FBTyxDQUFDLEdBQUcsQ0FBQyxpQkFBaUIsR0FBRyxvQkFBb0IsQ0FBQyxzQkFBdUIsQ0FBQyxlQUFlLENBQUMsQ0FBQTtJQUM3RixPQUFPLENBQUMsR0FBRyxDQUFDLGVBQWUsR0FBRyxhQUFhLENBQUMsQ0FBQTtJQUU1QyxNQUFNLGtCQUFrQixHQUFHLE1BQU0sR0FBRyxDQUFDLHFCQUFxQixDQUFDLEVBQUUsYUFBYSxFQUFFLG9CQUFvQixDQUFDLHNCQUF1QixDQUFDLGVBQWdCLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO0lBRXRKLE1BQU0sTUFBTSxHQUFHLGtCQUFXLENBQUMsT0FBTyxDQUFDLE1BQU0sQ0FBQyxJQUFJLENBQUMsa0JBQWtCLENBQUMsc0JBQXVCLENBQUMsY0FBZSxFQUFFLE1BQU0sQ0FBQyxDQUFDLENBQUE7SUFDbkgsTUFBTSxJQUFJLEdBQUcsa0JBQVcsQ0FBQyxPQUFPLENBQUMsTUFBTSxDQUFDLElBQUksQ0FBQyxvQkFBb0IsQ0FBQyxzQkFBdUIsQ0FBQyxjQUFlLEVBQUUsTUFBTSxDQUFDLENBQUMsQ0FBQTtJQUVuSCxNQUFNLFFBQVEsR0FBRyxJQUFJLENBQUMsT0FBTyxDQUFDLFVBQVUsQ0FBQTtJQUN4QyxPQUFPLENBQUMsR0FBRyxDQUFDLGFBQWEsR0FBRyxRQUFRLEdBQUcsZUFBZSxHQUFHLFNBQVMsR0FBRyxHQUFHLENBQUMsQ0FBQTtJQUV6RSxNQUFNLFVBQVUsR0FBRyxJQUFJLHdCQUFVLENBQUM7UUFDaEMsSUFBSSxFQUFFLHVDQUF1QztRQUM3QyxFQUFFLEVBQUUsRUFBRTtRQUNOLE1BQU0sRUFBRSxpQkFBaUI7UUFDekIsSUFBSSxFQUFFLElBQUksSUFBSSxFQUFFLENBQUMsUUFBUSxFQUFFO1FBQzNCLFdBQVcsRUFBRSxLQUFLO1FBQ2xCLElBQUksRUFBRTtZQUNKLEtBQUssRUFBRSxvQkFBb0IsQ0FBQyxzQkFBdUIsQ0FBQyxlQUFlO1lBQ25FLE9BQU8sRUFBRSxNQUFNLENBQUMsT0FBTyxDQUFDLFVBQVU7WUFDbEMsZ0JBQWdCLEVBQUUsS0FBSyxDQUFDLE1BQU0sQ0FBQyxZQUFZLEVBQUUsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLEdBQUcsQ0FBQztZQUN6RCxjQUFjLEVBQUUsYUFBYTtZQUM3QixhQUFhLEVBQUUsS0FBSyxDQUFDLElBQUksQ0FBQyxZQUFZLEVBQUUsQ0FBQyxDQUFDLENBQUMsSUFBSSxDQUFDLEdBQUcsQ0FBQztZQUVwRCxTQUFTLEVBQUUsUUFBUTtZQUNuQixNQUFNLEVBQUUsU0FBUztTQUNsQjtLQUNGLENBQUMsQ0FBQTtJQUNGLElBQUk7UUFDRixNQUFNLFdBQVcsR0FBRyxNQUFNLEdBQUcsQ0FBQyxXQUFXLENBQUMsRUFBRSxRQUFRLEVBQUUsT0FBTyxDQUFDLEdBQUcsQ0FBQyxzQkFBdUIsRUFBRSxXQUFXLEVBQUUsVUFBVSxDQUFDLFFBQVEsRUFBRSxFQUFFLENBQUMsQ0FBQyxPQUFPLEVBQUUsQ0FBQTtRQUMxSSxPQUFPLENBQUMsR0FBRyxDQUFDLFdBQVcsQ0FBQyxDQUFBO0tBQ3pCO0lBQUMsT0FBTyxHQUFHLEVBQUU7UUFDWixPQUFPLENBQUMsR0FBRyxDQUFDLGlDQUFpQyxFQUFFLEdBQUcsQ0FBQyxDQUFBO0tBQ3BEO0FBQ0gsQ0FBQyxDQUFBO0FBRUQsU0FBUyxLQUFLLENBQUUsR0FBVyxFQUFFLENBQVM7SUFDcEMsTUFBTSxHQUFHLEdBQUcsRUFBRSxDQUFBO0lBQ2QsSUFBSSxDQUFDLENBQUE7SUFDTCxJQUFJLEdBQUcsQ0FBQTtJQUVQLEtBQUssQ0FBQyxHQUFHLENBQUMsRUFBRSxHQUFHLEdBQUcsR0FBRyxDQUFDLE1BQU0sRUFBRSxDQUFDLEdBQUcsR0FBRyxFQUFFLENBQUMsSUFBSSxDQUFDLEVBQUU7UUFDN0MsR0FBRyxDQUFDLElBQUksQ0FBQyxHQUFHLENBQUMsTUFBTSxDQUFDLENBQUMsRUFBRSxDQUFDLENBQUMsQ0FBQyxDQUFBO0tBQzNCO0lBRUQsT0FBTyxHQUFHLENBQUE7QUFDWixDQUFDIiwic291cmNlc0NvbnRlbnQiOlsiaW1wb3J0IHsgSW90LCBTUVMgfSBmcm9tIFwiYXdzLXNka1wiXG5pbXBvcnQgeyBDbG91ZEV2ZW50IH0gZnJvbSBcImNsb3VkZXZlbnRzXCJcbmltcG9ydCB7IENlcnRpZmljYXRlIH0gZnJvbSBcIkBmaWRtL3g1MDlcIlxuXG5jb25zdCBpb3QgPSBuZXcgSW90KClcbmNvbnN0IHNxcyA9IG5ldyBTUVMoKVxuXG5leHBvcnQgY29uc3QgaGFuZGxlciA9IGFzeW5jIChldmVudDogYW55KSA9PiB7XG4gIGNvbnNvbGUubG9nKGV2ZW50KVxuICBjb25zdCByZXF1ZXN0UGFyYW1ldGVycyA9IGV2ZW50LmRldGFpbC5yZXF1ZXN0UGFyYW1ldGVyc1xuICBjb25zdCBuZXdTdGF0dXM6IHN0cmluZyA9IHJlcXVlc3RQYXJhbWV0ZXJzLm5ld1N0YXR1c1xuICBjb25zdCBjZXJ0aWZpY2F0ZUlkOiBzdHJpbmcgPSByZXF1ZXN0UGFyYW1ldGVycy5jZXJ0aWZpY2F0ZUlkXG5cbiAgY29uc3QgZGVzY3JpYmVDZXJ0UmVzcG9uc2UgPSBhd2FpdCBpb3QuZGVzY3JpYmVDZXJ0aWZpY2F0ZSh7IGNlcnRpZmljYXRlSWQ6IGNlcnRpZmljYXRlSWQgfSkucHJvbWlzZSgpXG5cbiAgY29uc29sZS5sb2coXCJjYUNlcnRpZmljYXRlSURcIiArIGRlc2NyaWJlQ2VydFJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNhQ2VydGlmaWNhdGVJZClcbiAgY29uc29sZS5sb2coXCJjZXJ0aWZpY2F0ZUlkXCIgKyBjZXJ0aWZpY2F0ZUlkKVxuXG4gIGNvbnN0IGRlc2NyaWJlQ0FSZXNwb25zZSA9IGF3YWl0IGlvdC5kZXNjcmliZUNBQ2VydGlmaWNhdGUoeyBjZXJ0aWZpY2F0ZUlkOiBkZXNjcmliZUNlcnRSZXNwb25zZS5jZXJ0aWZpY2F0ZURlc2NyaXB0aW9uIS5jYUNlcnRpZmljYXRlSWQhIH0pLnByb21pc2UoKVxuXG4gIGNvbnN0IGNhQ2VydCA9IENlcnRpZmljYXRlLmZyb21QRU0oQnVmZmVyLmZyb20oZGVzY3JpYmVDQVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNlcnRpZmljYXRlUGVtISwgXCJ1dGY4XCIpKVxuICBjb25zdCBjZXJ0ID0gQ2VydGlmaWNhdGUuZnJvbVBFTShCdWZmZXIuZnJvbShkZXNjcmliZUNlcnRSZXNwb25zZS5jZXJ0aWZpY2F0ZURlc2NyaXB0aW9uIS5jZXJ0aWZpY2F0ZVBlbSEsIFwidXRmOFwiKSlcblxuICBjb25zdCBkZXZpY2VJRCA9IGNlcnQuc3ViamVjdC5jb21tb25OYW1lXG4gIGNvbnNvbGUubG9nKFwiZGV2aWNlSUQ6IFtcIiArIGRldmljZUlEICsgXCJdIG5ld1N0YXR1czpbXCIgKyBuZXdTdGF0dXMgKyBcIl1cIilcblxuICBjb25zdCBjbG91ZEV2ZW50ID0gbmV3IENsb3VkRXZlbnQoe1xuICAgIHR5cGU6IFwiaW8ubGFtYXNzdS5pb3Rjb3JlLmNlcnQuc3RhdHVzLnVwZGF0ZVwiLFxuICAgIGlkOiBcIlwiLFxuICAgIHNvdXJjZTogXCJhd3MvY2xvdWQtdHJhaWxcIixcbiAgICB0aW1lOiBuZXcgRGF0ZSgpLnRvU3RyaW5nKCksXG4gICAgc3BlY3ZlcnNpb246IFwiMS4wXCIsXG4gICAgZGF0YToge1xuICAgICAgY2FfaWQ6IGRlc2NyaWJlQ2VydFJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNhQ2VydGlmaWNhdGVJZCxcbiAgICAgIGNhX25hbWU6IGNhQ2VydC5zdWJqZWN0LmNvbW1vbk5hbWUsXG4gICAgICBjYV9zZXJpYWxfbnVtYmVyOiBjaHVuayhjYUNlcnQuc2VyaWFsTnVtYmVyLCAyKS5qb2luKFwiLVwiKSxcbiAgICAgIGNlcnRpZmljYXRlX2lkOiBjZXJ0aWZpY2F0ZUlkLFxuICAgICAgc2VyaWFsX251bWJlcjogY2h1bmsoY2VydC5zZXJpYWxOdW1iZXIsIDIpLmpvaW4oXCItXCIpLFxuXG4gICAgICBkZXZpY2VfaWQ6IGRldmljZUlELFxuICAgICAgc3RhdHVzOiBuZXdTdGF0dXNcbiAgICB9XG4gIH0pXG4gIHRyeSB7XG4gICAgY29uc3Qgc3FzUmVzcG9uc2UgPSBhd2FpdCBzcXMuc2VuZE1lc3NhZ2UoeyBRdWV1ZVVybDogcHJvY2Vzcy5lbnYuU1FTX1JFU1BPTlNFX1FVRVVFX1VSTCEsIE1lc3NhZ2VCb2R5OiBjbG91ZEV2ZW50LnRvU3RyaW5nKCkgfSkucHJvbWlzZSgpXG4gICAgY29uc29sZS5sb2coc3FzUmVzcG9uc2UpXG4gIH0gY2F0Y2ggKGVycikge1xuICAgIGNvbnNvbGUubG9nKFwiZXJyb3Igd2hpbGUgc2VuZGluZyBTUVMgbWVzc2dhZVwiLCBlcnIpXG4gIH1cbn1cblxuZnVuY3Rpb24gY2h1bmsgKHN0cjogc3RyaW5nLCBuOiBudW1iZXIpIHtcbiAgY29uc3QgcmV0ID0gW11cbiAgbGV0IGlcbiAgbGV0IGxlblxuXG4gIGZvciAoaSA9IDAsIGxlbiA9IHN0ci5sZW5ndGg7IGkgPCBsZW47IGkgKz0gbikge1xuICAgIHJldC5wdXNoKHN0ci5zdWJzdHIoaSwgbikpXG4gIH1cblxuICByZXR1cm4gcmV0XG59XG4iXX0=