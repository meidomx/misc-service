package config

type Config struct {
	Storage struct {
		ConnString string `toml:"conn_string"`
	} `toml:"storage"`

	LDAP struct {
		Address    string `toml:"address"`
		BindBaseDN string `toml:"bind_base_dn"`

		Init struct {
			InitScripts []struct {
				DN            string   `toml:"dn"`
				ObjectClasses []string `toml:"object_classes"`
			} `toml:"run_simple_init_scripts"`
			InitAdmin struct {
				DN            string   `toml:"dn"`
				ObjectClasses []string `toml:"object_classes"`
				UserPassword  string   `toml:"user_password"`
			} `toml:"run_init_admin"`
		} `toml:"init"`
	} `toml:"ldap"`
}
