package app

import (
	"github.com/gin-gonic/gin"
	"github.com/zoueature/config"
	"github.com/zoueature/grpc"
	"github.com/zoueature/log"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"net"
	"net/http"
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
	registerRpcService  grpc.RegisterSvc
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

	serverNum := 0
	if app.runConf.routeRegister != nil {
		// 启动http服务
		if app.cfg.Listen == "" {
			panic("http listener is empty")
		}
		prome := ginprometheus.NewPrometheus("ea-app")
		prome.Use(app.engine)
		if app.runConf.beforeServe != nil {
			app.runConf.beforeServe()
		}
		go func() {
			listener, err := net.Listen("tcp", app.cfg.Listen)
			if err != nil {
				panic(err)
			}
			println("EA http(" + app.cfg.Listen + ") server starting up ..........")
			err = http.Serve(listener, app.engine)
			if err != nil {
				println(err.Error())
			}
		}()
		serverNum++
	}
	if app.runConf.registerRpcService != nil {
		// 启动grpc 服务
		if app.cfg.GrpcListen == "" {
			panic("rpc listener is empty")
		}
		go func() {
			println("EA grpc(" + app.cfg.GrpcListen + ") server starting up ..........")
			server := grpc.NewServer(app.cfg)
			err := server.Serve(app.runConf.registerRpcService)
			if err != nil {
				println(err.Error())
			}
		}()
		serverNum++
	}
	if serverNum == 0 {
		println("No server to run, exit ")
		return
	}
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

// RegisterRpcService 注册grpc service
func RegisterRpcService(svc grpc.RegisterSvc) Conf {
	return OpFunc(func(app *appRunConf) {
		app.registerRpcService = svc
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
	app.Engine().Use(logIdInjector, corsMiddleware())
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
