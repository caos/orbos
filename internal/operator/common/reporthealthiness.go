package common

func ReportHealthiness(healthyChan chan<- bool, err error, reportSuccess bool) {
	go func() {
		if err != nil {
			healthyChan <- false
			return
		}
		if reportSuccess {
			healthyChan <- true
		}
	}()
}
