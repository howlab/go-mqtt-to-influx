package converter

import (
	"log"
	"time"
	"strings"
	"regexp"
	"github.com/mainflux/senml"
	"github.com/voicera/gooseberry/urn"

)

type senmlOutputMessage struct {
	timestamp		    time.Time
	measurement		string
	deviceTagIdentifier	string
	deviceTagValue		string
	field			string
	unit			string
	value			float64
}

// example input: piegn/tele/senml/24v-bmv/meas
// -> use meas as measurement
var senmlTopicMatcher = regexp.MustCompile("^(.*)/([^/]*)$")


func init() {
	registerHandler("senml", senmlHandler)
}

// parses messages generated by senml producers in the SenML format
// and write one point per value to the influxdb
// example input:
// [
// 	{"bn":"urn:dev:ow:10e2073a0108006/,"bt":1.276020076001e+09,
// 	"bu":"A",
// 	"n":"voltage","u":"V","v":120.1},
//    {"n":"current","t":-5,"v":1.2},
//    {"n":"current","t":-4,"v":1.3},
//    {"n":"current","t":-3,"v":1.4},
//    {"n":"current","t":-2,"v":1.5},
//    {"n":"current","t":-1,"v":1.6},
//    {"n":"current","v":1.7}
//  ]
func senmlHandler(c Config, input Input, outputFunc OutputFunc) {
	// parse topic
	matches := senmlTopicMatcher.FindStringSubmatch(input.Topic())
	if len(matches) < 3 {
		log.Printf("senml[%s]: cannot extract device from topic='%s", c.Name(), input.Topic())
		return
	}
	measurement := matches[2]

	// parse payload - Decode senml 
	pack, err := senml.Decode(input.Payload(), senml.JSON)
	if err != nil {
		log.Printf("senml[%s]: cannot senml decode: %s", c.Name(), err)
		return
	}

	// Normalize into resolved records
	message, err := senml.Normalize(pack)

	if err != nil {
		log.Printf("senml[%s]: cannot senml normalize: %s", c.Name(), err)
		return
	}
	baseTime := time.Now()
	for _, record := range message.Records {

		deviceURN, success := urn.TryParseString(record.Name)

		if success != true {
			log.Printf("senml[%s]: Error parsing URN string: %s", c.Name(), record.Name)
			return
		}

		var deviceURNParts []string

		deviceURNParts = strings.SplitN(deviceURN.GetNamespaceSpecificString(),"/",2)

		if len(deviceURNParts) < 2 {
			log.Printf("senml[%s]: '/' not found for field: %s", c.Name(), deviceURN.GetNamespaceSpecificString())
			continue
		}

		var deviceBody []string

		if ( deviceURN.GetNamespaceID() == "dev" ) {
			deviceBody= strings.Split(deviceURNParts[0],":")
			if len(deviceBody) < 2 {
				log.Printf("senml[%s]: DEV URN subtype not found: %s", c.Name(), deviceURN)
				return
			}
		} else {
			log.Printf("senml[%s]: URN namespace not supported: %s. Supported namespaces: [dev].", "", deviceURN)
		}

		deviceTagIdentifier := deviceBody[0]
		deviceTagValue := deviceBody[1]
		field := deviceURNParts[1]

		var timestamp time.Time
		if record.Time < 1<<28 {
			timestamp = baseTime.Add(time.Duration(record.Time)*time.Millisecond)
		} else {
			timestamp = time.Unix(int64(record.Time),int64(0))
		}

		outputFunc(senmlOutputMessage{
			timestamp:		timestamp,
			measurement:		measurement,
			deviceTagIdentifier:	deviceTagIdentifier,
			deviceTagValue:		deviceTagValue,
			field:			field,
			unit:			record.Unit,
			value:			*record.Value,
		})
	}
}

func (m senmlOutputMessage) Measurement() string {
	return m.measurement
}

func (m senmlOutputMessage) Tags() map[string]string {
	return map[string]string{
		"unit":   m.unit,
		m.deviceTagIdentifier: m.deviceTagValue,
	}
}

func (m senmlOutputMessage) Fields() map[string]interface{} {
	return map[string]interface{}{
		m.field: m.value,
	}
}

func (m senmlOutputMessage) Time() time.Time {
	return m.timestamp
}
