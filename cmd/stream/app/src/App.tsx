import { BaseComponent } from "./BaseComponent";
import { EEGStream } from "./EEGStream";
import "./styles.css";

export interface AppState {
  connected: boolean;
  currentTime: Date;
}

export default class App extends BaseComponent<any, AppState> {
  private baseStreamChannel = "headset-02";
  private clockInterval: ReturnType<typeof setInterval> | null = null;

  constructor(props?: any) {
    super(props);
    this.state = {
      connected: false,
      currentTime: new Date(),
    };
  }

  componentDidMount() {
    this.connectBroker();
    this.clockInterval = setInterval(() => {
      this.setState({ currentTime: new Date() });
    }, 1000);
  }

  componentWillUnmount() {
    if (this.clockInterval) clearInterval(this.clockInterval);
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
        window.location.hostname,
        30080,
        "/ws",
        false,
        "/topic",
        "/queue"
      );
    }
  }

  render() {
    return (
      <div className="main-container">
        <header>
          <h1>Neurosky Events: {this.baseStreamChannel} &nbsp; {this.state.currentTime.toLocaleTimeString()}</h1>
        </header>
        <hr />
        <EEGStream baseStreamChannel={this.baseStreamChannel} connected={this.state.connected} />
      </div>
    );
  }
}
