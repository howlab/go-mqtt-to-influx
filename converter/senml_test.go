package converter

import (
	"github.com/golang/mock/gomock"
	"github.com/koestler/go-mqtt-to-influx/converter/mock"
	"testing"
	"time"
)

func TestSenML(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockConfig := converter_mock.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Name().Return("test-converter").AnyTimes()
	mockConfig.EXPECT().TargetMeasurement().Return("floatValue").MinTimes(1)

	stimuli := TestStimuliResponse{
		{
			Topic: "piegn/tele/senml/24v-bmv/meas",
			Payload: `[
	{"bn":"urn:dev:ow:10e2073a0108006/","bt":1.276020076001e+09,
	"bu":"A",
	"n":"voltage","u":"V","v":120.1},
   {"n":"current","v":1.7}
 ]`,
			ExpectedLines: []string{
				"meas,ow=10e2073a0108006,unit=V voltage=120.1 1276020076000000000",
				"meas,ow=10e2073a0108006,unit=A current=1.7 1276020076000000000",
				},
			ExpectedTimeStamp: time.Date(2010, time.June, 6, 18, 1, 16, 0, time.UTC),
		},
	}

	if h, err := GetHandler("senml"); err != nil {
		t.Errorf("did not expect an error while getting handler: %s", err)
	} else {
		testStimuliResponse(t, mockCtrl, mockConfig, h, stimuli)
	}
}
