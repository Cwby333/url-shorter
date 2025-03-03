package urlrouter

import (
	"context"
	"log/slog"
	"time"
)

const (
	defaultTimeoutForSendPop = time.Second * 3
)

func (r Router) StartProcessPopAlias() {
	ticker := time.NewTicker(r.popAlias.TimeSend)

	go func() {
		for {
			select {
			case <-r.mainCtx.Done():
				return
			case <-ticker.C:
				popAlias, countOfReq := r.popAlias.GetMostPopularAlias()

				if popAlias == "" {
					r.logger.Info("zero get urls requests in duration", slog.Any("duration", r.popAlias.TimeSend.Seconds()))
					continue
				}

				ctx, cancel := context.WithTimeout(r.mainCtx, defaultTimeSendPopAlias)
				defer cancel()
				err := r.urlService.SendPopAlias(ctx, popAlias, countOfReq)

				if err != nil {
					r.logger.Error("send pop alias", slog.String("error", err.Error()))
					continue
				}

				r.logger.Info("success send pop alias")
			}
		}
	}()
}
