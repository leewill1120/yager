package rtslib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"

	"leewill1120/yager/utils"
)

type Fabric_module struct {
}

type Storage_object struct {
	Attributes struct {
		Block_size                  int `json:"block_size"`
		Emulate_3pc                 int `json:"emulate_3pc"`
		Emulate_caw                 int `json:"emulate_caw"`
		Emulate_dpo                 int `json:"emulate_dpo"`
		Emulate_fua_read            int `json:"emulate_fua_read"`
		Emulate_fua_write           int `json:"emulate_fua_write"`
		Emulate_model_alias         int `json:"emulate_model_alias"`
		Emulate_rest_reord          int `json:"emulate_rest_reord"`
		Emulate_tas                 int `json:"emulate_tas"`
		Emulate_tpu                 int `json:"emulate_tpu"`
		Emulate_tpws                int `json:"emulate_tpws"`
		Emulate_ua_intlck_ctrl      int `json:"emulate_ua_intlck_ctrl"`
		Emulate_write_cache         int `json:"emulate_write_cache"`
		Enforce_pr_isids            int `json:"enforce_pr_isids"`
		Force_pr_aptpl              int `json:"force_pr_aptpl"`
		Is_nonrot                   int `json:"is_nonrot"`
		Max_unmap_block_desc_count  int `json:"max_unmap_block_desc_count"`
		Max_unmap_lba_count         int `json:"max_unmap_lba_count"`
		Max_write_same_len          int `json:"max_write_same_len"`
		Optimal_sectors             int `json:"optimal_sectors"`
		Pi_prot_format              int `json:"pi_prot_format"`
		Pi_prot_type                int `json:"pi_prot_type"`
		Queue_depth                 int `json:"queue_depth"`
		Unmap_granularity           int `json:"unmap_granularity"`
		Unmap_granularity_alignment int `json:"unmap_granularity_alignment"`
	} `json:"attributes"`
	Dev        string `json:"dev"`
	Name       string `json:"name"`
	Plugin     string `json:"plugin"`
	Readonly   bool   `json:"readonly,omitempty"`
	Size       int    `json:"size,omitempty"`
	Write_back bool   `json:"write_back"`
	Wwn        string `json:"wwn"`
}

type LUN struct {
	Index          int    `json:"index"`
	Storage_object string `json:"storage_object"`
}

type MappedLUN struct {
	Index         int  `json:"index"`
	Tpg_lun       int  `json:"tpg_lun"`
	Write_protect bool `json:"write_protect"`
}
type ACL struct {
	Attributes struct {
		Dataout_timeout           int `json:"dataout_timeout"`
		Dataout_timeout_retries   int `json:"dataout_timeout_retries"`
		Default_erl               int `json:"default_erl"`
		Nopin_response_timeout    int `json:"nopin_response_timeout"`
		Nopin_timeout             int `json:"nopin_timeout"`
		Random_datain_pdu_offsets int `json:"random_datain_pdu_offsets"`
		Random_datain_seq_offsets int `json:"random_datain_seq_offsets"`
		Random_r2t_offsets        int `json:"random_r2t_offsets"`
	} `json:"attributes"`
	Chap_userid   string      `json:"chap_userid"`
	Chap_password string      `json:"chap_password"`
	Mapped_luns   []MappedLUN `json:"mapped_luns"`
	Node_wwn      string      `json:"node_wwn"`
}

type Portal struct {
	Ip_address string `json:"ip_address"`
	Iser       bool   `json:"iser"`
	Port       int    `json:"port"`
}

type TPG struct {
	Attributes struct {
		Authentication          int `json:"authentication"`
		Cache_dynamic_acls      int `json:"cache_dynamic_acls"`
		Default_cmdsn_depth     int `json:"default_cmdsn_depth"`
		Default_erl             int `json:"default_erl"`
		Demo_mode_discovery     int `json:"demo_mode_discovery"`
		Demo_mode_write_protect int `json:"demo_mode_write_protect"`
		Generate_node_acls      int `json:"generate_node_acls"`
		Login_timeout           int `json:"login_timeout"`
		Netif_timeout           int `json:"netif_timeout"`
		Prod_mode_write_protect int `json:"prod_mode_write_protect"`
		T10_pi                  int `json:"t10_pi"`
	} `json:"attributes"`
	Enable     bool  `json:"enable"`
	Luns       []LUN `json:"luns"`
	Node_acls  []ACL `json:"node_acls"`
	Parameters struct {
		AuthMethod               string `json:"AuthMethod"`
		DataDigest               string `json:"DataDigest"`
		DataPDUInOrder           string `json:"DataPDUInOrder"`
		DataSequenceInOrder      string `json:"DataSequenceInOrder"`
		DefaultTime2Retain       string `json:"DefaultTime2Retain"`
		DefaultTime2Wait         string `json:"DefaultTime2Wait"`
		ErrorRecoveryLevel       string `json:"ErrorRecoveryLevel"`
		FirstBurstLength         string `json:"FirstBurstLength"`
		HeaderDigest             string `json:"HeaderDigest"`
		IFMarkInt                string `json:"IFMarkInt"`
		IFMarker                 string `json:"IFMarker"`
		ImmediateData            string `json:"ImmediateData"`
		InitialR2T               string `json:"InitialR2T"`
		MaxBurstLength           string `json:"MaxBurstLength"`
		MaxConnections           string `json:"MaxConnections"`
		MaxOutstandingR2T        string `json:"MaxOutstandingR2T"`
		MaxRecvDataSegmentLength string `json:"MaxRecvDataSegmentLength"`
		MaxXmitDataSegmentLength string `json:"MaxXmitDataSegmentLength"`
		OFMarkInt                string `json:"OFMarkInt"`
		OFMarker                 string `json:"OFMarker"`
		TargetAlias              string `json:"TargetAlias"`
	} `json:"parameters"`
	Portals []Portal `json:"portals"`
	Tag     int      `json:"tag"`
}

type Target struct {
	Fabric string `json:"fabric"`
	Tpgs   []TPG  `json:"tpgs"`
	Wwn    string `json:"wwn"`
}

type Config struct {
	Fabric_modules  []Fabric_module  `json:"fabric_modules"`
	Storage_objects []Storage_object `json:"storage_objects"`
	Targets         []Target         `json:"targets"`
}

func NewConfig() *Config {
	c := &Config{}
	if err := c.FromDisk(""); err != nil {
		log.Println(err)
		return nil
	} else {
		return c
	}
}

func (c *Config) Print() {
	if data, err := json.MarshalIndent(c, "", " "); err != nil {
		log.Println(err)
	} else {
		fmt.Println(string(data))
	}
}

func (c *Config) FromDisk(filePath string) error {
	if "" == filePath {
		filePath = defaultConfig
	}

	if data, err := ioutil.ReadFile(filePath); err != nil {
		return err
	} else {
		return json.Unmarshal(data, c)
	}
}

func (c *Config) ToDisk(filePath string) error {
	if "" == filePath {
		filePath = defaultConfig
	}

	if data, err := json.Marshal(c); err != nil {
		return err
	} else {
		return ioutil.WriteFile(filePath, data, 0755)
	}
}

func (c *Config) Restore(filePath string) error {
	if "" == filePath {
		filePath = defaultConfig
	}

	cmd := exec.Command("targetctl", "restore")
	return cmd.Run()
}

func (c *Config) AddBlockStore(dev, name string) error {
	for _, s := range c.Storage_objects {
		if s.Dev == dev || s.Name == name {
			return fmt.Errorf("dev or name already exists!")
		}
	}

	wwn, err := utils.Generate_wwn("unit_serail")
	if err != nil {
		return err
	}

	so := Storage_object{
		Dev:        dev,
		Name:       name,
		Plugin:     "block",
		Readonly:   false,
		Write_back: false,
		Wwn:        wwn,
	}
	so.Attributes.Block_size = 512
	so.Attributes.Emulate_3pc = 1
	so.Attributes.Emulate_caw = 1
	so.Attributes.Emulate_dpo = 0
	so.Attributes.Emulate_fua_read = 0
	so.Attributes.Emulate_fua_write = 1
	so.Attributes.Emulate_model_alias = 1
	so.Attributes.Emulate_rest_reord = 1
	so.Attributes.Emulate_tas = 1
	so.Attributes.Emulate_tpu = 0
	so.Attributes.Emulate_tpws = 0
	so.Attributes.Emulate_ua_intlck_ctrl = 0
	so.Attributes.Emulate_write_cache = 0
	so.Attributes.Enforce_pr_isids = 1
	so.Attributes.Force_pr_aptpl = 0
	so.Attributes.Is_nonrot = 0
	so.Attributes.Max_unmap_block_desc_count = 0
	so.Attributes.Max_unmap_lba_count = 0
	so.Attributes.Max_write_same_len = 65535
	so.Attributes.Optimal_sectors = 8192
	so.Attributes.Pi_prot_format = 0
	so.Attributes.Pi_prot_type = 0
	so.Attributes.Queue_depth = 128
	so.Attributes.Unmap_granularity = 0
	so.Attributes.Unmap_granularity_alignment = 0

	c.Storage_objects = append(c.Storage_objects, so)
	return nil
}

func (c *Config) GetStore(name string) *Storage_object {
	for _, s := range c.Storage_objects {
		if s.Name == name {
			return &s
		}
	}

	log.Printf("Target(%s) doesn't exist.", name)
	return nil
}

func (c *Config) RemoveStore(name string) error {
	for index, s := range c.Storage_objects {
		if s.Name == name {
			c.Storage_objects = append(c.Storage_objects[:index], c.Storage_objects[index+1:]...)
			return nil
		}
	}

	log.Printf("Storege_object(%s) doesn't exist.", name)
	return nil
}

func (c *Config) AddTarget(storage_object, node_wwn, userid, password string) (string, error) {
	wwn, err := utils.Generate_wwn("iqn")
	if err != nil {
		return "", err
	}

	for _, t := range c.Targets {
		if t.Wwn == wwn {
			return "", fmt.Errorf("This Target already exists in configFS")
		}
	}

	luns := make([]LUN, 1)
	luns[0].Index = 0
	luns[0].Storage_object = storage_object

	portals := make([]Portal, 1)
	portals[0].Ip_address = "0.0.0.0"
	portals[0].Iser = false
	portals[0].Port = 3260

	mapped_luns := make([]MappedLUN, 1)
	mapped_luns[0].Index = 0
	mapped_luns[0].Tpg_lun = 0
	mapped_luns[0].Write_protect = false

	node_acls := make([]ACL, 1)
	node_acls[0].Chap_password = password
	node_acls[0].Chap_userid = userid
	node_acls[0].Node_wwn = node_wwn
	node_acls[0].Mapped_luns = mapped_luns
	node_acls[0].Attributes.Dataout_timeout = 3
	node_acls[0].Attributes.Dataout_timeout_retries = 5
	node_acls[0].Attributes.Default_erl = 0
	node_acls[0].Attributes.Nopin_response_timeout = 30
	node_acls[0].Attributes.Nopin_timeout = 15
	node_acls[0].Attributes.Random_datain_pdu_offsets = 0
	node_acls[0].Attributes.Random_datain_seq_offsets = 0
	node_acls[0].Attributes.Random_r2t_offsets = 0

	tpgs := make([]TPG, 1)
	tpgs[0].Attributes.Authentication = 0
	tpgs[0].Attributes.Cache_dynamic_acls = 0
	tpgs[0].Attributes.Default_cmdsn_depth = 64
	tpgs[0].Attributes.Default_erl = 0
	tpgs[0].Attributes.Demo_mode_discovery = 1
	tpgs[0].Attributes.Demo_mode_write_protect = 1
	tpgs[0].Attributes.Generate_node_acls = 0
	tpgs[0].Attributes.Login_timeout = 15
	tpgs[0].Attributes.Netif_timeout = 2
	tpgs[0].Attributes.Prod_mode_write_protect = 0
	tpgs[0].Attributes.T10_pi = 0
	tpgs[0].Parameters.AuthMethod = "CHAP,None"
	tpgs[0].Parameters.DataDigest = "CRC32C,None"
	tpgs[0].Parameters.DataPDUInOrder = "Yes"
	tpgs[0].Parameters.DataSequenceInOrder = "Yes"
	tpgs[0].Parameters.DefaultTime2Retain = "20"
	tpgs[0].Parameters.DefaultTime2Wait = "2"
	tpgs[0].Parameters.ErrorRecoveryLevel = "0"
	tpgs[0].Parameters.FirstBurstLength = "65536"
	tpgs[0].Parameters.HeaderDigest = "CRC32C,None"
	tpgs[0].Parameters.IFMarkInt = "2048~65535"
	tpgs[0].Parameters.IFMarker = "No"
	tpgs[0].Parameters.ImmediateData = "Yes"
	tpgs[0].Parameters.InitialR2T = "Yes"
	tpgs[0].Parameters.MaxBurstLength = "262144"
	tpgs[0].Parameters.MaxConnections = "1"
	tpgs[0].Parameters.MaxOutstandingR2T = "1"
	tpgs[0].Parameters.MaxRecvDataSegmentLength = "8192"
	tpgs[0].Parameters.MaxXmitDataSegmentLength = "262144"
	tpgs[0].Parameters.OFMarkInt = "2048~65535"
	tpgs[0].Parameters.OFMarker = "No"
	tpgs[0].Parameters.TargetAlias = "LIO Target"

	tpgs[0].Luns = luns
	tpgs[0].Portals = portals
	tpgs[0].Node_acls = node_acls
	tpgs[0].Tag = 1
	tpgs[0].Enable = true

	target := Target{
		Fabric: "iscsi",
		Tpgs:   tpgs,
		Wwn:    wwn,
	}

	c.Targets = append(c.Targets, target)

	return wwn, nil
}

func (c *Config) GetTarget(wwn string) *Target {
	for _, t := range c.Targets {
		if t.Wwn == wwn {
			return &t
		}
	}
	log.Printf("Target(%s) doesn't exist.", wwn)
	return nil
}

func (c *Config) RemoveTarget(wwn string) error {
	for index, t := range c.Targets {
		if t.Wwn == wwn {
			c.Targets = append(c.Targets[:index], c.Targets[index+1:]...)
			return nil
		}
	}
	log.Printf("Target(%s) doesn't exist.", wwn)
	return nil
}
