package price

import (
	"time"
)

func (ps PriceService) startRoutine() {
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					ps.Logger.Error("Unknown error", r)
				}

				time.Sleep(1 * time.Second)
			}()

			for _, sourceMetaSet := range ps.sourceMetas {
				for _, sourceMeta := range sourceMetaSet {
					go sourceMeta.Fetch(ps.Logger)

					// Sleep a bit. It seems that logger is not thread safe.
					// Without sleeping, log is often broken.
					// By sleeping, mitigate this problem.
					time.Sleep(time.Millisecond * 100)
				}
			}
		}()
	}
}
