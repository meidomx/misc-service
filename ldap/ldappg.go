package ldap

import (
	"github.com/meidomx/misc-service/pgbackend"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/jimlambrt/gldap"
)

const (
	EntryTypeDomainComponent  = "dc"
	EntryTypeCountry          = "c"
	EntryTypeOrganization     = "o"
	EntryTypeOrganizationUnit = "ou"
	EntryTypeSurname          = "sn"
	EntryTypeCommonName       = "cn"
)

func FindEntry(entryPath []string) (*gldap.Entry, error) {
	entry := new(gldap.Entry)
	r, err := pgbackend.RunQuery("ldap", entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		return nil
	})

	return r, err
}
