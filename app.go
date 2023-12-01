package app

import (
	"github.com/gin-gonic/gin"
	"github.com/jiebutech/config"
	"github.com/jiebutech/log"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type App struct {
	engine *gin.Engine
	cfg    *config.AppConfig
}

// NewApp 实例化app对象
func NewApp(cfg *config.AppConfig) *App {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	app := &App{
		engine: gin.Default(),
		cfg:    cfg,
	}
	return app
}

type AppOpt func(app *App)

// Run 运行服务
func (app *App) Run() {
	pid := os.Getpid()
	err := os.WriteFile("./pid", []byte(strconv.Itoa(pid)), 0755)
	if err != nil {
		panic("write pid fail: " + err.Error())
	}

	prome := ginprometheus.NewPrometheus("gin")
	prome.Use(app.engine)
	go func() error {
		return app.engine.Run(app.cfg.Listen)
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGKILL, syscall.SIGTERM)
	<-ch
}

// Engine 获取gin engine
func (app *App) Engine() *gin.Engine {
	return app.engine
}

// WithOpt 配置
func (app *App) WithOpt(opts ...AppOpt) *App {
	for _, opt := range opts {
		opt(app)
	}
	return app
}

func logIdInjector(c *gin.Context) {
	log.InjectLogID(c)
}

// RunApp 启动http服务
func RunApp(cfg *config.Configuration, opts ...AppOpt) {
	app := NewApp(cfg.App)
	app.Engine().Use(logIdInjector)
	for _, opt := range opts {
		opt(app)
	}
	err := log.Init(cfg.Log)
	if err != nil {
		panic(err)
	}
	app.Run()
}
