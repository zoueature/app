package app

import (
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"gitlab.jiebu.com/base/config"
	"gitlab.jiebu.com/base/log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type appRunConf struct {
	beforeRouteRegister func()
	beforeServe         func()
	afterRouteRegister  func()
	beforeShutDown      func()
	routeRegister       func(c *gin.Engine)
}
type App struct {
	engine  *gin.Engine
	cfg     *config.AppConfig
	runConf *appRunConf
}

// NewApp 实例化app对象
func NewApp(cfg *config.AppConfig) *App {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	app := &App{
		engine:  gin.Default(),
		cfg:     cfg,
		runConf: &appRunConf{},
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
	if app.runConf.beforeServe != nil {
		app.runConf.beforeServe()
	}
	go func() error {
		println("EA App had start up ..........")
		return app.engine.Run(app.cfg.Listen)
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGKILL, syscall.SIGTERM)
	<-ch
	if app.runConf.beforeShutDown != nil {
		app.runConf.beforeShutDown()
	}
	println("EA App had shut down ........")
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

type OpFunc func(app *appRunConf)

func (f OpFunc) apply(app *appRunConf) {
	f(app)
}

type Conf interface {
	apply(app *appRunConf)
}

// BeforeServe 服务启动前的回调
func BeforeServe(f func()) Conf {
	return OpFunc(func(app *appRunConf) {
		app.beforeServe = f
	})
}

// BeforeRegister 路由注册前回调
func BeforeRegister(f func()) Conf {
	return OpFunc(func(app *appRunConf) {
		app.beforeRouteRegister = f
	})
}

// AfterRegister 路由注册后回调
func AfterRegister(f func()) Conf {
	return OpFunc(func(app *appRunConf) {
		app.afterRouteRegister = f
	})
}

// BeforeShutdown 服务关闭前回调
func BeforeShutdown(f func()) Conf {
	return OpFunc(func(app *appRunConf) {
		app.beforeShutDown = f
	})
}

// RouteRegister 路由注册器
func RouteRegister(f func(c *gin.Engine)) Conf {
	return OpFunc(func(app *appRunConf) {
		app.routeRegister = f
	})
}

// RunApp 启动http服务
func RunApp(cfg *config.Configuration, opts ...Conf) {
	err := log.Configure(cfg)
	if err != nil {
		panic(err)
	}
	app := NewApp(cfg.App)
	for _, opt := range opts {
		opt.apply(app.runConf)
	}
	app.Engine().Use(logIdInjector)
	if app.runConf.beforeRouteRegister != nil {
		// 路由注册前回调
		app.runConf.beforeRouteRegister()
	}
	if app.runConf.routeRegister != nil {
		// 路由注册
		println("Register the http route")
		app.runConf.routeRegister(app.Engine())
	}
	if app.runConf.afterRouteRegister != nil {
		// 路由注册后回调
		app.runConf.afterRouteRegister()
	}
	// 启动http服务
	app.Run()
}
