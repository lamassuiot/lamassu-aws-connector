package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	api "github.com/lamassuiot/aws-connector/pkg/common"
	cloudApi "github.com/lamassuiot/lamassuiot/pkg/cloud-provider/common/api"
	log "github.com/sirupsen/logrus"
)

type Middleware func(Service) Service

func LoggingMiddleware() Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next: next,
		}
	}
}

type loggingMiddleware struct {
	next Service
}

func (mw loggingMiddleware) Health() (output bool) {
	defer func(begin time.Time) {
		log.WithFields(log.Fields{
			"method": "Health",
			"took":   time.Since(begin),
		}).Trace(output)
	}(time.Now())
	return mw.next.Health()
}

func (mw loggingMiddleware) RegisterCA(ctx context.Context, input *cloudApi.RegisterCAInput) (output *cloudApi.RegisterCAOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "RegisterCA"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.RegisterCA(ctx, input)
}

func (mw loggingMiddleware) UpdateConfiguration(ctx context.Context, input *cloudApi.UpdateConfigurationInput) (output *cloudApi.UpdateConfigurationOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "UpdateConfiguration"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.UpdateConfiguration(ctx, input)
}

func (mw loggingMiddleware) GetConfiguration(ctx context.Context, input *cloudApi.GetConfigurationInput) (output *cloudApi.GetConfigurationOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "GetConfiguration"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.GetConfiguration(ctx, input)
}

func (mw loggingMiddleware) GetDeviceConfiguration(ctx context.Context, input *cloudApi.GetDeviceConfigurationInput) (output *cloudApi.GetDeviceConfigurationOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "GetDeviceConfiguration"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.GetDeviceConfiguration(ctx, input)
}

func (mw loggingMiddleware) UpdateCAStatus(ctx context.Context, input *cloudApi.UpdateCAStatusInput) (output *cloudApi.UpdateCAStatusOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "UpdateCAStatus"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.UpdateCAStatus(ctx, input)
}

func (mw loggingMiddleware) UpdateDMSCaCerts(ctx context.Context, input *cloudApi.UpdateDMSCaCertsInput) (output *cloudApi.UpdateDMSCaCertsOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "UpdateDMSCaCerts"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.UpdateDMSCaCerts(ctx, input)
}

func (mw loggingMiddleware) UpdateDeviceCertificateStatus(ctx context.Context, input *cloudApi.UpdateDeviceCertificateStatusInput) (output *cloudApi.UpdateDeviceCertificateStatusOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "UpdateDeviceCertificateStatus"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.UpdateDeviceCertificateStatus(ctx, input)
}

func (mw loggingMiddleware) UpdateDeviceDigitalTwinReenrollmentStatus(ctx context.Context, input *cloudApi.UpdateDeviceDigitalTwinReenrollmentStatusInput) (output *cloudApi.UpdateDeviceDigitalTwinReenrollmentStatusOutput, err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "UpdateDeviceDigitalTwinReenrollmentStatus"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.UpdateDeviceDigitalTwinReenrollmentStatus(ctx, input)
}

func (mw loggingMiddleware) HandleUpdateCertificateStatus(ctx context.Context, input *api.HandleUpdateCertificateStatusInput) (err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "HandleUpdateCertificateStatus"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace()
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.HandleUpdateCertificateStatus(ctx, input)
}
func (mw loggingMiddleware) HandleUpdateCAStatus(ctx context.Context, input *api.HandleUpdateCAStatusInput) (err error) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "HandleUpdateCAStatus"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = input

		if err == nil {
			log.WithFields(logMsg).Trace()
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.HandleUpdateCAStatus(ctx, input)
}

func (mw loggingMiddleware) HandleCloudEvents(ctx context.Context, event cloudevents.Event) (err error) {
	defer func(begin time.Time) {
		msg, _ := json.Marshal(event)
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "HandleCloudEvents"
		logMsg["took"] = time.Since(begin)
		logMsg["input"] = string(msg)

		if err == nil {
			log.WithFields(logMsg).Trace()
		} else {
			log.WithFields(logMsg).Error(err)
		}
	}(time.Now())
	return mw.next.HandleCloudEvents(ctx, event)
}
func (mw loggingMiddleware) GetAccountID() (output string) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "GetAccountID"
		logMsg["took"] = time.Since(begin)
		log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
	}(time.Now())
	return mw.next.GetAccountID()
}
func (mw loggingMiddleware) GetDefaultRegion() (output string) {
	defer func(begin time.Time) {
		var logMsg = map[string]interface{}{}
		logMsg["method"] = "GetDefaultRegion"
		logMsg["took"] = time.Since(begin)
		log.WithFields(logMsg).Trace(fmt.Sprintf("output: %v", output))
	}(time.Now())
	return mw.next.GetDefaultRegion()
}
