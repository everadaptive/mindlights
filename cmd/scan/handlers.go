package main

func poorSignalHandler(signalQuality int) {
	log.Infof("signalQuality: %d", signalQuality)
}

func attentionHandler(attention int) {
	log.Infof("attention: %d", attention)

}

func meditationHandler(meditation int) {
	log.Infof("meditation: %d", meditation)

}

func eegPowerHandler(eegPower MindflexEEGPower) {
	log.Infof("eegPower: %+v", eegPower)

}
