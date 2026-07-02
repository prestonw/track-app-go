package timer

import "time"

type timeTicker struct {
	stop chan struct{}
}

func newTimeTicker(onTick func()) *timeTicker {
	t := &timeTicker{stop: make(chan struct{})}
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				onTick()
			case <-t.stop:
				return
			}
		}
	}()
	return t
}

func (t *timeTicker) Stop() {
	select {
	case <-t.stop:
	default:
		close(t.stop)
	}
}