package oauth

import (
	"errors"
	"oauth-server-lite/g"
	"oauth-server-lite/models/utils"
	"time"
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

func GetClientByAppId(appId int64) (client OauthClient) {
	db := g.ConnectDB()
	db.Where("app_id = ?", appId).First(&client)
	return
}

func GetClientByAppIdDel(appId int64) (client OauthClient) {
	db := g.ConnectDB()
	db.Unscoped().Where("app_id = ? and deleted_at is not null", appId).First(&client)
	return
}

func UpdateAppInfo(client OauthClient) (err error) {
	db := g.ConnectDB()
	err = db.Model(&client).Select("app_name", "description").Updates(&client).Error
	return
}

func UpdateAppInfoDel(client OauthClient) (err error) {
	db := g.ConnectDB()
	err = db.Unscoped().Model(&client).Update("deleted_at", nil).Error
	return
}

func GetExceedClients(ts time.Time) (clients []OauthClient) {
	db := g.ConnectDB()
	db.Where("updated_at < ?", ts).Find(&clients)
	return
}

func GetClients() (clients []OauthClient) {
	db := g.ConnectDB()
	print("db:", db)
	db.Find(&clients)
	return
}

func ResetClientSecret(client OauthClient) (ClientSecret string, err error) {
	db := g.ConnectDB()
	//随机字符串，client_secret 32位
	ClientSecret, err = utils.RandHashString(g.SALT, 32)
	if err != nil {
		return
	}
	err = db.Model(&client).Update("client_secret", ClientSecret).Error
	return
}

func UpdateClientDevelop(client OauthClient, columns map[string]interface{}) (err error) {
	db := g.ConnectDB()
	err = db.Model(&client).Updates(columns).Error
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

func GenerateAuthorizationCodeClient(description, domain string) (client OauthClient, err error) {
	ClientID, ClientSecret, err := GenerateClient()
	if err != nil {
		return
	}

	client = OauthClient{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		GrantTypes:   "authorization_code",
		Domains:      domain,
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

func CheckDomainValid(domain string) bool {
	var clients []OauthClient
	db := g.ConnectDB()
	db.Where("grant_types like ?", "%authorization_code%").Find(&clients)
	for _, client := range clients {
		if utils.InStrings(domain, client.Domains, ",") {
			return true
		}
	}
	return false
}
