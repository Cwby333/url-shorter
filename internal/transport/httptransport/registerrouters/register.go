package registerrouters

import (
	"fmt"
	"net/http"

	"github.com/Cwby333/url-shorter/internal/config"
	"github.com/Cwby333/url-shorter/internal/logger"
	"github.com/Cwby333/url-shorter/internal/transport/httptransport/urlrouter"
)

func New(urlService urlrouter.URLService, logger logger.Logger, cfg config.Owner) (*http.ServeMux, error) {
	const op = "internal/transports/httptransport/registerrouters/register.go/Register"

	mux := http.NewServeMux()

	routerURLS, err := urlrouter.New(urlService, logger, cfg)
	routerURLS.Run()

	if err != nil {
		return nil, fmt.Errorf("%s:%w", op, err)
	}

	mux.Handle("/", http.StripPrefix("/urls", routerURLS.Router))

	return mux, nil
}
