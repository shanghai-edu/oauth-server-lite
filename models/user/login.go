package user

import (
	"strings"

	"oauth-server-lite/g"
	"oauth-server-lite/models/ldap"
	"oauth-server-lite/models/oauth"
)

func LdapLogin(username, password string) (err error) {
	lc := ldap.LDAP_CONFIG{
		Addr:       g.Config().LDAP.Addr,
		BaseDn:     g.Config().LDAP.BaseDn,
		BindDn:     g.Config().LDAP.BindDn,
		BindPass:   g.Config().LDAP.BindPass,
		AuthFilter: g.Config().LDAP.AuthFilter,
		Attributes: g.Config().LDAP.Attributes,
		TLS:        g.Config().LDAP.TLS,
		StartTLS:   g.Config().LDAP.StartTLS,
	}
	err = lc.Connect()
	defer lc.Close()
	if err != nil {
		return
	}
	_, err = lc.Auth(username, password)
	return
}

func getLdapAttr(username string) (result ldap.LDAP_RESULT, err error) {
	lc := ldap.LDAP_CONFIG{
		Addr:       g.Config().LDAP.Addr,
		BaseDn:     g.Config().LDAP.BaseDn,
		BindDn:     g.Config().LDAP.BindDn,
		BindPass:   g.Config().LDAP.BindPass,
		AuthFilter: g.Config().LDAP.AuthFilter,
		Attributes: g.Config().LDAP.Attributes,
		TLS:        g.Config().LDAP.TLS,
		StartTLS:   g.Config().LDAP.StartTLS,
	}
	err = lc.Connect()
	defer lc.Close()
	if err != nil {
		return
	}
	result, err = lc.Search_User(username)
	return
}

func GetAttrByAccessToken(accessToken string) (attributes map[string]string, err error) {
	token, err := oauth.GetAccessToken(accessToken)
	if err != nil {
		return
	}
	result, err := getLdapAttr(token.UserID)

	if err != nil {
		return
	}
	attributes = make(map[string]string)
	attributes[g.SUB] = token.UserID
	for k, v := range result.Attributes {
		attributes[k] = strings.Join(v, ";")
	}

	return
}
