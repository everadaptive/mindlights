import { BaseComponent } from "./BaseComponent";
import {
  EEGStreamState,
  EEGStreamProps,
  EEGStreamChannelProps,
  ListenButtonProps,
  EEGEvent,
  DataPoint
} from "./EEGStreamModel";
import { APIResponse } from "@vmw/transport";
import { MessageHandler } from "@vmw/transport/bus.api";

import { CdsButton } from "@cds/react/button";

import "./styles.css";
import { Bar, BarChart, Cell, CartesianGrid, Label, LabelList, Legend, Line, LineChart, PolarAngleAxis, PolarGrid, PolarRadiusAxis, Radar, RadarChart, ReferenceLine, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { BAND_ORDER, isDeadBand, percentileRank } from "./bandCalibration";

// One colour per band, light -> dark in frequency order, so the bars carry the
// delta -> gamma ordering that a radar throws away.
const BAND_COLORS: Record<string, string> = {
  delta: "#cfe6e4", theta: "#a8d5d2", lowAlpha: "#7fc2c0", highAlpha: "#57adad",
  lowBeta: "#36969a", highBeta: "#217e84", lowGamma: "#13666e", highGamma: "#0b4e57",
};

// One CSV sample per second; ease across that window instead of snapping.
const TWEEN_MS = 1000;

// The raw-FFT scope is fed by EEG_RAW (0x80), which only a live headset emits -
// a CSV log has band power but no raw sample column, so on replay the plot is a
// permanently empty axis. Flip this back to true when driving from a headset.
const SHOW_RAW_FFT: boolean = false;

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
      // Marking a channel as galactic before the broker bridge is up silently
      // drops the subscription: the button flips to "Stop Stream" but no data
      // ever arrives. Stay disabled until App reports the fabric connected.
      return (
        <CdsButton
          status="success"
          disabled={!this.props.connected}
          onClick={this.startStream}
        >
          {this.props.connected ? "Start Stream" : "Connecting..."}
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

    // Every one of these updates derives new state from old state, so they must
    // use the updater form. Reading this.state directly loses history whenever
    // React batches two events before committing - the rolling arrays rebuild
    // from a stale (often empty) array and the chart lines vanish, while
    // currentTick keeps updating because a spread of stale state still looks
    // right. Under load that is not a rare race: at 10x replay the arrays never
    // grew past a single point.
    const ROLLING = 60;
    const push = (series: DataPoint[], point: DataPoint) => {
      const next = [...series, point];
      return next.length > ROLLING ? next.slice(next.length - ROLLING) : next;
    };

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
          this.setState(prev => ({
            currentTick: {
              ...prev.currentTick,
              signalQuality: streamVal.signalQuality
            }
          }))
        }
          break;
        case 0x04: {
          const point = { date: new Date(), value: streamVal.attention, subject: "", index: 0 };
          this.setState(prev => ({
            attentionData: push(prev.attentionData, point),
            currentTick: { ...prev.currentTick, attention: streamVal.attention },
          }));
        }
          break;
        case 0x05: {
          const point = { date: new Date(), value: streamVal.meditation, subject: "", index: 0 };
          this.setState(prev => ({
            meditationData: push(prev.meditationData, point),
            currentTick: { ...prev.currentTick, meditation: streamVal.meditation },
          }));
        }
          break;
        case 0x83: {
          const now = new Date();
          // Fixed band order, so the bars read low -> high frequency and the
          // radar spokes stay put between ticks.
          // Bands that never move in the calibration recording are dropped
          // rather than drawn as a permanent zero. Data-driven, so regenerating
          // bandCalibration.ts from a clean recording brings them straight back.
          const d = BAND_ORDER.filter(k => !isDeadBand(k)).map(k => ({
            date: now,
            value: (streamVal.eegPower as any)[k] ?? 0,
            rank: percentileRank(k, (streamVal.eegPower as any)[k] ?? 0),
            subject: k,
            index: 0,
          }));

          this.setState(prev => {
            const max = { ...prev.eegPowerMax } as any;
            for (const point of d) {
              if (point.value > max[point.subject]) {
                max[point.subject] = point.value;
              }
            }
            return { eegPowerMax: max, eegPowerData: d };
          });
        }
          break;
        case 0x80: {
          const now = new Date();
          const d = streamVal.eegRawPowerFft.map(v =>
            ({ date: now, value: v.real, subject: "", index: v.imaginary }));

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

  private streamChannels(): string[] {
    return [
      "/signal-quality",
      "/attention",
      "/meditation",
      "/eeg-power",
      "/eeg-raw",
    ].map(suffix => this.props.baseStreamChannel + suffix);
  }

  // subscribe/unsubscribe every stream channel at the broker
  private setGalactic(galactic: boolean) {
    for (const channel of this.streamChannels()) {
      if (galactic) {
        this.bus.markChannelAsGalactic(channel);
      } else {
        this.bus.markChannelAsLocal(channel);
      }
    }
  }

  listenToStream(listen: boolean) {
    this.setGalactic(listen);
    this.setState({ listening: listen });
  }

  componentDidUpdate(prevProps: EEGStreamChannelProps) {
    // The bridge dropping takes the galactic subscriptions with it, so re-arm
    // them once it comes back rather than leaving the charts silently frozen.
    if (!prevProps.connected && this.props.connected && this.state.listening) {
      this.setGalactic(true);
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
            connected={this.props.connected}
            fireListenHandler={this.listenToStream}
          />
        </div>
        <div cds-layout="col@sm:9">
          <RenderStreamState currentTick={this.state.currentTick} />
        </div>
        <div cds-layout="col@sm:12" className="chart-row">
          <LineChart
            width={1300}
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
            <Line type="monotone" dataKey="value" stroke="#ff7300" yAxisId={0} isAnimationActive={false} dot={false} />
          </LineChart>
          <div className="hero-value">{this.state.currentTick.attention}</div>
        </div>
        <div cds-layout="col@sm:12" className="chart-row">
          <LineChart
            width={1300}
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
            <Line type="monotone" dataKey="value" stroke="#ff7300" yAxisId={0} isAnimationActive={false} dot={false} />
          </LineChart>
          <div className="hero-value">{this.state.currentTick.meditation}</div>
        </div>
        {SHOW_RAW_FFT && (
          <div cds-layout="col@sm:12">
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
          </div>
        )}
        {/*
          Band power is plotted as a percentile against each band's own
          distribution, not as a raw value on a shared axis. Raw power spans
          three orders of magnitude between delta and highGamma, so a shared
          axis auto-fits to the loudest band and the rest collapse to nothing.
          A regular octagon here means "every band at its median".
        */}
        <div className="band-row">
          <RadarChart
            cx={300}
            cy={300}
            outerRadius={180}
            width={600}
            height={600}
            data={this.state.eegPowerData}
          >
            <PolarGrid gridType='circle' />
            <PolarAngleAxis dataKey="subject" />
            <PolarRadiusAxis angle={30} domain={[0, 100]} tickCount={5} />
            <Radar
              dataKey="rank"
              stroke="#0b7a7e"
              fill="#0b7a7e"
              fillOpacity={0.35}
              isAnimationActive={true}
              animationDuration={TWEEN_MS}
              animationEasing="ease-in-out"
            />
          </RadarChart>
          {/*
            The same percentile scaling as bars, in frequency order. The radar
            hides that delta -> highGamma is an ordered axis and implies the
            last band wraps around to the first; bars don't.
          */}
          <BarChart
            layout="vertical"
            width={700}
            height={600}
            data={this.state.eegPowerData}
            margin={{ top: 30, right: 40, left: 30, bottom: 30 }}
          >
            <XAxis type="number" domain={[0, 100]} tickCount={5}>
              <Label value='percentile of this band' position='insideBottom' offset={-8} />
            </XAxis>
            <YAxis type="category" dataKey="subject" width={80} interval={0} />
            <CartesianGrid stroke="#f5f5f5" horizontal={false} />
            <Tooltip
              formatter={(v: any, _n: any, item: any) =>
                [`p${Math.round(v)} (${item.payload.value.toLocaleString()})`, item.payload.subject]}
            />
            <Bar
              dataKey="rank"
              isAnimationActive={true}
              animationDuration={TWEEN_MS}
              animationEasing="ease-in-out"
            >
              {this.state.eegPowerData.map(d => (
                <Cell
                  key={d.subject}
                  fill={BAND_COLORS[d.subject] ?? "#82ca9d"}
                />
              ))}
            </Bar>
          </BarChart>
        </div>
      </div>
    );
  }
}
