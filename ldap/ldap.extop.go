package ldap

import (
	"fmt"
	"log"

	"github.com/jimlambrt/gldap"
)

func (this *ldapServer) ExtendedOperationStartTLS(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleStartTLS"
	log.Println("operation:", op)

	res := r.NewExtendedResponse(gldap.WithResponseCode(gldap.ResultSuccess))
	res.SetResponseName(gldap.ExtendedOperationStartTLS)
	w.Write(res)
	if err := r.StartTLS(this.serverTlsConfig); err != nil {
		log.Println("StartTLS Handshake error", "op", op, "err", err)
		res.SetDiagnosticMessage(fmt.Sprintf("StartTLS Handshake error : \"%s\"", err.Error()))
		res.SetResultCode(gldap.ResultOperationsError)
		w.Write(res)
		return
	}
	log.Println("StartTLS OK", "op", op)
}
