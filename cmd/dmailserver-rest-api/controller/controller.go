package controller

import (
	"log/slog"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pjotrscholtze/dmailserver-rest-api/cmd/dmailserver-rest-api/repo"
	"github.com/pjotrscholtze/dmailserver-rest-api/models"
	"github.com/pjotrscholtze/dmailserver-rest-api/restapi/operations"
	"github.com/pjotrscholtze/dmailserver-rest-api/restapi/operations/email"
	"github.com/pjotrscholtze/dmailserver-rest-api/restapi/operations/fail2ban"
)

func SetupController(api *operations.DmailserverRestAPIAPI, sr repo.SetupRepo) {
	api.EmailAddEmailAccountHandler = email.AddEmailAccountHandlerFunc(func(aeap email.AddEmailAccountParams, i interface{}) middleware.Responder {
		has, err := sr.HasEmail(aeap.Body.Address)
		if has {
			slog.Info("[POST] /email", "duplicateEmail", true, "error", false, "errorMessage", err, "success", false, "address", aeap.Body.Address)
			return email.NewAddEmailAccountNotAcceptable()
		}

		if err != nil {
			slog.Info("[POST] /email", "duplicateEmail", false, "error", true, "errorMessage", err, "success", false, "address", aeap.Body.Address)
			return email.NewAddEmailAccountInternalServerError()
		}

		err = sr.AddEmail(*aeap.Body)
		if err != nil {
			slog.Info("[POST] /email", "duplicateEmail", false, "error", true, "errorMessage", err, "success", false, "address", aeap.Body.Address)
			return email.NewAddEmailAccountInternalServerError()
		}

		slog.Info("[POST] /email", "duplicateEmail", false, "error", false, "success", true, "address", aeap.Body.Address)
		return email.NewAddEmailAccountOK()
	})

	api.EmailDeleteEmailAccountHandler = email.DeleteEmailAccountHandlerFunc(func(deap email.DeleteEmailAccountParams, i interface{}) middleware.Responder {
		has, err := sr.HasEmail(deap.EmailAddress)
		if !has {
			slog.Info("[DELETE] /email", "emailExists", false, "error", false, "errorMessage", err, "success", false, "address", deap.EmailAddress)
			return email.NewDeleteEmailAccountNotFound()
		}

		if err != nil {
			slog.Info("[DELETE] /email", "emailExists", false, "error", true, "errorMessage", err, "success", false, "address", deap.EmailAddress)
			return email.NewDeleteEmailAccountInternalServerError()
		}

		err = sr.RemoveEmail(deap.EmailAddress)
		if err != nil {
			slog.Info("[DELETE] /email", "emailExists", true, "error", true, "errorMessage", err, "success", false, "address", deap.EmailAddress)
			return email.NewDeleteEmailAccountInternalServerError()
		}

		slog.Info("[DELETE] /email", "emailExists", true, "error", false, "success", true, "address", deap.EmailAddress)
		return email.NewDeleteEmailAccountOK()
	})

	api.EmailListEmailAccountsHandler = email.ListEmailAccountsHandlerFunc(func(leap email.ListEmailAccountsParams, i interface{}) middleware.Responder {
		emails, err := sr.ListEmail()
		if err != nil {
			slog.Info("[GET] /email", "error", true, "errorMessage", err, "success", false)
			return email.NewListEmailAccountsInternalServerError()
		}

		slog.Info("[GET] /email", "error", false, "success", true)
		return email.NewListEmailAccountsOK().WithPayload(emails)
	})

	api.EmailUpdateEmailAddressHandler = email.UpdateEmailAddressHandlerFunc(func(ueap email.UpdateEmailAddressParams, i interface{}) middleware.Responder {
		has, err := sr.HasEmail(ueap.EmailAddress)
		if !has {
			slog.Info("[PUT] /email", "error", false, "errorMessage", err, "success", false, "exists", false, "address", ueap.EmailAddress)
			return email.NewUpdateEmailAddressNotFound()
		}

		if err != nil {
			slog.Info("[PUT] /email", "error", true, "errorMessage", err, "success", false, "exists", true, "address", ueap.EmailAddress)
			return email.NewUpdateEmailAddressInternalServerError()
		}

		err = sr.UpdateEmail(models.EmailAccount{
			Address:  ueap.EmailAddress,
			Password: ueap.Password,
		})
		if err != nil {
			slog.Info("[PUT] /email", "error", true, "errorMessage", err, "success", false, "exists", true, "address", ueap.EmailAddress)
			return email.NewUpdateEmailAddressInternalServerError()
		}

		slog.Info("[PUT] /email", "error", false, "success", true, "exists", true, "address", ueap.EmailAddress)
		return email.NewUpdateEmailAddressOK()
	})

	api.Fail2banDeleteFail2banIPHandler = fail2ban.DeleteFail2banIPHandlerFunc(func(dfi fail2ban.DeleteFail2banIPParams, i interface{}) middleware.Responder {
		has, err := sr.HasFail2banIp(dfi.IP)
		if !has {
			slog.Info("[DELETE] /fail2ban", "error", false, "errorMessage", err, "success", false, "exists", false, "ip", dfi.IP)
			return fail2ban.NewDeleteFail2banIPNotFound()
		}

		if err != nil {
			slog.Info("[DELETE] /fail2ban", "error", true, "errorMessage", err, "success", false, "exists", true, "ip", dfi.IP)
			return fail2ban.NewDeleteFail2banIPInternalServerError()
		}

		err = sr.RemoveFail2ban(dfi.IP)
		if err != nil {
			slog.Info("[DELETE] /fail2ban", "error", true, "errorMessage", err, "success", false, "exists", true, "ip", dfi.IP)
			return fail2ban.NewDeleteFail2banIPInternalServerError()
		}

		slog.Info("[DELETE] /fail2ban", "error", false, "success", true, "exists", true, "ip", dfi.IP)
		return fail2ban.NewDeleteFail2banIPOK()
	})

	api.Fail2banGetFail2banIpsHandler = fail2ban.GetFail2banIpsHandlerFunc(func(gfip fail2ban.GetFail2banIpsParams, i interface{}) middleware.Responder {
		ips, err := sr.ListFail2ban()
		if err != nil {
			slog.Info("[GET] /fail2ban", "error", true, "errorMessage", err, "success", false)
			return fail2ban.NewGetFail2banIpsInternalServerError()
		}

		slog.Info("[GET] /fail2ban", "error", false, "success", true)
		return fail2ban.NewGetFail2banIpsOK().WithPayload(&ips)
	})

	api.Fail2banPostFail2banIPHandler = fail2ban.PostFail2banIPHandlerFunc(func(pfi fail2ban.PostFail2banIPParams, i interface{}) middleware.Responder {
    // 检查IP是否已经存在
    has, err := sr.HasFail2banIp(pfi.Ipaddress)
    if err != nil {
        slog.Info("[POST] /fail2ban", "error", true, "errorMessage", err, "success", false, "ip", pfi.Ipaddress)
        return fail2ban.NewPostFail2banIPInternalServerError()
    }
    
    if has {
        slog.Info("[POST] /fail2ban", "error", false, "success", false, "exists", true, "ip", pfi.Ipaddress)
        return fail2ban.NewPostFail2banIPOK() // IP已经在封禁列表中，返回成功
    }

    // 添加封禁（使用 AddFail2ban，不是 RemoveFail2ban）
    err = sr.AddFail2ban(pfi.Ipaddress)
    if err != nil {
        slog.Info("[POST] /fail2ban", "error", true, "errorMessage", err, "success", false, "ip", pfi.Ipaddress)
        return fail2ban.NewPostFail2banIPInternalServerError()
    }

    slog.Info("[POST] /fail2ban", "error", false, "success", true, "ip", pfi.Ipaddress)
    return fail2ban.NewPostFail2banIPOK()
})

	api.EmailAddEmailAliasHandler = email.AddEmailAliasHandlerFunc(func(aeap email.AddEmailAliasParams, i interface{}) middleware.Responder {
		err := sr.AddAlias(aeap.Alias, aeap.EmailAddress)
		if err != nil {
			slog.Info("[POST] /email/{emailAddress}/aliasses", "error", true, "errorMessage", err, "success", false, "alias", aeap.Alias, "emailAddress", aeap.EmailAddress)
			return email.NewAddEmailAliasInternalServerError()
		}

		slog.Info("[POST] /email/{emailAddress}/aliasses", "error", false, "success", true, "alias", aeap.Alias, "emailAddress", aeap.EmailAddress)
		return email.NewAddEmailAliasOK()
	})

	api.EmailDeleteEmailAliasHandler = email.DeleteEmailAliasHandlerFunc(func(deap email.DeleteEmailAliasParams, i interface{}) middleware.Responder {
		err := sr.RemoveAlias(deap.Alias, deap.EmailAddress)
		if err != nil {
			slog.Info("[DELETE] /email/{emailAddress}/aliasses", "error", true, "errorMessage", err, "success", false, "alias", deap.Alias, "emailAddress", deap.EmailAddress)
			return email.NewAddEmailAliasInternalServerError()
		}

		slog.Info("[DELETE] /email/{emailAddress}/aliasses", "error", false, "success", true, "alias", deap.Alias, "emailAddress", deap.EmailAddress)
		return email.NewDeleteEmailAliasOK()
	})
}
