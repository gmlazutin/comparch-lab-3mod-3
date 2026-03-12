package gin

//todo: severe refactoring needed here

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/gin/gen"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/util"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth/session"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/contactbook"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
	oapitypes "github.com/oapi-codegen/runtime/types"
)

type ginApiSrvCtxKey int

const (
	ginApiSrvSession       ginApiSrvCtxKey = 1
	ginApiSrvLastAuthError ginApiSrvCtxKey = 2
)

type Options struct {
	Opts api.APIServerOptions
}

type APIServer struct {
	opts Options
	http *http.Server
}

// implements ServerInterface
type serverMethods struct {
	server *APIServer
}

func NewAPIServer(options Options) (*APIServer, error) {
	//enforce ReleaseMode for gin to avoid custom debug info
	gin.SetMode(gin.ReleaseMode)

	if options.Opts.Logger == nil {
		options.Opts.Logger = logging.EmptyLogger()
	}
	options.Opts.Logger = options.Opts.Logger.With(logging.Service("ginApiServer"))

	r := gin.New()

	srv := &APIServer{
		opts: options,
		http: &http.Server{
			Addr:    options.Opts.Addr,
			Handler: r,
		},
	}

	corsconfig := cors.Config{
		AllowOrigins:     []string{options.Opts.PublicUrl},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}
	if err := corsconfig.Validate(); err != nil {
		return nil, fmt.Errorf("ginApiServer: cors config validation fail: %w", err)
	}

	r.Use(cors.New(corsconfig))

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

	return srv, nil
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

func (s *APIServer) authFunc(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
	ginCtx := ginmiddleware.GetGinContext(ctx)
	hdr := ginCtx.Request.Header.Get("Authorization")
	sess, err := util.ValidateAuthTkn(ginCtx, hdr, ai, s.opts.Opts.AuthService)
	if err != nil {
		//todo: hack: as gin-middleware and oapi-codegen have different
		//errorFunc signatures, write an actual error into Ctx to
		//gain it from handler properly
		ginCtx.Set(ginApiSrvLastAuthError, err)
		return ai.NewError(err)
	}
	ginCtx.Set(ginApiSrvSession, sess)
	return nil
}

func respondError(c *gin.Context, code int, step string, err string) {
	c.JSON(code, gen.ErrorObject{
		Step:  step,
		Error: err,
	})
}

//todo: hide server errors properly

func (s *APIServer) reqValidateStepHandler(c *gin.Context, message string, statusCode int) {
	lasterr, ok := c.Get(ginApiSrvLastAuthError)

	step := "REQUEST_VALIDATE"

	if ok {
		lasterror := lasterr.(error)
		if errors.Is(lasterror, service.ErrInvalidToken) {
			statusCode = http.StatusUnauthorized
		} else {
			statusCode = http.StatusInternalServerError
		}
	}

	respondError(c, statusCode, step, message)
}

func getGinCtxSession(c *gin.Context) *session.Session {
	v, _ := c.Get(ginApiSrvSession)
	return v.(*session.Session)
}

//serverMethods

func bindGinParamsJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		respondError(c, http.StatusInternalServerError, "REQEST_PARAMETERS_BIND", err.Error())
		return err
	}

	return nil
}

func makeAuthObject(tkn string, session *session.Session) *gen.AuthObject {
	return &gen.AuthObject{
		Expires: session.Expires,
		Token:   tkn,
	}
}

func (si *serverMethods) AuthUser(c *gin.Context) {
	var req gen.AuthRequest
	if err := bindGinParamsJSON(c, &req); err != nil {
		return
	}

	sess, tkn, err := si.server.opts.Opts.AuthService.AuthUserByPassword(c, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrIncorrectPassword) {
			respondError(c, http.StatusUnauthorized, "AUTH", err.Error())
			return
		}

		respondError(c, http.StatusInternalServerError, "AUTH", err.Error())
		return
	}

	c.JSON(http.StatusOK, gen.AuthResponse{
		Auth: *makeAuthObject(tkn, sess),
	})
}

func (si *serverMethods) RegisterUser(c *gin.Context) {
	var req gen.RegisterRequest
	if err := bindGinParamsJSON(c, &req); err != nil {
		return
	}

	sess, tkn, err := si.server.opts.Opts.AuthService.CreateUserSimple(c, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			respondError(c, http.StatusConflict, "AUTH_REGISTER", err.Error())
			return
		}

		respondError(c, http.StatusInternalServerError, "AUTH_REGISTER", err.Error())
		return
	}

	c.JSON(http.StatusOK, gen.RegisterResponse{
		Auth: *makeAuthObject(tkn, sess),
	})
}

func parseSelector(sel gen.SelectorObject) *contactbook.Selector {
	var offset uint
	if sel.Offset != nil {
		offset = uint(*sel.Offset)
	}
	return &contactbook.Selector{
		Limit:  uint(sel.Limit),
		Offset: offset,
	}
}

func genContactResponse(cont contactbook.Contact, ph []contactbook.Phone) *gen.ContactObject {
	var noteptr *string
	if len(cont.Info.Note) > 0 {
		noteptr = &cont.Info.Note
	}

	var phonesresp *[]gen.PhoneObject
	if len(ph) > 0 {
		tmp := make([]gen.PhoneObject, len(ph))
		phonesresp = &tmp
		for i, v := range ph {
			var prim *bool
			if v.Info.Primary == true {
				tmp := true
				prim = &tmp
			}

			(*phonesresp)[i] = gen.PhoneObject{
				Id:        int(v.PhoneID.ID),
				IsPrimary: prim,
				Phone:     v.Info.Phone,
			}
		}
	}

	return &gen.ContactObject{
		Birthday: oapitypes.Date{
			Time: cont.Info.Birthday,
		},
		Id:     int(cont.ContactID.ID),
		Name:   cont.Info.Name,
		Note:   noteptr,
		Phones: phonesresp,
	}
}

func (si *serverMethods) AddContact(c *gin.Context) {
	var req gen.AddContactRequest
	if err := bindGinParamsJSON(c, &req); err != nil {
		return
	}
	sess := getGinCtxSession(c)
	var note string
	if req.Note != nil {
		note = *req.Note
	}

	var phones = make([]contactbook.PhoneInfo, len(req.InitialPhones))
	for i, v := range req.InitialPhones {
		prim := true
		if v.IsPrimary == nil {
			prim = false
		} else {
			prim = *v.IsPrimary
		}

		phones[i] = contactbook.PhoneInfo{
			Phone:   v.Phone,
			Primary: prim,
		}
	}

	cont, ph, err := si.server.opts.Opts.ContactbookService.AddContact(c, sess.UserID, contactbook.ContactInfo{
		Name:     req.Name,
		Birthday: req.Birthday.Time,
		Note:     note,
	}, phones)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "ADD_CONTACT", err.Error())
		return
	}

	c.JSON(http.StatusOK, gen.AddContactResponse{
		Contact: *genContactResponse(*cont, ph),
	})
}

func (si *serverMethods) DeleteContact(c *gin.Context, contactId int) {
	sess := getGinCtxSession(c)

	err := si.server.opts.Opts.ContactbookService.DeleteContact(c, contactbook.ContactID{
		ID:     uint(contactId),
		UserID: sess.UserID,
	})
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			respondError(c, http.StatusNotFound, "DELETE_CONTACT", err.Error())
			return
		}

		respondError(c, http.StatusInternalServerError, "DELETE_CONTACT", err.Error())
		return
	}

	c.JSON(http.StatusOK, gen.DeleteContactResponse{})
}

func (si *serverMethods) GetContact(c *gin.Context, contactId int) {
	var req gen.GetContactRequest
	if err := bindGinParamsJSON(c, &req); err != nil {
		return
	}
	sess := getGinCtxSession(c)
	var preload *contactbook.PhonesPreload
	if req.Preload != nil && req.Preload.Enabled {
		var prim bool
		if req.Preload.PrimaryOnly != nil {
			prim = *req.Preload.PrimaryOnly
		}

		preload = &contactbook.PhonesPreload{
			PrimaryOnly: prim,
		}
	}
	var notes bool
	if req.WithNote != nil {
		notes = *req.WithNote
	}

	cont, ph, err := si.server.opts.Opts.ContactbookService.GetContact(c, contactbook.ContactID{
		ID:     uint(contactId),
		UserID: sess.UserID,
	}, preload, notes)
	if err != nil {
		if errors.Is(err, service.ErrContactNotFound) {
			respondError(c, http.StatusNotFound, "GET_CONTACT", err.Error())
			return
		}

		respondError(c, http.StatusInternalServerError, "GET_CONTACT", err.Error())
		return
	}

	c.JSON(http.StatusOK, gen.GetContactResponse{
		Contact: *genContactResponse(*cont, ph),
	})
}

func (si *serverMethods) GetContacts(c *gin.Context) {
	var req gen.GetContactsRequest
	if err := bindGinParamsJSON(c, &req); err != nil {
		return
	}
	sess := getGinCtxSession(c)

	conts, err := si.server.opts.Opts.ContactbookService.GetContacts(c, sess.UserID, *parseSelector(req.Selector))
	if err != nil {
		respondError(c, http.StatusInternalServerError, "GET_CONTACTS", err.Error())
		return
	}

	var resp gen.GetContactsResponse
	resp.Contacts = make([]gen.ContactObject, len(conts))
	for i, v := range conts {
		resp.Contacts[i] = *genContactResponse(v.Contact, v.Phones)
	}

	c.JSON(http.StatusOK, resp)
}
