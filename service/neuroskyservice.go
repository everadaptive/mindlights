package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	dsp "github.com/eripe970/go-dsp-utils"
	"github.com/everadaptive/mindlights/eeg/neurosky"
	"github.com/vmware/transport-go/model"
	"github.com/vmware/transport-go/service"
)

const (
	NeuroskyServiceChan = "neurosky-service"

	NeuroskySignalChan     = "signal-quality"
	NeuroskyAttentionChan  = "attention"
	NeuroskyMeditationChan = "meditation"
	NeuroskyEEGPowerChan   = "eeg-power"
	NeuroskyEEGRawChan     = "eeg-raw"
)

// NeuroskyService is a very simple service to demonstrate how request-response cycles are handled in Transport & Plank.
// this service has two requests named "ping-post" and "ping-get", the first accepts the payload and expects it to be of
// a POJO type (e.g. {"anything": "here"}), whereas the second expects the payload to be a pure string.
// a request made through the Event Bus API like bus.RequestOnce() will be routed to HandleServiceRequest()
// which will match the request's Request to the list of available service request types and return the response.
type NeuroskyService struct {
	events    chan neurosky.MindflexEvent
	rawValues []float64
}

func NewNeuroskyService(events chan neurosky.MindflexEvent) *NeuroskyService {
	return &NeuroskyService{
		events:    events,
		rawValues: make([]float64, 2048),
	}
}

// Init will fire when the service is being registered by the fabric, it passes a reference of the same core
// Passed through when implementing HandleServiceRequest
func (ps *NeuroskyService) Init(core service.FabricServiceCore) error {
	core.Bus().GetChannelManager().CreateChannel(fmt.Sprintf("%s/%s", "HEADSET-03", NeuroskySignalChan)).SetGalactic(fmt.Sprintf("/topic/%s/%s", "HEADSET-03", NeuroskySignalChan))
	core.Bus().GetChannelManager().CreateChannel(fmt.Sprintf("%s/%s", "HEADSET-03", NeuroskyAttentionChan)).SetGalactic(fmt.Sprintf("/topic/%s/%s", "HEADSET-03", NeuroskyAttentionChan))
	core.Bus().GetChannelManager().CreateChannel(fmt.Sprintf("%s/%s", "HEADSET-03", NeuroskyMeditationChan)).SetGalactic(fmt.Sprintf("/topic/%s/%s", "HEADSET-03", NeuroskyMeditationChan))
	core.Bus().GetChannelManager().CreateChannel(fmt.Sprintf("%s/%s", "HEADSET-03", NeuroskyEEGPowerChan)).SetGalactic(fmt.Sprintf("/topic/%s/%s", "HEADSET-03", NeuroskyEEGPowerChan))
	core.Bus().GetChannelManager().CreateChannel(fmt.Sprintf("%s/%s", "HEADSET-03", NeuroskyEEGRawChan)).SetGalactic(fmt.Sprintf("/topic/%s/%s", "HEADSET-03", NeuroskyEEGRawChan))

	go func() {
		count := 0
		for v := range ps.events {
			switch v.Type {
			case neurosky.POOR_SIGNAL:
				core.Bus().SendResponseMessage(fmt.Sprintf("%s/%s", v.Source, NeuroskySignalChan), v, nil)
			case neurosky.ATTENTION:
				core.Bus().SendResponseMessage(fmt.Sprintf("%s/%s", v.Source, NeuroskyAttentionChan), v, nil)
			case neurosky.MEDITATION:
				core.Bus().SendResponseMessage(fmt.Sprintf("%s/%s", v.Source, NeuroskyMeditationChan), v, nil)
			case neurosky.EEG_POWER:
				core.Bus().SendResponseMessage(fmt.Sprintf("%s/%s", v.Source, NeuroskyEEGPowerChan), v, nil)
			case neurosky.EEG_RAW:
				ps.rawValues = append(ps.rawValues[1:], float64(v.EEGRawPower))
				if count%100 == 0 {
					count = 0
					s := dsp.Signal{
						SampleRate: 512,
						Signal:     ps.rawValues,
					}
					n, _ := s.Normalize()
					filt, _ := n.LowPassFilter(110)

					fs, _ := filt.FrequencySpectrum()
					f := []neurosky.ComplexValue{}
					for k := range fs.Frequencies {
						if fs.Frequencies[k] < 110 {
							f = append(f, neurosky.ComplexValue{Real: fs.Spectrum[k], Imaginary: fs.Frequencies[k]})
						}
					}
					v.EEGRawPowerFFT = f
					core.Bus().SendResponseMessage(fmt.Sprintf("%s/%s", v.Source, NeuroskyEEGRawChan), v, nil)
				}
				count++
			}
		}
	}()
	return nil
}

// HandleServiceRequest routes the incoming request and based on the Request property of request, it invokes the
// appropriate handler logic defined and separated by a switch statement like the one shown below.
func (ps *NeuroskyService) HandleServiceRequest(request *model.Request, core service.FabricServiceCore) {
	switch request.Request {
	// ping-post request type accepts the payload as a POJO
	case "ping-post":
		m := make(map[string]interface{})
		m["timestamp"] = time.Now().Unix()
		err := json.Unmarshal(request.Payload.([]byte), &m)
		if err != nil {
			core.SendErrorResponse(request, 400, err.Error())
		} else {
			core.SendResponse(request, m)
		}
	// ping-get request type accepts the payload as a string
	case "ping-get":
		rsp := make(map[string]interface{})
		val := request.Payload.(string)
		rsp["payload"] = val + "-response"
		rsp["timestamp"] = time.Now().Unix()
		core.SendResponse(request, rsp)
	default:
		core.HandleUnknownRequest(request)
	}
}

// OnServiceReady contains logic that handles the service initialization that needs to be carried out
// before it is ready to accept user requests. Plank monitors and waits for service initialization to
// complete by trying to receive a boolean payload from a channel of boolean type. as a service developer
// you need to perform any and every init logic here and return a channel that would receive a payload
// once your service truly becomes ready to accept requests.
func (ps *NeuroskyService) OnServiceReady() chan bool {
	// for sample purposes this service initializes instantly
	readyChan := make(chan bool, 1)
	readyChan <- true
	return readyChan
}

// OnServerShutdown is the opposite of OnServiceReady. it is called when the server enters graceful shutdown
// where all the running services need to complete before the server could shut down finally. this method does not need
// to return anything because the main server thread is going to shut down soon, but if there's any important teardown
// or cleanup that needs to be done, this is the right place to perform that.
func (ps *NeuroskyService) OnServerShutdown() {
	// for sample purposes emulate a 1 second teardown process
	time.Sleep(1 * time.Second)
}

// GetRESTBridgeConfig returns a list of REST bridge configurations that Plank will use to automatically register
// REST endpoints that map to the requests for this service. this means you can map any request types defined under
// HandleServiceRequest with any combination of URI, HTTP verb, path parameter, query parameter and request headers.
// as the service author you have full control over every aspect of the translation process which basically turns
// an incoming *http.Request into model.Request. See FabricRequestBuilder below to see it in action.
func (ps *NeuroskyService) GetRESTBridgeConfig() []*service.RESTBridgeConfig {
	return []*service.RESTBridgeConfig{
		{
			ServiceChannel: NeuroskyServiceChan,
			Uri:            "/rest/signalQuality",
			Method:         http.MethodPost,
			AllowHead:      true,
			AllowOptions:   true,
			FabricRequestBuilder: func(w http.ResponseWriter, r *http.Request) model.Request {
				body, _ := ioutil.ReadAll(r.Body)
				return model.CreateServiceRequest("signal-quality", body)
			},
		},
	}
}
