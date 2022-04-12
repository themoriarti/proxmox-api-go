package proxmox

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

// Acme Plugin options for the Proxmox API
type ConfigAcmePlugin struct {
	ID              string   `json:"pluginid"`
	API             string   `json:"api"`
	Data            string   `json:"data,omitempty"`
	Enable          bool     `json:"enable"`
	Nodes           []string `json:"nodes,omitempty"`
	ValidationDelay int      `json:"validation-delay"`
}

func (config ConfigAcmePlugin) MapAcmePluginValues()(params map[string]interface{}) {
	params = map[string]interface{}{
		"api":              config.API,
		"data":             base64.StdEncoding.EncodeToString([]byte(config.Data)),
		"disable":          BoolInvert(config.Enable),
		"nodes":            ArrayToCSV(config.Nodes),
		"validation-delay": config.ValidationDelay,
	}
	return
}

func (config ConfigAcmePlugin) SetAcmePlugin(pluginid string, client *Client) (err error) {
	err = ValidateIntInRange(0, 172800, config.ValidationDelay, "validation-delay")
	if err != nil {return}

	config.ID = pluginid

	pluginExists, err := client.CheckAcmePluginExistance(pluginid)
	if err != nil {return}

	if pluginExists == true {
		err = config.UpdateAcmePlugin(client)
	} else {
		err = config.CreateAcmePlugin(client)
	}
	return
}

func (config ConfigAcmePlugin) CreateAcmePlugin(client *Client) (err error) {
	params := config.MapAcmePluginValues()
	params["id"] = config.ID
	params["type"] = "dns"
	exitStatus, err := client.CreateAcmePlugin(params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error creating Acme plugin: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}
	return
}

func (config ConfigAcmePlugin) UpdateAcmePlugin(client *Client) (err error) {
	params := config.MapAcmePluginValues()
	exitStatus, err := client.UpdateAcmePlugin(config.ID, params)
	if err != nil {
		params, _ := json.Marshal(&params)
		return fmt.Errorf("error updating Acme plugin: %v, error status: %s (params: %v)", err, exitStatus, string(params))
	}
	return
}

func NewConfigAcmePluginFromApi(id string, client *Client) (config *ConfigAcmePlugin, err error) {
	// prepare json map to receive the information from the api
	var rawConfig map[string]interface{}
	rawConfig, err = client.GetAcmePluginConfig(id)
	if err != nil {return nil, err}

	config = new(ConfigAcmePlugin)

	config.ID = id
	config.API = rawConfig["api"].(string)
	
	if _, isSet := rawConfig["data"]; isSet {config.Data = rawConfig["data"].(string)}
	if _, isSet := rawConfig["disable"]; isSet {config.Enable = BoolInvert(Itob(int(rawConfig["disable"].(float64))))} else {config.Enable = true}
	if _, isSet := rawConfig["validation-delay"]; isSet {config.ValidationDelay = int(rawConfig["validation-delay"].(float64))} else {config.ValidationDelay = 30}

	return
}

func NewConfigAcmePluginFromJson(io io.Reader) (config *ConfigAcmePlugin, err error) {
	config = &ConfigAcmePlugin{}
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return
}