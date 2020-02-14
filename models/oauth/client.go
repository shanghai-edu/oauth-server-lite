package oauth

import (
	"encoding/json"
	"errors"

	"oauth-server-lite/g"
	"oauth-server-lite/models/utils"
)

func CreateClient(client OauthClient) error {
	db := g.ConnectDB()
	err := db.Create(&client).Error
	return err
}

func UpdateClient(client OauthClient) error {
	db := g.ConnectDB()
	err := db.Save(&client).Error
	return err
}

func DeleteClient(client OauthClient) error {
	db := g.ConnectDB()
	err := db.Delete(&client).Error
	return err
}

func GetClientByClientID(clientID string) (client OauthClient) {
	db := g.ConnectDB()
	db.Where("client_id = ?", clientID).First(&client)
	return
}

func GetClients() (clients []OauthClient) {
	db := g.ConnectDB()
	db.Find(&clients)
	return
}

func GenerateClient() (ClientID, ClientSecret string, err error) {
	//随机字符串，client_id 16位
	ClientID, err = utils.RandHashString(g.SALT, 16)
	if err != nil {
		return
	}
	//随机字符串，client_secret 32位
	ClientSecret, err = utils.RandHashString(g.SALT, 32)
	if err != nil {
		return
	}
	return
}

func GenerateClientCredentialsClient(description string, whiteIPArray []string) (client OauthClient, err error) {
	ClientID, ClientSecret, err := GenerateClient()
	if err != nil {
		return
	}
	if len(whiteIPArray) == 0 {
		err = errors.New("must have at least one white ip")
		return
	}
	bs, err := json.Marshal(whiteIPArray)
	if err != nil {
		return
	}
	client = OauthClient{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		GrantType:    "client_credentials",
		Description:  description,
		WhiteIP:      string(bs),
		Scope:        "Advance",
	}
	err = CreateClient(client)
	return
}

func GenerateAuthorizationCodeClient(description, domain string) (client OauthClient, err error) {
	ClientID, ClientSecret, err := GenerateClient()
	if err != nil {
		return
	}

	client = OauthClient{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		GrantType:    "authorization_code",
		Domain:       domain,
		Description:  description,
		Scope:        "Basic",
	}
	err = CreateClient(client)
	return
}

func CheckClientPass(clientID, clientSecret string) (oauthClient OauthClient, err error) {
	oauthClient = GetClientByClientID(clientID)
	if oauthClient.ID == 0 {
		err = errors.New("cannot found such client id")
		return
	}
	if oauthClient.ClientSecret != clientSecret {
		err = errors.New("app secret is not correct")
		return
	}
	return
}

func CheckClientIP(clientIP string, clientID string) (oauthClient OauthClient, err error) {
	oauthClient = GetClientByClientID(clientID)
	if oauthClient.ID == 0 {
		err = errors.New("cannot found such client id")
		return
	}
	if oauthClient.WhiteIP == "" {
		err = errors.New("client ip is not in white ip list")
		return
	}
	var whiteIPArray []string
	err = json.Unmarshal([]byte(oauthClient.WhiteIP), &whiteIPArray)
	if err != nil {
		return
	}
	if !utils.IPCheck(clientIP, whiteIPArray) {
		err = errors.New("client ip is not in white ip list")
		return
	}
	return
}
