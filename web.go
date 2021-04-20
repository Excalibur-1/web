package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Excalibur-1/configuration"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

const (
	app      = "base"
	group    = "server"
	tag      = ""
	LogLevel = "LOG_LEVEL"
)

type Server func(eng *echo.Echo)

type Config struct {
	Addr string `json:"addr"`
	Gzip int    `json:"gzip"` // gzip压缩等级
	Csrf struct {
		TokenLength    uint8  `json:"tokenLength"`
		TokenLookup    string `json:"tokenLookup"`
		ContextKey     string `json:"contextKey"`
		CookieName     string `json:"cookieName"`
		CookieDomain   string `json:"cookieDomain"`
		CookiePath     string `json:"cookiePath"`
		CookieMaxAge   int    `json:"cookieMaxAge"`
		CookieSecure   bool   `json:"cookieSecure"`
		CookieHTTPOnly bool   `json:"cookieHttpOnly"`
	} `json:"csrf"` // 用于防跨站请求
	Cors struct {
		AllowOrigins     []string `json:"allowOrigins"`
		AllowMethods     []string `json:"allowMethods"`
		AllowHeaders     []string `json:"allowHeaders"`
		AllowCredentials bool     `json:"allowCredentials"`
		ExposeHeaders    []string `json:"exposeHeaders"`
		MaxAge           int      `json:"maxAge"`
	} `json:"cors"` // 用于跨域请求支持域名集
}

var SwagHandler echo.HandlerFunc

func init() {
	fmt.Println("Loading Web Engine ver:1.0.0")
}

func App(serv Server, namespace, systemId string, conf configuration.Configuration) {
	OptApp(serv, namespace, systemId, conf).Close(func() {})
}

type web struct {
	server *echo.Echo
}

func OptApp(serv Server, namespace, systemId string, conf configuration.Configuration) *web {
	debug := strings.ToUpper(os.Getenv(LogLevel))
	w := newWeb()
	var config Config
	if err := conf.Clazz(namespace, app, group, tag, systemId, &config); err != nil {
		panic("加载web引擎配置出错")
	}
	w.server.HideBanner = true
	if debug == "DEBUG" {
		w.server.Use(middleware.Logger())
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	w.server.Use(middleware.Recover())
	// 为请求生成唯一id
	// Dependency Injection & Route Register
	serv(w.server)
	w.server.GET("/healthy", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "我思故我在!")
	})
	if SwagHandler != nil {
		w.server.GET("/doc/*", SwagHandler)
	} else {
		w.server.Use(middleware.Gzip())
	}
	w.cors(config)
	w.csrf(config)
	w.start(config)
	return w
}

func newWeb() *web {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	return &web{server: echo.New()}
}

// 用于防跨站请求
func (w *web) csrf(config Config) {
	if len(config.Csrf.TokenLookup) > 0 {
		w.server.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
			TokenLength:    config.Csrf.TokenLength,
			TokenLookup:    config.Csrf.TokenLookup,
			ContextKey:     config.Csrf.ContextKey,
			CookieName:     config.Csrf.CookieName,
			CookieDomain:   config.Csrf.CookieDomain,
			CookiePath:     config.Csrf.CookiePath,
			CookieMaxAge:   config.Csrf.CookieMaxAge,
			CookieSecure:   config.Csrf.CookieSecure,
			CookieHTTPOnly: config.Csrf.CookieHTTPOnly,
		}))
	}
}

// 用于跨域请求支持
func (w *web) cors(config Config) {
	if len(config.Cors.AllowOrigins) > 0 {
		w.server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:     config.Cors.AllowOrigins,
			AllowMethods:     config.Cors.AllowMethods,
			AllowHeaders:     config.Cors.AllowHeaders,
			AllowCredentials: config.Cors.AllowCredentials,
			ExposeHeaders:    config.Cors.ExposeHeaders,
			MaxAge:           config.Cors.MaxAge,
		}))
	}
}

func (w *web) start(config Config) {
	go func() {
		if err := w.server.Start(config.Addr); err != nil {
			fmt.Println("Echo Engine Start has error")
		}
	}()
}

func (w *web) Close(close func()) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	close()
	w.shutdown()
}

func (w *web) shutdown() {
	fmt.Println("Web Engine Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.server.Shutdown(ctx); err != nil {
		fmt.Println("Web Engine Shutdown has error")
	} else {
		fmt.Println("Web Engine exiting")
	}
}
