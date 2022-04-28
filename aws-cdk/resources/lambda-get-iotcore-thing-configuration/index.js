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
    let thingResult;
    const searchResponse = await iot.searchIndex({ queryString: `thingName:${deviceID}` }).promise();
    if (searchResponse.things.length === 0) {
        console.log("No results with device ID (" + deviceID + ")");
        thingResult = {
            device_id: deviceID,
            status: 404
        };
    }
    else if (searchResponse.things.length !== 1) {
        console.log("Inconsistent thing repo: More than one result for the same device ID (" + deviceID + ")");
        thingResult = {
            device_id: deviceID,
            status: 409
        };
    }
    else {
        const thing = searchResponse.things[0];
        const principalResponse = await iot.listThingPrincipals({ thingName: thing.thingName, maxResults: 25 }).promise();
        const certificates = [];
        const principals = principalResponse.principals;
        for (const principal of principals) {
            const splitiedPrincipal = principal.split(":");
            const certificateID = splitiedPrincipal[splitiedPrincipal.length - 1].replace("cert/", "");
            const certificateResponse = await iot.describeCertificate({ certificateId: certificateID }).promise();
            const certPem = certificateResponse.certificateDescription.certificatePem;
            const cert = x509_1.Certificate.fromPEM(Buffer.from(certPem, "utf8"));
            certificates.push({
                serial_number: chunk(cert.serialNumber, 2).join("-"),
                status: certificateResponse.certificateDescription.status,
                arn: certificateResponse.certificateDescription.certificateArn,
                id: certificateResponse.certificateDescription.certificateId,
                update_date: certificateResponse.certificateDescription.lastModifiedDate,
                ca_name: cert.issuer.commonName
            });
        }
        thingResult = {
            device_id: thing.thingName,
            status: 200,
            config: {
                aws_id: thing.thingId,
                last_connection: (_a = thing.connectivity) === null || _a === void 0 ? void 0 : _a.timestamp,
                certificates: certificates
            }
        };
    }
    const cloudEvent = new cloudevents_1.CloudEvent({
        type: "io.lamassu.iotcore.thing.config.response",
        id: requestedCloudEvent.id,
        source: "aws/lambda",
        time: new Date().toString(),
        specversion: "1.0",
        data: thingResult
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
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyJpbmRleC50cyJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiOzs7QUFBQSxxQ0FBa0M7QUFDbEMsNkNBQXdDO0FBQ3hDLHFDQUF3QztBQUV4QyxNQUFNLEdBQUcsR0FBRyxJQUFJLGFBQUcsRUFBRSxDQUFBO0FBQ3JCLE1BQU0sR0FBRyxHQUFHLElBQUksYUFBRyxFQUFFLENBQUE7QUFFUixRQUFBLE9BQU8sR0FBRyxLQUFLLEVBQUUsS0FBVSxFQUFFLEVBQUU7O0lBQzFDLE1BQU0sbUJBQW1CLEdBQUcsSUFBSSx3QkFBVSxDQUFNLEtBQUssQ0FBQyxDQUFBO0lBQ3RELE1BQU0sUUFBUSxHQUFHLG1CQUFtQixDQUFDLElBQUksQ0FBQyxTQUFTLENBQUE7SUFDbkQsT0FBTyxDQUFDLEdBQUcsQ0FBQyxRQUFRLENBQUMsQ0FBQTtJQUVyQixJQUFJLFdBQTRCLENBQUE7SUFDaEMsTUFBTSxjQUFjLEdBQUcsTUFBTSxHQUFHLENBQUMsV0FBVyxDQUFDLEVBQUUsV0FBVyxFQUFFLGFBQWEsUUFBUSxFQUFFLEVBQUUsQ0FBQyxDQUFDLE9BQU8sRUFBRSxDQUFBO0lBQ2hHLElBQUksY0FBYyxDQUFDLE1BQU8sQ0FBQyxNQUFNLEtBQUssQ0FBQyxFQUFFO1FBQ3ZDLE9BQU8sQ0FBQyxHQUFHLENBQUMsNkJBQTZCLEdBQUcsUUFBUSxHQUFHLEdBQUcsQ0FBQyxDQUFBO1FBQzNELFdBQVcsR0FBRztZQUNaLFNBQVMsRUFBRSxRQUFRO1lBQ25CLE1BQU0sRUFBRSxHQUFHO1NBQ1osQ0FBQTtLQUNGO1NBQU0sSUFBSSxjQUFjLENBQUMsTUFBTyxDQUFDLE1BQU0sS0FBSyxDQUFDLEVBQUU7UUFDOUMsT0FBTyxDQUFDLEdBQUcsQ0FBQyx3RUFBd0UsR0FBRyxRQUFRLEdBQUcsR0FBRyxDQUFDLENBQUE7UUFDdEcsV0FBVyxHQUFHO1lBQ1osU0FBUyxFQUFFLFFBQVE7WUFDbkIsTUFBTSxFQUFFLEdBQUc7U0FDWixD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aUMsRUFBRSxHQUFHLENBQUMsQ0FBQTtLQUNwRDtBQUNILENBQUMsQ0FBQTtBQUVELFNBQVMsS0FBSyxDQUFFLEdBQVcsRUFBRSxDQUFTO0lBQ3BDLE1BQU0sR0FBRyxHQUFHLEVBQUUsQ0FBQTtJQUNkLElBQUksQ0FBQyxDQUFBO0lBQ0wsSUFBSSxHQUFHLENBQUE7SUFFUCxLQUFLLENBQUMsR0FBRyxDQUFDLEVBQUUsR0FBRyxHQUFHLEdBQUcsQ0FBQyxNQUFNLEVBQUUsQ0FBQyxHQUFHLEdBQUcsRUFBRSxDQUFDLElBQUksQ0FBQyxFQUFFO1FBQzdDLEdBQUcsQ0FBQyxJQUFJLENBQUMsR0FBRyxDQUFDLE1BQU0sQ0FBQyxDQUFDLEVBQUUsQ0FBQyxDQUFDLENBQUMsQ0FBQTtLQUMzQjtJQUVELE9BQU8sR0FBRyxDQUFBO0FBQ1osQ0FBQyIsInNvdXJjZXNDb250ZW50IjpbImltcG9ydCB7IElvdCwgU1FTIH0gZnJvbSBcImF3cy1zZGtcIlxuaW1wb3J0IHsgQ2xvdWRFdmVudCB9IGZyb20gXCJjbG91ZGV2ZW50c1wiXG5pbXBvcnQgeyBDZXJ0aWZpY2F0ZSB9IGZyb20gXCJAZmlkbS94NTA5XCJcblxuY29uc3QgaW90ID0gbmV3IElvdCgpXG5jb25zdCBzcXMgPSBuZXcgU1FTKClcblxuZXhwb3J0IGNvbnN0IGhhbmRsZXIgPSBhc3luYyAoZXZlbnQ6IGFueSkgPT4ge1xuICBjb25zdCByZXF1ZXN0ZWRDbG91ZEV2ZW50ID0gbmV3IENsb3VkRXZlbnQ8YW55PihldmVudClcbiAgY29uc3QgZGV2aWNlSUQgPSByZXF1ZXN0ZWRDbG91ZEV2ZW50LmRhdGEuZGV2aWNlX2lkXG4gIGNvbnNvbGUubG9nKGRldmljZUlEKVxuXG4gIGxldCB0aGluZ1Jlc3VsdDogdW5kZWZpbmVkIHwgYW55XG4gIGNvbnN0IHNlYXJjaFJlc3BvbnNlID0gYXdhaXQgaW90LnNlYXJjaEluZGV4KHsgcXVlcnlTdHJpbmc6IGB0aGluZ05hbWU6JHtkZXZpY2VJRH1gIH0pLnByb21pc2UoKVxuICBpZiAoc2VhcmNoUmVzcG9uc2UudGhpbmdzIS5sZW5ndGggPT09IDApIHtcbiAgICBjb25zb2xlLmxvZyhcIk5vIHJlc3VsdHMgd2l0aCBkZXZpY2UgSUQgKFwiICsgZGV2aWNlSUQgKyBcIilcIilcbiAgICB0aGluZ1Jlc3VsdCA9IHtcbiAgICAgIGRldmljZV9pZDogZGV2aWNlSUQsXG4gICAgICBzdGF0dXM6IDQwNFxuICAgIH1cbiAgfSBlbHNlIGlmIChzZWFyY2hSZXNwb25zZS50aGluZ3MhLmxlbmd0aCAhPT0gMSkge1xuICAgIGNvbnNvbGUubG9nKFwiSW5jb25zaXN0ZW50IHRoaW5nIHJlcG86IE1vcmUgdGhhbiBvbmUgcmVzdWx0IGZvciB0aGUgc2FtZSBkZXZpY2UgSUQgKFwiICsgZGV2aWNlSUQgKyBcIilcIilcbiAgICB0aGluZ1Jlc3VsdCA9IHtcbiAgICAgIGRldmljZV9pZDogZGV2aWNlSUQsXG4gICAgICBzdGF0dXM6IDQwOVxuICAgIH1cbiAgfSBlbHNlIHtcbiAgICBjb25zdCB0aGluZyA9IHNlYXJjaFJlc3BvbnNlLnRoaW5ncyFbMF1cbiAgICBjb25zdCBwcmluY2lwYWxSZXNwb25zZSA9IGF3YWl0IGlvdC5saXN0VGhpbmdQcmluY2lwYWxzKHsgdGhpbmdOYW1lOiB0aGluZy50aGluZ05hbWUhLCBtYXhSZXN1bHRzOiAyNSB9KS5wcm9taXNlKClcbiAgICBjb25zdCBjZXJ0aWZpY2F0ZXMgPSBbXVxuICAgIGNvbnN0IHByaW5jaXBhbHMgPSBwcmluY2lwYWxSZXNwb25zZS5wcmluY2lwYWxzXG4gICAgZm9yIChjb25zdCBwcmluY2lwYWwgb2YgcHJpbmNpcGFscyEpIHtcbiAgICAgIGNvbnN0IHNwbGl0aWVkUHJpbmNpcGFsID0gcHJpbmNpcGFsLnNwbGl0KFwiOlwiKVxuICAgICAgY29uc3QgY2VydGlmaWNhdGVJRCA9IHNwbGl0aWVkUHJpbmNpcGFsW3NwbGl0aWVkUHJpbmNpcGFsLmxlbmd0aCAtIDFdLnJlcGxhY2UoXCJjZXJ0L1wiLCBcIlwiKVxuICAgICAgY29uc3QgY2VydGlmaWNhdGVSZXNwb25zZSA9IGF3YWl0IGlvdC5kZXNjcmliZUNlcnRpZmljYXRlKHsgY2VydGlmaWNhdGVJZDogY2VydGlmaWNhdGVJRCB9KS5wcm9taXNlKClcbiAgICAgIGNvbnN0IGNlcnRQZW0gPSBjZXJ0aWZpY2F0ZVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNlcnRpZmljYXRlUGVtIVxuXG4gICAgICBjb25zdCBjZXJ0ID0gQ2VydGlmaWNhdGUuZnJvbVBFTShCdWZmZXIuZnJvbShjZXJ0UGVtLCBcInV0ZjhcIikpXG5cbiAgICAgIGNlcnRpZmljYXRlcy5wdXNoKHtcbiAgICAgICAgc2VyaWFsX251bWJlcjogY2h1bmsoY2VydC5zZXJpYWxOdW1iZXIsIDIpLmpvaW4oXCItXCIpLFxuICAgICAgICBzdGF0dXM6IGNlcnRpZmljYXRlUmVzcG9uc2UuY2VydGlmaWNhdGVEZXNjcmlwdGlvbiEuc3RhdHVzLFxuICAgICAgICBhcm46IGNlcnRpZmljYXRlUmVzcG9uc2UuY2VydGlmaWNhdGVEZXNjcmlwdGlvbiEuY2VydGlmaWNhdGVBcm4sXG4gICAgICAgIGlkOiBjZXJ0aWZpY2F0ZVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmNlcnRpZmljYXRlSWQsXG4gICAgICAgIHVwZGF0ZV9kYXRlOiBjZXJ0aWZpY2F0ZVJlc3BvbnNlLmNlcnRpZmljYXRlRGVzY3JpcHRpb24hLmxhc3RNb2RpZmllZERhdGUsXG4gICAgICAgIGNhX25hbWU6IGNlcnQuaXNzdWVyLmNvbW1vbk5hbWVcbiAgICAgIH0pXG4gICAgfVxuXG4gICAgdGhpbmdSZXN1bHQgPSB7XG4gICAgICBkZXZpY2VfaWQ6IHRoaW5nLnRoaW5nTmFtZSxcbiAgICAgIHN0YXR1czogMjAwLFxuICAgICAgY29uZmlnOiB7XG4gICAgICAgIGF3c19pZDogdGhpbmcudGhpbmdJZCxcbiAgICAgICAgbGFzdF9jb25uZWN0aW9uOiB0aGluZy5jb25uZWN0aXZpdHk/LnRpbWVzdGFtcCxcbiAgICAgICAgY2VydGlmaWNhdGVzOiBjZXJ0aWZpY2F0ZXNcbiAgICAgIH1cbiAgICB9XG4gIH1cblxuICBjb25zdCBjbG91ZEV2ZW50ID0gbmV3IENsb3VkRXZlbnQoe1xuICAgIHR5cGU6IFwiaW8ubGFtYXNzdS5pb3Rjb3JlLnRoaW5nLmNvbmZpZy5yZXNwb25zZVwiLFxuICAgIGlkOiByZXF1ZXN0ZWRDbG91ZEV2ZW50LmlkLFxuICAgIHNvdXJjZTogXCJhd3MvbGFtYmRhXCIsXG4gICAgdGltZTogbmV3IERhdGUoKS50b1N0cmluZygpLFxuICAgIHNwZWN2ZXJzaW9uOiBcIjEuMFwiLFxuICAgIGRhdGE6IHRoaW5nUmVzdWx0XG4gIH0pXG4gIHRyeSB7XG4gICAgY29uc29sZS5sb2coXCJDbG91ZCBFdmVudFwiLCBjbG91ZEV2ZW50KVxuICAgIGNvbnN0IHNxc1Jlc3BvbnNlID0gYXdhaXQgc3FzLnNlbmRNZXNzYWdlKHsgUXVldWVVcmw6IHByb2Nlc3MuZW52LlNRU19SRVNQT05TRV9RVUVVRV9VUkwhLCBNZXNzYWdlQm9keTogY2xvdWRFdmVudC50b1N0cmluZygpIH0pLnByb21pc2UoKVxuICAgIGNvbnNvbGUubG9nKHNxc1Jlc3BvbnNlKVxuICB9IGNhdGNoIChlcnIpIHtcbiAgICBjb25zb2xlLmxvZyhcImVycm9yIHdoaWxlIHNlbmRpbmcgU1FTIG1lc3NnYWVcIiwgZXJyKVxuICB9XG59XG5cbmZ1bmN0aW9uIGNodW5rIChzdHI6IHN0cmluZywgbjogbnVtYmVyKSB7XG4gIGNvbnN0IHJldCA9IFtdXG4gIGxldCBpXG4gIGxldCBsZW5cblxuICBmb3IgKGkgPSAwLCBsZW4gPSBzdHIubGVuZ3RoOyBpIDwgbGVuOyBpICs9IG4pIHtcbiAgICByZXQucHVzaChzdHIuc3Vic3RyKGksIG4pKVxuICB9XG5cbiAgcmV0dXJuIHJldFxufVxuIl19