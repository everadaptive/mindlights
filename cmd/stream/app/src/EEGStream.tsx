import { BaseComponent } from "./BaseComponent";
import {
  EEGStreamState,
  EEGStreamProps,
  EEGStreamChannelProps,
  ListenButtonProps,
  EEGEvent
} from "./EEGStreamModel";
import { APIResponse } from "@vmw/transport";
import { MessageHandler } from "@vmw/transport/bus.api";

import { CdsButton } from "@cds/react/button";

import "./styles.css";
import { CartesianGrid, Label, LabelList, Legend, Line, LineChart, PolarAngleAxis, PolarGrid, PolarRadiusAxis, Radar, RadarChart, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

export function RenderStreamState(props: EEGStreamProps) {
  const attention = props.currentTick.attention;
  const meditation = props.currentTick.meditation;
  const signalQuality = props.currentTick.signalQuality;
  return (
    <div>
      <h1>
        <span className="word">Signal {signalQuality}</span>
      </h1>
    </div>
  );
}

export class ListenButton extends BaseComponent<ListenButtonProps, any> {
  constructor(props: ListenButtonProps) {
    super(props);
    this.startStream = this.startStream.bind(this);
    this.stopStream = this.stopStream.bind(this);
  }

  startStream() {
    this.props.fireListenHandler(true);
  }

  stopStream() {
    this.props.fireListenHandler(false);
  }

  render() {
    if (this.props.listening) {
      return (
        <CdsButton status="danger" onClick={this.stopStream}>
          Stop Stream
        </CdsButton>
      );
    } else {
      return (
        <CdsButton status="success" onClick={this.startStream}>
          Start Stream
        </CdsButton>
      );
    }
  }
}

export class EEGStream extends BaseComponent<
  EEGStreamChannelProps,
  EEGStreamState
> {
  private ssmhS: MessageHandler<EEGEvent, any>;
  private ssmhA: MessageHandler<EEGEvent, any>;
  private ssmhM: MessageHandler<EEGEvent, any>;
  private ssmhEP: MessageHandler<EEGEvent, any>;
  private ssmhER: MessageHandler<EEGEvent, any>;

  constructor(props: EEGStreamChannelProps) {
    super(props);

    this.state = {
      currentTick: {
        signalQuality: 0,
        attention: 0,
        meditation: 0,
      },
      signalQualityData: [],
      attentionData: [],
      meditationData: [],
      eegPowerData: [],
      eegPowerMax: {
        delta: 0,
        theta: 0,
        lowAlpha: 0,
        highAlpha: 0,
        lowBeta: 0,
        highBeta: 0,
        lowGamma: 0,
        highGamma: 0,
      },
      eegPowerMin: {
        delta: 0,
        theta: 0,
        lowAlpha: 0,
        highAlpha: 0,
        lowBeta: 0,
        highBeta: 0,
        lowGamma: 0,
        highGamma: 0,
      },
      eegPowerRawData: [],

      listening: false
    };

    this.listenToStream = this.listenToStream.bind(this);

    const handler = (streamVal: EEGEvent) => {
      switch (streamVal.type) {
        case 0x02: {
          // let c = [...this.state.signalQualityData, { date: new Date(), value: streamVal.signalQuality, subject: "", index: 0 }]
          // if (c.length > 60) {
          //   c = c.slice(1)
          // }
          // this.setState({
          //   signalQualityData: c,
          // });
          this.setState({
            currentTick: {
              attention: 0,
              meditation: 0,
              signalQuality: streamVal.signalQuality
            }
          })
        }
          break;
        case 0x04: {
          let c = [...this.state.attentionData, { date: new Date(), value: streamVal.attention, subject: "", index: 0 }]
          if (c.length > 60) {
            c = c.slice(1)
          }
          this.setState({
            attentionData: c,
          });
        }
          break;
        case 0x05: {
          let c = [...this.state.meditationData, { date: new Date(), value: streamVal.meditation, subject: "", index: 0 }]
          if (c.length > 60) {
            c = c.slice(1)
          }
          this.setState({
            meditationData: c,
          });
        }
          break;
        case 0x83: {
          let d = []
          let max = JSON.parse(JSON.stringify(this.state.eegPowerMax));
          for (const [k, v] of Object.entries(streamVal.eegPower)) {
            if (v > max[k]) {
              max[k] = v
            }

            d.push({ date: new Date(), value: v / max[k], subject: k, index: 0 })
          }

          this.setState({
            eegPowerMax: max,
            eegPowerData: d,
          });
        }
          break;
        case 0x80: {
          let d = []

          for (const v of streamVal.eegRawPowerFft) {
            d.push({ date: new Date(), value: v.real, subject: "", index: v.imaginary })
          }

          this.setState({
            eegPowerRawData: d,
          });
        }
          break;
      }

    }

    // capture reference to our stream listener
    this.ssmhS = this.bus.listenStream(this.props.baseStreamChannel + "/signal-quality");
    this.ssmhA = this.bus.listenStream(this.props.baseStreamChannel + "/attention");
    this.ssmhM = this.bus.listenStream(this.props.baseStreamChannel + "/meditation");
    this.ssmhEP = this.bus.listenStream(this.props.baseStreamChannel + "/eeg-power");
    this.ssmhER = this.bus.listenStream(this.props.baseStreamChannel + "/eeg-raw");

    // handle every tick on the stream.
    this.ssmhS.handle(handler);
    this.ssmhA.handle(handler);
    this.ssmhM.handle(handler);
    this.ssmhEP.handle(handler);
    this.ssmhER.handle(handler);
  }

  listenToStream(listen: boolean) {
    if (listen) {
      // mark simpleStreamChannel as galactic (subscribe to destination)
      this.bus.markChannelAsGalactic(this.props.baseStreamChannel + "/signal-quality");
      this.bus.markChannelAsGalactic(this.props.baseStreamChannel + "/attention");
      this.bus.markChannelAsGalactic(this.props.baseStreamChannel + "/meditation");
      this.bus.markChannelAsGalactic(this.props.baseStreamChannel + "/eeg-power");
      this.bus.markChannelAsGalactic(this.props.baseStreamChannel + "/eeg-raw");
      this.setState({
        listening: true
      });
    } else {
      // mark simpleStreamChannel as local (unsubscribe from destination)
      this.bus.markChannelAsLocal(this.props.baseStreamChannel + "/signal-quality");
      this.bus.markChannelAsLocal(this.props.baseStreamChannel + "/attention");
      this.bus.markChannelAsLocal(this.props.baseStreamChannel + "/meditation");
      this.bus.markChannelAsLocal(this.props.baseStreamChannel + "/eeg-power");
      this.bus.markChannelAsLocal(this.props.baseStreamChannel + "/eeg-raw");
      this.setState({
        listening: false
      });
    }
  }

  componentWillUnmount() {
    // stop listening to the channel.
    this.ssmhS.close();
    this.ssmhA.close();
    this.ssmhM.close();
    this.ssmhEP.close();
    this.ssmhER.close();
  }

  render() {
    const formatXAxis = (tickItem: Date, index: any) => {
      return tickItem.getMinutes()+":"+tickItem.getSeconds();
    }

    return (
      <div className="simple-stream" cds-layout="grid gap:md">
        <div cds-layout="col@sm:3">
          <ListenButton
            listening={this.state.listening}
            fireListenHandler={this.listenToStream}
          />
        </div>
        <div cds-layout="col@sm:9">
          <RenderStreamState currentTick={this.state.currentTick} />
        </div>
        <div cds-layout="col@sm:12">
          <LineChart
            width={1400}
            height={200}
            data={this.state.attentionData}
            margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
          >
            <XAxis
              dataKey="date"
              scale="time"
              tickFormatter={(tick, index) => formatXAxis(tick, index)}
            />
            <YAxis domain={[0, 100]}>
              <Label angle={-90} value='attention' position='insideLeft' style={{ textAnchor: 'middle' }} />
            </YAxis>
            {/* <Tooltip /> */}
            <CartesianGrid stroke="#f5f5f5" />
            <Line type="monotone" dataKey="value" stroke="#ff7300" yAxisId={0} isAnimationActive={false}>
              {/* <LabelList dataKey="value" position="right" /> */}
            </Line>
          </LineChart>
        </div>
        <div cds-layout="col@sm:12">
          <LineChart
            width={1400}
            height={200}
            data={this.state.meditationData}
            margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
          >
            <XAxis
              dataKey="date"
              scale="time"
              tickFormatter={(tick, index) => formatXAxis(tick, index)}
            />
            <YAxis domain={[0, 100]}>
              <Label angle={-90} value='meditation' position='insideLeft' style={{ textAnchor: 'middle' }} />
            </YAxis>
            {/* <Tooltip /> */}
            <CartesianGrid stroke="#f5f5f5" />
            <Line type="monotone" dataKey="value" stroke="#ff7300" yAxisId={0} isAnimationActive={false}>
              {/* <LabelList dataKey="value" position="insideRight" /> */}
            </Line>
          </LineChart>
        </div>
        {/* <div cds-layout="col@sm:12">
          <LineChart
            width={1400}
            height={200}
            data={this.state.eegPowerRawData}
            margin={{ top: 5, right: 20, left: 10, bottom: 5 }}
          >
            <XAxis dataKey="index" label={"frequency"} />
            <YAxis>
              <Label angle={-90} value='signal strength' position='insideLeft' style={{ textAnchor: 'middle' }} />
            </YAxis>
            <Tooltip />
            <CartesianGrid stroke="#f5f5f5" />
            <Line type="monotone" dataKey="value" stroke="#ff7300" yAxisId={0} />
          </LineChart>
        </div> */}
        <RadarChart cx={300} cy={250} outerRadius={150} width={500} height={500} data={this.state.eegPowerData}>
          <PolarGrid gridType='circle' />
          <PolarAngleAxis dataKey="subject" />
          <PolarRadiusAxis angle={30} domain={[0, 1]} />
          <Radar dataKey="value" stroke="#82ca9d" fill="#82ca9d" fillOpacity={0.6} />
        </RadarChart>
      </div>
    );
  }
}
