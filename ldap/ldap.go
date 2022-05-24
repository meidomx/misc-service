package ldap

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/meidomx/misc-service/id"

	"github.com/jimlambrt/gldap"
)

type ldapServer struct {
	IdGen *id.IdGen

	BindBaseDN string
}

func StartService(idgen *id.IdGen, bindBaseDn string) {
	s, err := gldap.NewServer()
	if err != nil {
		log.Fatalf("unable to create server: %s", err.Error())
	}

	server := new(ldapServer)
	server.IdGen = idgen
	server.BindBaseDN = bindBaseDn

	// create a router and add a bind handler
	r, err := gldap.NewMux()
	if err != nil {
		log.Fatalf("unable to create router: %s", err.Error())
	}
	if err := r.Add(server.Add, gldap.WithLabel("Add")); err != nil {
		log.Fatalf("add op error: %s", err.Error())
	}
	if err := r.Delete(server.Delete, gldap.WithLabel("Delete")); err != nil {
		log.Fatalf("del op error: %s", err.Error())
	}
	if err := r.Search(server.Search, gldap.WithLabel("Search - Generic")); err != nil {
		log.Fatalf("search op error: %s", err.Error())
	}
	if err := r.Bind(server.Bind); err != nil {
		log.Fatalf("bind op error: %s", err.Error())
	}
	if err := s.Router(r); err != nil {
		log.Fatalf("router error: %s", err.Error())
	}
	go s.Run(":10389") // listen on port 10389

	// stop server gracefully when ctrl-c, sigint or sigterm occurs
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	select {
	case <-ctx.Done():
		log.Printf("\nstopping directory")
		s.Stop()
	}
}

/**
	mux.ExtendedOperation(d.handleStartTLS(t), gldap.ExtendedOperationStartTLS)
	mux.Modify(d.handleModify(t), gldap.WithLabel("Modify"))
    mux.Unbind
*/
func (this *ldapServer) Bind(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleBind"
	log.Println("operation:", op)
	resp := r.NewBindResponse(gldap.WithResponseCode(gldap.ResultInvalidCredentials))
	defer func() {
		_ = w.Write(resp)
	}()

	m, err := r.GetSimpleBindMessage()
	if err != nil {
		log.Println("not a simple bind message", "op", op, "err", err)
		return
	}

	log.Println("bind username:", m.UserName, "baseDN:", this.BindBaseDN)

	if m.AuthChoice != gldap.SimpleAuthChoice {
		// if it's not a simple auth request, then the bind failed...
		//TODO need support more auth methods
		return
	}

	// user + BaseDN
	//TODO need optimize the query
	entries, err := FindChildren(this.BindBaseDN)
	if err != nil {
		log.Println("find children error", "op", op, "err", err)
		return
	}
	for _, v := range entries {
		cns := v.GetAttributeValues("cn")
		for _, cn := range cns {
			if cn == m.UserName {
				log.Println("found bind user", "op", op, "DN", v.DN)
				values := v.GetAttributeValues("userPassword")
				if len(values) > 0 && string(m.Password) == values[0] {
					resp.SetResultCode(gldap.ResultSuccess)
					return
				}
			}
		}
	}

	// user is full DN
	entry, err := FindOneEntry(m.UserName)
	if err != nil {
		log.Println("FindOneEntry error", "op", op, "err", err)
		return
	}
	if len(entry.DN) > 0 {
		values := entry.GetAttributeValues("userPassword")
		if len(values) > 0 && string(m.Password) == values[0] {
			resp.SetResultCode(gldap.ResultSuccess)
			return
		}
	}

}

func (this *ldapServer) Search(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleSearchGeneric"
	log.Println("operation:", op)

	res := r.NewSearchDoneResponse(gldap.WithResponseCode(gldap.ResultNoSuchObject))
	defer w.Write(res)
	m, err := r.GetSearchMessage()
	if err != nil {
		log.Println("not a search message: %s", "op", op, "err", err)
		return
	}
	logSearchRequest(m)

	filter := m.Filter

	switch m.Scope {
	case gldap.BaseObject:
		{
			var entry *gldap.Entry
			var err error
			if len(m.BaseDN) > 0 {
				entry, err = FindOneEntry(m.BaseDN)
			} else {
				entry, err = FindSingleRoot()
			}
			if err != nil {
				log.Println("FindOneEntry error: %s", "op", op, "err", err)
				return
			}
			if len(entry.DN) <= 0 {
				return
			} else {
				//TODO filter entries
				var _ = filter
			}

			res.SetResultCode(gldap.ResultSuccess)
			result := r.NewSearchResponseEntry(entry.DN)
			for _, attr := range entry.Attributes {
				result.AddAttribute(attr.Name, attr.Values)
			}
			if err := w.Write(result); err != nil {
				log.Println("write result error: %s", "op", op, "err", err)
				return
			}
		}
	case gldap.SingleLevel:
		{
			var entries []*gldap.Entry
			var err error
			if len(m.BaseDN) > 0 {
				entries, err = FindChildren(m.BaseDN)
			} else {
				entries, err = FindAllRoots()
			}
			if err != nil {
				log.Println("FindChildren error: %s", "op", op, "err", err)
				return
			}
			if len(entries) <= 0 {
				return
			} else {
				//TODO filter entries
				var _ = filter
			}
			for _, e := range entries {
				result := r.NewSearchResponseEntry(e.DN)
				for _, attr := range e.Attributes {
					result.AddAttribute(attr.Name, attr.Values)
				}
				if err := w.Write(result); err != nil {
					log.Println("write result error: %s", "op", op, "err", err)
					return
				}
			}
			res.SetResultCode(gldap.ResultSuccess)
		}
	case gldap.WholeSubtree:
		{

			//TODO need support whole subtree
		}
	}
}

func (this *ldapServer) Delete(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleDelete"
	log.Println("operation:", op)

	res := r.NewResponse(gldap.WithResponseCode(gldap.ResultNoSuchObject), gldap.WithApplicationCode(gldap.ApplicationDelResponse))
	defer w.Write(res)
	m, err := r.GetDeleteMessage()
	if err != nil {
		log.Println("not a delete message: %s", "op", op, "err", err)
		return
	}
	log.Println("delete request", "dn", m.DN)

	entry, err := FindOneEntry(m.DN)
	if err != nil {
		log.Println("find entry error: %s", "op", op, "err", err)
		res.SetResultCode(gldap.ResultOperationsError)
		res.SetDiagnosticMessage(fmt.Sprintf("find entry error"))
		return
	}
	if len(entry.DN) > 0 {
		if err := DeleteEntry(m.DN); err != nil {
			log.Println("delete entry error: %s", "op", op, "err", err)
			res.SetResultCode(gldap.ResultOperationsError)
			res.SetDiagnosticMessage(fmt.Sprintf("delete entry error"))
			return
		}
	}

	return
}

func (this *ldapServer) Add(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleAdd"
	log.Println("operation:", op)

	res := r.NewResponse(gldap.WithApplicationCode(gldap.ApplicationAddResponse), gldap.WithResponseCode(gldap.ResultOperationsError))
	defer w.Write(res)
	m, err := r.GetAddMessage()
	if err != nil {
		log.Println("not an add message: %s", "op", op, "err", err)
		return
	}
	log.Println("add request", "dn", m.DN)

	entry, err := FindOneEntry(m.DN)
	if err != nil {
		log.Println("FindOneEntry error: %s", "op", op, "err", err)
		return
	}
	if len(entry.DN) > 0 {
		res.SetResultCode(gldap.ResultEntryAlreadyExists)
		res.SetDiagnosticMessage(fmt.Sprintf("entry exists for DN: %s", m.DN))
		return
	}

	attrs := map[string][]string{}
	for _, a := range m.Attributes {
		attrs[a.Type] = a.Vals
	}
	newEntry := gldap.NewEntry(m.DN, attrs)
	id, err := this.IdGen.Next()
	if err != nil {
		log.Println("generate id error: %s", "op", op, "err", err)
		return
	}
	err = SaveEntry(newEntry, id)
	if err != nil {
		log.Println("SaveEntry error: %s", "op", op, "err", err)
		return
	}
	res.SetResultCode(gldap.ResultSuccess)
}

func logSearchRequest(m *gldap.SearchMessage) {
	log.Println("search request",
		"baseDN", m.BaseDN,
		"scope", m.Scope,
		"filter", m.Filter,
		"attributes", m.Attributes,
	)
}
