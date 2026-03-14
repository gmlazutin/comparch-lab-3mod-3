package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/api/gin/gen"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/logging"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/auth/session"
	"github.com/gmlazutin/comparch-lab-3mod-3/internal/service/contactbook"
	oapitypes "github.com/oapi-codegen/runtime/types"
)

// implements ServerInterface
type serverMethods struct {
	server *APIServer
}

func (si *serverMethods) getGinCtxSession(c *gin.Context) *session.Session {
	v, _ := c.Get(ginApiSrvSession)
	return v.(*session.Session)
}

func (si *serverMethods) bindGinParamsJSON(c *gin.Context, obj any) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		si.server.opts.Opts.Logger.Error("bind failture", logging.Error(err))
		si.server.respondError(c, http.StatusInternalServerError, "REQEST_PARAMETERS_BIND", "please try again later")
		return err
	}

	return nil
}

func (si *serverMethods) makeAuthObject(tkn string, session *session.Session) *gen.AuthObject {
	return &gen.AuthObject{
		Expires: session.Expires,
		Token:   tkn,
	}
}

func (si *serverMethods) respondGinTranslatedError(c *gin.Context, err error, step string) {
	status, msg := si.server.translateError(err)
	si.server.respondError(c, status, step, msg)
}

func (si *serverMethods) AuthUser(c *gin.Context) {
	var req gen.AuthRequest
	if err := si.bindGinParamsJSON(c, &req); err != nil {
		return
	}

	sess, tkn, err := si.server.opts.Opts.AuthService.AuthUserByPassword(c, req.Login, req.Password)
	if err != nil {
		si.respondGinTranslatedError(c, err, "AUTH")
		return
	}

	c.JSON(http.StatusOK, gen.AuthResponse{
		Auth: *si.makeAuthObject(tkn, sess),
	})
}

func (si *serverMethods) RegisterUser(c *gin.Context) {
	var req gen.RegisterRequest
	if err := si.bindGinParamsJSON(c, &req); err != nil {
		return
	}

	sess, tkn, err := si.server.opts.Opts.AuthService.CreateUserSimple(c, req.Login, req.Password)
	if err != nil {
		si.respondGinTranslatedError(c, err, "AUTH_REGISTER")
		return
	}

	c.JSON(http.StatusOK, gen.RegisterResponse{
		Auth: *si.makeAuthObject(tkn, sess),
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
	if err := si.bindGinParamsJSON(c, &req); err != nil {
		return
	}
	sess := si.getGinCtxSession(c)
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
		si.respondGinTranslatedError(c, err, "ADD_CONTACT")
		return
	}

	c.JSON(http.StatusOK, gen.AddContactResponse{
		Contact: *genContactResponse(*cont, ph),
	})
}

func (si *serverMethods) DeleteContact(c *gin.Context, contactId int) {
	sess := si.getGinCtxSession(c)

	err := si.server.opts.Opts.ContactbookService.DeleteContact(c, contactbook.ContactID{
		ID:     uint(contactId),
		UserID: sess.UserID,
	})
	if err != nil {
		si.respondGinTranslatedError(c, err, "DELETE_CONTACT")
		return
	}

	c.JSON(http.StatusOK, gen.DeleteContactResponse{})
}

func (si *serverMethods) GetContact(c *gin.Context, contactId int) {
	var req gen.GetContactRequest
	if err := si.bindGinParamsJSON(c, &req); err != nil {
		return
	}
	sess := si.getGinCtxSession(c)
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
		si.respondGinTranslatedError(c, err, "GET_CONTACT")
		return
	}

	c.JSON(http.StatusOK, gen.GetContactResponse{
		Contact: *genContactResponse(*cont, ph),
	})
}

func (si *serverMethods) GetContacts(c *gin.Context) {
	var req gen.GetContactsRequest
	if err := si.bindGinParamsJSON(c, &req); err != nil {
		return
	}
	sess := si.getGinCtxSession(c)

	conts, err := si.server.opts.Opts.ContactbookService.GetContacts(c, sess.UserID, *parseSelector(req.Selector))
	if err != nil {
		si.respondGinTranslatedError(c, err, "GET_CONTACTS")
		return
	}

	var resp gen.GetContactsResponse
	resp.Contacts = make([]gen.ContactObject, len(conts))
	for i, v := range conts {
		resp.Contacts[i] = *genContactResponse(v.Contact, v.Phones)
	}

	c.JSON(http.StatusOK, resp)
}
