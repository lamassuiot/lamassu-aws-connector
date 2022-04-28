# Cloud Events

![Cloud events logo](docs/img/cloud-events_logo.png)

[< :house:](.)
### Table of contents
* [Events](#Events)
* [References](#References)

## Events

| **Event Type**                                                          | **Source**                            | **Description** |
|-------------------------------------------------------------------------|---------------------------------------|-----------------|
| io.lamassu.ca.create                                                    | lamassu/ca                            |                 |
| io.lamassu.ca.import                                                    | lamassu/ca                            |                 |
| io.lamassu.ca.update                                                    | lamassu/ca                            |                 |
| io.lamassu.cert.update                                                  | lamassu/ca                            |                 |
| [io.lamassu.iotcore.config.request](#io.lamassu.iotcore.config.request) | lamassu/aws-connector/${connector-id} |                 |
| io.lamassu.iotcore.config.response                                      | aws/lambda                            |                 |
| io.lamassu.iotcore.ca.registration.request-code                         | lamassu/aws-connector/${connector-id} |                 |
| io.lamassu.iotcore.ca.registration.response-code                        | aws/lambda                            |                 |
| io.lamassu.iotcore.ca.registration.signed-code                          | lamassu/aws-connector/${connector-id} |                 |
| io.lamassu.iotcore.ca.policy.attach                                     | lamassu/aws-connector/${connector-id} |                 |
| io.lamassu.iotcore.cert.update-status                                   | aws/cloud-trail                       |                 |
| io.lamassu.iotcore.thing.config.request                                 | aws/lambda                            |                 |

### io.lamassu.ca.create

```json 
{
    "specversion":"1.0",
    "id":"40492df26ca4d1e6:40492df26ca4d1e6:0000000000000000:1",
    "source":"lamassu/ca",
    "type":"io.lamassu.ca.create",
	"datacontenttype":"application/json",
	"time":"2022-03-31T06:52:13.207672275Z",
    "data":{
        "name":"CA1",
		"serial_number":"23-f9-3e-b0-9c-0e-f7-6b-1d-d9-0f-a2-22-47-93-45-23-53-f1-03",
		"cert":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUQyVENDQXNHZ0F3SUJBZ0lVSS9rK3NKd085MnNkMlEraUlrZVRSU05UOFFNd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1R6RUxNQWtHQTFVRUJoTUNSVk14RVRBUEJnTlZCQWdUQ0VkSlVGVmFTMDlCTVJFd0R3WURWUVFIRXdoQgpVbEpCVTBGVVJURU1NQW9HQTFVRUNoTURTVXRNTVF3d0NnWURWUVFERXdORFFURXdIaGNOTWpJd016TXhNRFkxCk1UUXpXaGNOTWpJd05qSXlNVFExTWpFeVdqQlBNUXN3Q1FZRFZRUUdFd0pGVXpFUk1BOEdBMVVFQ0JNSVIwbFEKVlZwTFQwRXhFVEFQQmdOVkJBY1RDRUZTVWtGVFFWUkZNUXd3Q2dZRFZRUUtFd05KUzB3eEREQUtCZ05WQkFNVApBME5CTVRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTGkvRGthM3FaMTkwS1BoClp0YUV3cXkySFRaK1VUVkxYaWt6N0Vwcit0VHc2OHBDQUJTUEZXQW43Rm1STDU0Mjg0dHl1QjdxZlFKdjRFamIKb041U3FJbGpRS2NSMU1pY3czUlBKWWlhSDdkRDJNZXk5NEQ3SlhCYkNvQXBUaUxIcUMvbEFHT2g2dFZqNTFwaApXYlhWMVBsTzUxT3l4S2FIM1dtWmFSL0FubTdXMnNCbFVvdGRrZ1dVVmZCRG0vSEtOUHZQeDU1MkJpOHl1TjloCnk2Vzg3Y252OUNWYm9YcmJFSFBmUENBb09Fa01TUjFJU0NrRFVncXcwN1BqRUEzSTYxWmN2SWFDYkY3d2p1RkEKMUtueVRBWUR5UHlMbHJrOEJzcElCamF6Mm1BbTNTeld2N1FWRU1hWlVaUVZOU3NScEVoSGpQY0JRWko5Snp6bwppcS85aUwwQ0F3RUFBYU9CckRDQnFUQU9CZ05WSFE4QkFmOEVCQU1DQVFZd0R3WURWUjBUQVFIL0JBVXdBd0VCCi96QWRCZ05WSFE0RUZnUVV1Y01KUVROSG5Od2IvZEdPZ2hUTFVscEMzZFF3SHdZRFZSMGpCQmd3Rm9BVXVjTUoKUVROSG5Od2IvZEdPZ2hUTFVscEMzZFF3TmdZSUt3WUJCUVVIQVFFRUtqQW9NQ1lHQ0NzR0FRVUZCekFCaGhwbwpkSFJ3T2k4dlpHVjJMbXhoYldGemMzVXVhVzg2T1RBNU9EQU9CZ05WSFJFRUJ6QUZnZ05EUVRFd0RRWUpLb1pJCmh2Y05BUUVMQlFBRGdnRUJBSmszQUE3M3h0aU5QTzl1dTIvTDhGQUxNQXpaTzFKSmd0Tkxja2RBYzhXQnRLd0MKZHZtT1V1Q251R2tnQkhOZkJxMWRsTlNqaGVSL2g2a1dLYmdiK1hBczZRbTJGMUJURVA5Sitpc1dyaGgyQU5GYworVWxtT0FST2dSTkdGMEkyeFpaOUpEZFA0MUY4RXc0ODJyT3dpZFl2Q2tWa1p4RUFjRVA3Y2QyWmZpWnhhL0d6CjhPQ3psSUhVZTBTKzZJalVleC9rc2pzcDRrclFpMVZRYnFTT040ZTFLTVZyWUtLTzZHR3orNW0zUWtSZGVXUDQKM28rN2J5UDdtNEkvdm9wTUhzaWl0OWtwOFBMWWg4Q0RoelV2MzkyR2EzdXl6SDV4ZWUvdFhRT3ppdTlWRTV0agpPZHgyTHVtTUJoMzFuT09GUHorL3dCM3VIdEdvRHlKWXJDVGNYREE9Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0="
		
    }    
}
```

### io.lamassu.ca.import

```json 
{
    "specversion":"1.0",
    "id":"57cc9014a8021b01:57cc9014a8021b01:0000000000000000:1",
    "source":"lamassu/ca",
    "type":"io.lamassu.ca.import",
	"datacontenttype":"application/json",
	"time":"2022-03-31T07:06:34.265731829Z",
    "data":{
        "name":"CA2",
		"serial_number":"04-22-b1-ac-5a-3c-5d-f7-f8-8c-a5-cf-cb-e5-4a-d2-f5-df-f3-35",
		"cert":"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURtekNDQW9PZ0F3SUJBZ0lVQkNLeHJGbzhYZmY0aktYUHkrVkswdlhmOHpVd0RRWUpLb1pJaHZjTkFRRUwKQlFBd1hURUxNQWtHQTFVRUJoTUNSVk14RVRBUEJnTlZCQWdNQ0VkSlVGVmFTMDlCTVJFd0R3WURWUVFIREFoQgpVbEpCVTBGVVJURU1NQW9HQTFVRUNnd0RTVXRNTVF3d0NnWURWUVFMREFOYVVFUXhEREFLQmdOVkJBTU1BME5CCk1qQWVGdzB5TWpBek16RXdOalU1TlRWYUZ3MHlNekF6TXpFd05qVTVOVFZhTUYweEN6QUpCZ05WQkFZVEFrVlQKTVJFd0R3WURWUVFJREFoSFNWQlZXa3RQUVRFUk1BOEdBMVVFQnd3SVFWSlNRVk5CVkVVeEREQUtCZ05WQkFvTQpBMGxMVERFTU1Bb0dBMVVFQ3d3RFdsQkVNUXd3Q2dZRFZRUUREQU5EUVRJd2dnRWlNQTBHQ1NxR1NJYjNEUUVCCkFRVUFBNElCRHdBd2dnRUtBb0lCQVFETWFUOWo0bHB4RzdHVjRKdkRyMnFGWXNPUkpVN083cWUxd2k2SzNCZmIKZk4ySWw1dDdGK2tSc25BWnFnM3R3T25Hb1Q2czZVZUhZc1hxTDJzMURlbUZzM3dtRGFhZUJJVldzRlZUKzV1WQpYWWZaZVVBdC9FRStWMWRXaFdwcUtHYk9LeWhFcVQ1Rm1jU2psTitQeUpkV1RVSEFZcUtNcTZPSnJmdFFIMUtJCnFoazZseVFRTzNla3RwZXBCWUxjSU8yazg3Rm5WZ0o5RWFlektmUG1EVDRXaitMOUNmb1pZTnFSOTNjcFM1ZzEKMWRYenl5ZUEvV3Zmbk5vdm5rbDdFQk9oNHI0NFhVOFVWbWkrUWtqQWdyTklpRlRIUHp4OGtpS0dvVkZDQ1A3OQo4Y0NzM1JxTVJhYk9BdENOemFWRzdnNXNlRVJLOXVOMk1vY1FuWkdKQzF4REFnTUJBQUdqVXpCUk1CMEdBMVVkCkRnUVdCQlJacFhvRThBSE9HL0tMVkhDU3hmZ2xsOEl3TVRBZkJnTlZIU01FR0RBV2dCUlpwWG9FOEFIT0cvS0wKVkhDU3hmZ2xsOEl3TVRBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFCVAp4akQzcjd6L3dQVHBFbUdnN25xd3hoRXB3cHJjMDR1VWxFUjRFdUl4UlVWSjh0TDNnRzdrbVBhK0RsbHpVdkl1ClhNWlpFMUJwck9QbTYzb3V4cThJb0dXZW5DYWIzUU95V1o2YTBPcllXNmJYcnROdzJuTFdKd2UvbGM1a0ViTnYKNjdNVmM4emovcFY5TXFuTmxsMWc2THN3V0kreFBqaXNqSFl5MnlMVGlRcHk2WC8yK3B1VDZHRHBIK3pCenRUagpqVTh3Um1WZWloeHFZNTBrVGcreGRmaUE0OWhMM3VkcHRmU0VXd1BSQW9zUHVTT1NTTndHM1ZGQjAvcjRTWWROCjE1TWM3SEpuMlR3OEttbkRYMTVzL3hYMWROaW51RmVZMklCUzQzZkZmT1Viczd5SXc1eFF3a0JjMm5PQVlJdDgKSUhtQm1MRVArWktmNms5Q1JIc2kKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQ=="	
    }    
}
```

### io.lamassu.ca.update

```json 
{
    "specversion":"1.0",
    "id":"208e73dd2a283350:208e73dd2a283350:0000000000000000:1",
    "source":"lamassu/ca",
    "type":"io.lamassu.ca.update",
	"datacontenttype":"application/json",
	"time":"2022-03-31T07:10:55.371676082Z",
    "data":{
        "name":"CA2",
        "status": "REVOKED"
    }    
}
```

### io.lamassu.cert.update

```json 
{
    "specversion":"1.0",
    "id":"0814bee304159b19:0814bee304159b19:0000000000000000:1",
    "source":"lamassu/ca",
    "type":"io.lamassu.cert.update",
	"datacontenttype":"application/json",
	"time":"2022-03-31T07:17:36.907957488Z",
    "data":{
        "name":"CA2",
        "status": "REVOKED",
        "serial_number":"23-f9-3e-b0-9c-0e-f7-6b-1d-d9-0f-a2-22-47-93-45-23-53-f1-03"
    }    
}
```

### io.lamassu.iotcore.thing.config.response

```json 
{
    "specversion":"1.0",
    "id":"8308f9c3-4f08-44a3-9b1d-49818447b6b1",
    "source":"aws/lambda",
    "type":"io.lamassu.iotcore.thing.config.response",
    "time":"2022-04-04T11:33:36Z",
    "data":[
        {
            "aws_id":"b92c131c-7017-4f54-971c-2b12138a46bf",
            "device_id":"796786f3-eea5-4cd8-bf9e-9aae738b4176",
            "last_connection":null,
            "certificates":[
                {
                    "status":"ACTIVE",
                    "arn":"arn:aws:iot:eu-west-1:345876576284:cert/3d5621f7822da199cfa20dfb89f36e7d4d7cba915a61952202fb57428dd0a48c",
                    "id":"3d5621f7822da199cfa20dfb89f36e7d4d7cba915a61952202fb57428dd0a48c",
                    "update_date":"2022-03-31T11:58:28.405Z"
                }
            ]
        },
        {
            "aws_id":"e324c233-54de-4783-9bff-528b02eae499",
            "device_id":"51809c2a-0429-4719-9ede-8a0336c07532",
            "last_connection":null,
            "certificates":[
                {
                    "status":"REVOKED",
                    "arn":"arn:aws:iot:eu-west-1:345876576284:cert/b749a584b26e74662ea529bb58caadec96e822e6d436a7237a0ff5805e971563",
                    "id":"b749a584b26e74662ea529bb58caadec96e822e6d436a7237a0ff5805e971563",
                    "update_date":"2022-03-29T16:37:18.527Z"
                }
            ]
        }
    ]
}
```


## References

* Cloud Events web page: [Cloud events](https://cloudevents.io/)
* Cloud Events github: [Cloud events GITHUB](https://github.com/cloudevents)
* Clud Events SDK for Go: [Cloud Events SDK for GO](https://github.com/cloudevents/sdk-go)
