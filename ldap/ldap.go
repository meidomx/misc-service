package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"

	"github.com/meidomx/misc-service/config"
	"github.com/meidomx/misc-service/id"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/jimlambrt/gldap"
)

type ldapServer struct {
	IdGen *id.IdGen

	BindBaseDN string

	serverTlsConfig *tls.Config
}

func StartService(idGen *id.IdGen, c *config.Config, container *config.Container) {
	s, err := gldap.NewServer()
	if err != nil {
		log.Fatalf("unable to create server: %s", err.Error())
	}

	server := new(ldapServer)
	server.IdGen = idGen
	server.BindBaseDN = c.LDAP.BindBaseDN
	if err := InitBaseDN(c, idGen); err != nil {
		log.Fatalf("unable to cinit base dn: %s", err.Error())
	}

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
	if err := r.Unbind(server.Unbind); err != nil {
		log.Fatalf("bind op error: %s", err.Error())
	}
	if err := r.Modify(server.Modify, gldap.WithLabel("Modify")); err != nil {
		log.Fatalf("bind op error: %s", err.Error())
	}
	if c.LDAP.TLS.Enable {
		if err := r.ExtendedOperation(server.ExtendedOperationStartTLS, gldap.ExtendedOperationStartTLS); err != nil {
			log.Fatalf("bind ExtendedOperationStartTLS op error: %s", err.Error())
		}
	}
	if err := s.Router(r); err != nil {
		log.Fatalf("router error: %s", err.Error())
	}

	var connOpts []gldap.Option
	if c.LDAP.TLS.Enable {
		serverCert, err := tls.LoadX509KeyPair(c.LDAP.TLS.CertPath, c.LDAP.TLS.KeyPath)
		if err != nil {
			log.Fatalf("prepare server cert error: %s", err.Error())
		}
		server.serverTlsConfig = &tls.Config{
			Certificates: []tls.Certificate{serverCert},
		}
		connOpts = append(connOpts, gldap.WithTLSConfig(server.serverTlsConfig))
	}

	go func() {
		fmt.Println("start ldap on:", c.LDAP.Address)
		if err := s.Run(c.LDAP.Address); err != nil {
			log.Fatalf("run ldap error: %s", err.Error())
		}
	}()
	if c.LDAP.TLS.Enable {
		go func() {
			fmt.Println("start tls ldap on:", c.LDAP.TLS.TLSAddress)
			if err := s.Run(c.LDAP.TLS.TLSAddress, connOpts...); err != nil {
				log.Fatalf("run tls ldap error: %s", err.Error())
			}
		}()
	}

	container.GldapServer = s
}

func convertLDAPStringToNormal(ldapstrings []string) ([]string, error) {
	n := make([]string, len(ldapstrings))
	for i, v := range ldapstrings {
		data := []byte(v)
		// convert to normal
		// see comments in github.com/go-asn1-ber/asn1-ber@v1.5.4/ber.go func -> readPacket(reader io.Reader) (*Packet, int, error)
		if ber.Tag(data[0]) == ber.TagOctetString {
			_, s, err := readLength(data[1:])
			if err != nil {
				return nil, err
			}
			n[i] = string(data[(1 + s):])
		} else {
			n[i] = v
		}
	}
	return n, nil
}

// from github.com/go-asn1-ber/asn1-ber@v1.5.4/length.go
func readLength(bytes []byte) (length int, read int, err error) {
	// length byte
	b := bytes[0]
	read++

	switch {
	case b == 0xFF:
		// Invalid 0xFF (x.600, 8.1.3.5.c)
		return 0, read, errors.New("invalid length byte 0xff")

	case b == ber.LengthLongFormBitmask:
		// Indefinite form, we have to decode packets until we encounter an EOC packet (x.600, 8.1.3.6)
		length = ber.LengthIndefinite

	case b&ber.LengthLongFormBitmask == 0:
		// Short definite form, extract the length from the bottom 7 bits (x.600, 8.1.3.4)
		length = int(b) & ber.LengthValueBitmask

	case b&ber.LengthLongFormBitmask != 0:
		// Long definite form, extract the number of length bytes to follow from the bottom 7 bits (x.600, 8.1.3.5.b)
		lengthBytes := int(b) & ber.LengthValueBitmask
		// Protect against overflow
		// TODO: support big int length?
		if lengthBytes > 8 {
			return 0, read, errors.New("long-form length overflow")
		}

		// Accumulate into a 64-bit variable
		var length64 int64
		for i := 0; i < lengthBytes; i++ {
			b = bytes[read]
			read++

			// x.600, 8.1.3.5
			length64 <<= 8
			length64 |= int64(b)
		}

		// Cast to a platform-specific integer
		length = int(length64)
		// Ensure we didn't overflow
		if int64(length) != length64 {
			return 0, read, errors.New("long-form length overflow")
		}

	default:
		return 0, read, errors.New("invalid length byte")
	}

	return length, read, nil
}

func (this *ldapServer) Modify(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleModify"
	log.Println("operation:", op)

	res := r.NewModifyResponse(gldap.WithResponseCode(gldap.ResultNoSuchObject))
	defer w.Write(res)
	m, err := r.GetModifyMessage()
	if err != nil {
		log.Println("not a modify message: %s", "op", op, "err", err)
		return
	}
	log.Println("modify request", "dn", m.DN)

	entry, err := FindOneEntry(m.DN)
	if err != nil {
		log.Println("FindOneEntry error", "op", op, "err", err)
		return
	}
	if len(entry.DN) <= 0 {
		log.Println("FindOneEntry empty", "op", op, "err", err)
		return
	}

	if entry.Attributes == nil {
		entry.Attributes = []*gldap.EntryAttribute{}
	}
	res.SetMatchedDN(entry.DN)
	for _, chg := range m.Changes {
		// find specific attr
		var foundAttr *gldap.EntryAttribute
		var foundAt int
		for i, a := range entry.Attributes {
			if a.Name == chg.Modification.Type {
				foundAttr = a
				foundAt = i
			}
		}

		converted, err := convertLDAPStringToNormal(chg.Modification.Vals)
		if err != nil {
			log.Println("convertLDAPStringToNormal error", "op", op, "err", err)
			return
		}
		// then apply operation
		switch chg.Operation {
		case gldap.AddAttribute:
			if foundAttr != nil {
				foundAttr.AddValue(converted...)
			} else {
				entry.Attributes = append(entry.Attributes, gldap.NewEntryAttribute(chg.Modification.Type, converted))
			}
		case gldap.DeleteAttribute:
			if foundAttr != nil {
				// slice out the deleted attribute
				copy(entry.Attributes[foundAt:], entry.Attributes[foundAt+1:])
				entry.Attributes = entry.Attributes[:len(entry.Attributes)-1]
			}
		case gldap.ReplaceAttribute:
			if foundAttr != nil {
				*foundAttr = *gldap.NewEntryAttribute(chg.Modification.Type, converted)
			}
		}
	}

	if err := UpdateEntry(entry); err != nil {
		log.Println("UpdateEntry error", "op", op, "err", err)
		return
	}

	res.SetResultCode(gldap.ResultSuccess)
}

func (this *ldapServer) Unbind(w *gldap.ResponseWriter, r *gldap.Request) {
	const op = "ldap.(Directory).handleUnbind"
	log.Println("operation:", op)
}

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
	userDN := fmt.Sprint("cn=", m.UserName, ",", this.BindBaseDN)
	//TODO need optimize the query
	entry, err := FindOneEntry(userDN)
	if err != nil {
		log.Println("find children error", "op", op, "err", err)
		return
	}
	if len(entry.DN) > 0 {
		log.Println("found bind user", "op", op, "DN", userDN)
		values := entry.GetAttributeValues("userPassword")
		if len(values) > 0 && string(m.Password) == values[0] {
			resp.SetResultCode(gldap.ResultSuccess)
			return
		}
	}

	// user is full DN
	entry, err = FindOneEntry(m.UserName)
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
