package ldap

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/meidomx/misc-service/id"

	"github.com/jimlambrt/gldap"
)

type ldapServer struct {
	IdGen *id.IdGen
}

func StartService(idgen *id.IdGen) {
	s, err := gldap.NewServer()
	if err != nil {
		log.Fatalf("unable to create server: %s", err.Error())
	}

	server := new(ldapServer)
	server.IdGen = idgen

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
	if err := r.Bind(server.bindHandler); err != nil {
		log.Fatalf("bind op error: %s", err.Error())
	}
	if err := r.Search(server.searchHandler); err != nil {
		log.Fatalf("search op error: %s", err.Error())
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
    mux.Bind(d.handleBind(t))
	mux.ExtendedOperation(d.handleStartTLS(t), gldap.ExtendedOperationStartTLS)
	mux.Search(d.handleSearchUsers(t), gldap.WithBaseDN(d.userDN), gldap.WithLabel("Search - Users"))
	mux.Search(d.handleSearchGroups(t), gldap.WithBaseDN(d.groupDN), gldap.WithLabel("Search - Groups"))
	mux.Search(d.handleSearchGeneric(t), gldap.WithLabel("Search - Generic"))
	mux.Modify(d.handleModify(t), gldap.WithLabel("Modify"))
	mux.Delete(d.handleDelete(t), gldap.WithLabel("Delete"))
    mux.Unbind
*/
func (this *ldapServer) Delete(w *gldap.ResponseWriter, r *gldap.Request) {

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

	entry, err := FindEntry(m.DN)
	if err != nil {
		log.Println("FindEntry error: %s", "op", op, "err", err)
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

func (this *ldapServer) bindHandler(w *gldap.ResponseWriter, r *gldap.Request) {
	resp := r.NewBindResponse(
		gldap.WithResponseCode(gldap.ResultInvalidCredentials),
	)
	defer func() {
		w.Write(resp)
	}()

	m, err := r.GetSimpleBindMessage()
	if err != nil {
		log.Printf("not a simple bind message: %s", err)
		return
	}

	if m.UserName == "alice" {
		resp.SetResultCode(gldap.ResultSuccess)
		log.Println("bind success")
		return
	}
}

func (this *ldapServer) searchHandler(w *gldap.ResponseWriter, r *gldap.Request) {
	resp := r.NewSearchDoneResponse()
	defer func() {
		w.Write(resp)
	}()
	m, err := r.GetSearchMessage()
	if err != nil {
		log.Printf("not a search message: %s", err)
		return
	}
	log.Printf("search base dn: %s", m.BaseDN)
	log.Printf("search scope: %d", m.Scope)
	log.Printf("search filter: %s", m.Filter)

	if strings.Contains(m.Filter, "uid=alice") || m.BaseDN == "uid=alice,ou=people,cn=example,dc=org" {
		entry := r.NewSearchResponseEntry(
			"uid=alice,ou=people,cn=example,dc=org",
			gldap.WithAttributes(map[string][]string{
				"objectclass": {"top", "person", "organizationalPerson", "inetOrgPerson"},
				"uid":         {"alice"},
				"cn":          {"alice eve smith"},
				"givenname":   {"alice"},
				"sn":          {"smith"},
				"ou":          {"people"},
				"description": {"friend of Rivest, Shamir and Adleman"},
				"password":    {"{SSHA}U3waGJVC7MgXYc0YQe7xv7sSePuTP8zN"},
			}),
		)
		entry.AddAttribute("email", []string{"alice@example.org"})
		w.Write(entry)
		resp.SetResultCode(gldap.ResultSuccess)
	}
	if m.BaseDN == "ou=people,cn=example,dc=org" {
		entry := r.NewSearchResponseEntry(
			"ou=people,cn=example,dc=org",
			gldap.WithAttributes(map[string][]string{
				"objectclass": {"organizationalUnit"},
				"ou":          {"people"},
			}),
		)
		w.Write(entry)
		resp.SetResultCode(gldap.ResultSuccess)
	}
	return
}
