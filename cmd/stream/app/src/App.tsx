import { BaseComponent } from "./BaseComponent";
import { EEGStream } from "./EEGStream";
import "./styles.css";

export interface AppState {
  connected: boolean;
}

export default class App extends BaseComponent<any, AppState> {
  // simple stream is broadcasting on this channel over at transport-bus.io
  private baseStreamChannel = "HEADSET-03";

  constructor(props?: any) {
    super(props);
    this.state = {
      connected: false
    };
  }

  componentDidMount() {
    this.connectBroker();
  }

  connectBroker() {
    if (!this.state.connected) {
      this.fabric.connect(
        () => {
          this.log.info("application has connected to broker", "App.tsx");
          this.setState({
            connected: true
          });
        },
        () => {
          this.log.info("application has disconnected from broker", "App.tsx");
          this.setState({
            connected: false
          });
        },
        "transport-bus.io",
        443,
        "/ws",
        true,
        "/topic",
        "/queue"
      );
    }
  }

  render() {
    return (
      <div className="main-container">
        <header>
          <h1>Neurosky Events: {this.baseStreamChannel}</h1>
        </header>
        <hr />
        <EEGStream baseStreamChannel={this.baseStreamChannel} />
      </div>
    );
  }
}
