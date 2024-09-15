package v1

import (
	"avito/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func NewRouter(handler *echo.Echo, services *service.Services) {
	handler.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}", "method":"${method}","uri":"${uri}", "status":${status},"error":"${error}"}` + "\n",
		Output: setLogsFile(),
	}))
	handler.Use(middleware.Recover())

	v1 := handler.Group("/api")
	{
		v1.GET("/ping", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
		newTenderRoutes(v1.Group("/tenders"), services.Tender, services.Employee)
		newBidRoutes(v1.Group("/bids"), services.Bid, services.Employee, services.Tender)
	}
}

func setLogsFile() *os.File {
	file, err := os.OpenFile("./logs/requests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}
	return file
}
