import { MessageFunction } from "@vmw/transport/bus.api";

export interface DataPoint {
    date: Date;
    value: number;
}

export interface EEGEvent {
    type: number;

	signalQuality: number
	attention:     number
	meditation:    number
}

export interface StreamTickState {
    signalQuality: number;
    attention: number;
    meditation: number;
}

export interface EEGStreamState {
    currentTick: StreamTickState;
    signalQualityData: DataPoint[];
    meditationData: DataPoint[];
    attentionData: DataPoint[];

    listening: boolean;
}

export interface EEGStreamProps {
    currentTick: StreamTickState;
}

export interface EEGStreamChannelProps {
    baseStreamChannel: string;
}

export interface ListenButtonProps {
    listening: boolean;
    fireListenHandler: MessageFunction<boolean>;
}
