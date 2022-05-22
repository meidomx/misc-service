package ldap

import (
	"context"
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
