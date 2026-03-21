package gin

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/gin/gen"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/util"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

type ginApiSrvCtxKey int

const (
	ginApiSrvSession       ginApiSrvCtxKey = 1
	ginApiSrvLastAuthError ginApiSrvCtxKey = 2
)

type Options struct {
	Opts      api.APIServerOptions
	PublicUrl string
	StaticFS  fs.FS
}

type APIServer struct {
	opts Options
	http *http.Server

	httpFS http.FileSystem
}

func NewAPIServer(options Options) (*APIServer, error) {
	//enforce ReleaseMode for gin to avoid custom debug info
	gin.SetMode(gin.ReleaseMode)

	if options.Opts.Logger == nil {
		options.Opts.Logger = logging.EmptyLogger()
	}
	options.Opts.Logger = options.Opts.Logger.With(logging.Service("ginApiServer"))

	if len(options.Opts.Addr) == 0 {
		options.Opts.Addr = ":8080"
	}

	r := gin.New()

	var httpFS http.FileSystem
	if options.StaticFS != nil {
		httpFS = http.FS(options.StaticFS)
	}
	srv := &APIServer{
		opts: options,
		http: &http.Server{
			Addr:    options.Opts.Addr,
			Handler: r,
		},
		httpFS: httpFS,
	}

	r.NoRoute(srv.noRouteFunc)

	if len(options.PublicUrl) > 0 {
		corsconfig := cors.Config{
			AllowOrigins:     []string{options.PublicUrl},
			AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE"},
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			AllowCredentials: true,
		}
		if err := corsconfig.Validate(); err != nil {
			return nil, fmt.Errorf("ginApiServer: cors config validation fail: %w", err)
		}

		r.Use(cors.New(corsconfig))
	}

	if options.Opts.Logger.Enabled(nil, slog.LevelDebug) {
		r.Use(srv.slogLogger)
	}

	swagger, err := gen.GetSwagger()
	if err != nil {
		panic(fmt.Sprintf("ginApiServer: unable to load swagger spec: %s", err.Error()))
	}

	si := &serverMethods{
		server: srv,
	}
	gen.RegisterHandlersWithOptions(r, si, gen.GinServerOptions{
		ErrorHandler: func(ctx *gin.Context, err error, i int) {
			srv.reqValidateStepHandler(ctx, err.Error(), i)
		},
		Middlewares: []gen.MiddlewareFunc{
			gen.MiddlewareFunc(ginmiddleware.OapiRequestValidatorWithOptions(swagger, &ginmiddleware.Options{
				ErrorHandler: srv.reqValidateStepHandler,
				Options: openapi3filter.Options{
					AuthenticationFunc: srv.authFunc,
				},
			})),
		},
	})

	//for static mode
	if httpFS != nil {
		assets, err := fs.Sub(options.StaticFS, "assets")
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, fmt.Errorf("ginApiServer: unable to serve static assets: %s", err.Error())
			}
		} else {
			r.StaticFS("/assets", http.FS(assets))
		}
	}

	return srv, nil
}

func (s *APIServer) Addr() string {
	return s.http.Addr
}

func (s *APIServer) Start() error {
	err := s.http.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *APIServer) Stop(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *APIServer) slogLogger(c *gin.Context) {
	start := time.Now()

	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()

	s.opts.Opts.Logger.Debug("HTTP request",
		slog.String("method", c.Request.Method),
		slog.String("path", c.Request.URL.Path),
		slog.Int("status", status),
		slog.Duration("latency", latency),
		slog.String("client_ip", c.ClientIP()),
	)
}

func (s *APIServer) noRouteFunc(c *gin.Context) {
	if s.httpFS == nil || (s.httpFS != nil && strings.HasPrefix(c.Request.URL.Path, "/api")) {
		return
	}

	f, err := s.httpFS.Open("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html not found")
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	http.ServeContent(c.Writer, c.Request, "index.html", stat.ModTime(), f)
}

func (s *APIServer) authFunc(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	ginCtx := ginmiddleware.GetGinContext(ctx)
	hdr := ginCtx.Request.Header.Get("Authorization")
	sess, err := util.ValidateAuthTkn(ginCtx, hdr, ai, s.opts.Opts.AuthService)
	if err != nil {
		//as gin-middleware and oapi-codegen have different
		//errorFunc signatures, write an actual error into Ctx to
		//gain it from handler properly
		ginCtx.Set(ginApiSrvLastAuthError, err)
		return ai.NewError(err)
	}
	ginCtx.Set(ginApiSrvSession, sess)
	return nil
}

func (s *APIServer) respondError(c *gin.Context, code int, step string, err string) {
	c.JSON(code, gen.ErrorObject{
		Step:  step,
		Error: err,
	})
}

func (s *APIServer) translateError(err error) (int, string) {
	msg := "please try again later"

	cerr := service.CustomValidationError{}
	if errors.As(err, &cerr) {
		return http.StatusBadRequest, cerr.Error()
	}

	if errors.Is(err, service.ErrIncorrectPassword) {
		return http.StatusUnauthorized, service.ErrIncorrectPassword.Error()
	}
	if errors.Is(err, service.ErrUserAlreadyExists) {
		return http.StatusConflict, service.ErrUserAlreadyExists.Error()
	}
	if errors.Is(err, service.ErrUserNotFound) {
		return http.StatusUnauthorized, service.ErrUserNotFound.Error()
	}
	if errors.Is(err, service.ErrInvalidToken) {
		return http.StatusUnauthorized, service.ErrInvalidToken.Error()
	}
	if errors.Is(err, service.ErrContactNotFound) {
		return http.StatusNotFound, service.ErrContactNotFound.Error()
	}

	s.opts.Opts.Logger.Error("unable to process error", logging.Error(err))
	return http.StatusInternalServerError, msg
}

func (s *APIServer) reqValidateStepHandler(c *gin.Context, message string, statusCode int) {
	lasterr, ok := c.Get(ginApiSrvLastAuthError)

	step := "REQUEST_VALIDATE"

	if ok {
		statusCode, message = s.translateError(lasterr.(error))
	}

	//todo: also mask the openapi validator errors
	s.respondError(c, statusCode, step, message)
}
