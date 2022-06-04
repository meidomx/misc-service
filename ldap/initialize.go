package ldap

import (
	"log"

	"github.com/meidomx/misc-service/config"
	"github.com/meidomx/misc-service/id"

	"github.com/jimlambrt/gldap"
)

func InitBaseDN(c *config.Config, idGen *id.IdGen) error {
	// skip if initialized
	if r, err := FindAllRoots(); err != nil {
		log.Println("InitBaseDN - query roots error:", err)
		return err
	} else if len(r) > 0 {
		log.Println("ldap is already initialized")
		return nil
	}

	// run dn init
	for _, v := range c.LDAP.Init.InitScripts {
		if err := runInit(v, idGen); err != nil {
			log.Println("InitBaseDN - run init error:", err)
			return err
		}
	}

	// run admin init
	if err := runAdminInit(c.LDAP.Init.InitAdmin, idGen); err != nil {
		log.Println("InitBaseDN - run admin init error:", err)
		return err
	}

	log.Println("ldap has successfully initialized")

	return nil
}

func runAdminInit(admin struct {
	DN            string   `toml:"dn"`
	ObjectClasses []string `toml:"object_classes"`
	UserPassword  string   `toml:"user_password"`
}, idGen *id.IdGen) error {

	attrs := map[string][]string{}
	attrs["objectClass"] = admin.ObjectClasses
	attrs["userPassword"] = []string{admin.UserPassword}

	newEntry := gldap.NewEntry(admin.DN, attrs)

	newId, err := idGen.Next()
	if err != nil {
		log.Println("generate id error:", err)
		return err
	}
	err = SaveEntry(newEntry, newId)
	if err != nil {
		log.Println("SaveEntry error:", err)
		return err
	}

	return nil
}

func runInit(v struct {
	DN            string   `toml:"dn"`
	ObjectClasses []string `toml:"object_classes"`
}, idGen *id.IdGen) error {

	attrs := map[string][]string{}
	attrs["objectClass"] = v.ObjectClasses

	newEntry := gldap.NewEntry(v.DN, attrs)

	newId, err := idGen.Next()
	if err != nil {
		log.Println("generate id error:", err)
		return err
	}
	err = SaveEntry(newEntry, newId)
	if err != nil {
		log.Println("SaveEntry error:", err)
		return err
	}

	return nil
}
