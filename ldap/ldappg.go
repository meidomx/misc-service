package ldap

import (
	"context"

	"github.com/meidomx/misc-service/id"
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

const (
	ServiceName = "ldap"
)

func FindEntry(dn string) (*gldap.Entry, error) {
	entryPath := SplitDN(dn)
	entry := new(gldap.Entry)
	r, err := pgbackend.RunQuery(ServiceName, entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		parent := CombineParentDN(entryPath)
		if len(parent) != 0 {
			rows, err := conn.Query(context.Background(),
				"select attribute from misc_ldap_entries where entry_name = $1 and parent_full_entry_path = $2",
				entryPath[0], parent)
			if err != nil {
				return err
			}
			defer rows.Close()

			if rows.Next() {
				if err := rows.Scan(result); err != nil {
					return err
				}
			}

			return nil
		} else {
			rows, err := conn.Query(context.Background(),
				"select attribute from misc_ldap_entries where entry_name = $1 and parent_full_entry_path is NULL",
				entryPath[0])
			if err != nil {
				return err
			}
			defer rows.Close()

			if rows.Next() {
				if err := rows.Scan(result); err != nil {
					return err
				}
			}

			return nil
		}
	})

	return r, err
}

func SaveEntry(entry *gldap.Entry, i id.ItemId) error {
	_, err := pgbackend.RunQuery(ServiceName, entry, func(conn *pgxpool.Conn, result *gldap.Entry) error {
		sp := SplitDN(entry.DN)
		entryType := EntryType(sp[0])
		parent := CombineParentDN(sp)
		if len(parent) > 0 {
			_, err := conn.Exec(context.Background(),
				"insert into misc_ldap_entries(entry_id, entry_name, parent_full_entry_path, entry_type, attribute, metadata) values($1, $2, $3, $4, $5, $6)",
				i.HexString(), sp[0], parent, entryType, result, nil)
			return err
		} else {
			_, err := conn.Exec(context.Background(),
				"insert into misc_ldap_entries(entry_id, entry_name, entry_type, attribute, metadata) values($1, $2, $3, $4, $5)",
				i.HexString(), sp[0], entryType, result, nil)
			return err
		}
	})
	return err
}
