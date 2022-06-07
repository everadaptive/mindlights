import { MessageFunction } from "@vmw/transport/bus.api";

export interface DataPoint {
    date: Date;
    index: number;
    value: number;
    subject: string;
}

export interface ComplexValue {
    real: number;
    imaginary: number;
}

export interface EEGPower {
    delta: number;
    theta: number;
    lowAlpha: number;
    highAlpha: number;
    lowBeta: number;
    highBeta: number;
    lowGamma: number;
    highGamma: number;
}

export interface EEGEvent {
    type: number;
    source: string;
    signalQuality: number;
    attention: number;
    meditation: number;
    eegPower: EEGPower;
    eegRawPower: number;
    eegRawPowerFft: ComplexValue[];
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
    eegPowerRawData: DataPoint[];

    eegPowerData: DataPoint[];
    eegPowerMin: EEGPower;
    eegPowerMax: EEGPower;

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
