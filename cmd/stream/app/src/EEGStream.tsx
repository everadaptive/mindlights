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

export function RenderStreamState(props: EEGStreamProps) {
  const attention = props.currentTick.attention;
  const meditation = props.currentTick.meditation;
  const signalQuality = props.currentTick.signalQuality;
  return (
    <div className="response">
      <h1>
        <span className="word">{attention}</span>
        <span className="word">{meditation}</span>
        <span className="word">{signalQuality}</span>
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
  private ssmh: MessageHandler<APIResponse<EEGEvent>, any>;

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
      meditationData:[],
      listening: false
    };

    this.listenToStream = this.listenToStream.bind(this);

    // capture reference to our stream listener
    this.ssmh = this.bus.listenStream(this.props.baseStreamChannel + "/signal-quality");

    // handle every tick on the stream.
    this.ssmh.handle((streamVal: APIResponse<EEGEvent>) => {
      switch (streamVal.payload.type) {
        case 0x02:
          let c = this.state.attentionData.concat({ date: new Date(), value: streamVal.payload.attention })
          if (c.length > 200) {
            c.shift()
          }
          this.setState({
            attentionData: c,
          });
      }

    });
  }

  listenToStream(listen: boolean) {
    if (listen) {
      // mark simpleStreamChannel as galactic (subscribe to destination)
      this.bus.markChannelAsGalactic(this.props.baseStreamChannel + "/signal-quality");
      this.setState({
        listening: true
      });
    } else {
      // mark simpleStreamChannel as local (unsubscribe from destination)
      this.bus.markChannelAsLocal(this.props.baseStreamChannel + "/signal-quality");
      this.setState({
        listening: false
      });
    }
  }

  componentWillUnmount() {
    // stop listening to the channel.
    this.ssmh.close();
  }

  render() {
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
      </div>
    );
  }
}
