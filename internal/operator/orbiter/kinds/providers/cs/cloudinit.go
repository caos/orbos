package cs

import "gopkg.in/yaml.v3"

type Cloudinit struct {
	Groups  map[string][]string `yaml:"groups,omitempty"`
	Users   []*CloudinitUser    `yaml:"users,omitempty"`
	RunCMDs []string            `yaml:"runcmd,omitempty"`
}

type CloudinitUser struct {
	Name              string   `yaml:"name,omitempty"`
	PlainTextPasswd   string   `yaml:"plain_text_passwd,omitempty"`
	ExpireDate        string   `yaml:"expiredate,omitempty"`
	Gecos             string   `yaml:"gecos,omitempty"`
	Homedir           string   `yaml:"homedir,omitempty"`
	PrimaryGroup      string   `yaml:"primary_group,omitempty"`
	Groups            []string `yaml:"groups,omitempty"`
	SelinuxUser       string   `yaml:"selinux_user,omitempty"`
	LockPasswd        *bool    `yaml:"lock_passwd,omitempty"`
	Inactive          *bool    `yaml:"inactive,omitempty"`
	Passwd            string   `yaml:"passwd,omitempty"`
	SSHImportID       string   `yaml:"ssh_import_id,omitempty"`
	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
	SSHRedirectUser   *bool    `yaml:"ssh_redirect_user,omitempty"`
	Sudo              string   `yaml:"sudo,omitempty"`
}

func NewCloudinit() *Cloudinit {
	return &Cloudinit{
		Groups: map[string][]string{},
		Users:  []*CloudinitUser{},
	}
}

func (c *Cloudinit) AddGroupWithoutUsers(name string) *Cloudinit {
	c.Groups[name] = []string{}
	return c
}

func (c *Cloudinit) AddUser(
	name string,
	lockPasswd bool,
	password string,
	groups []string,
	primaryGroup string,
	sshAuthorizedKeys []string,
	sudo string,
) *Cloudinit {
	c.Users = append(c.Users, &CloudinitUser{
		Name:              name,
		PlainTextPasswd:   password,
		ExpireDate:        "",
		Gecos:             "",
		Homedir:           "",
		PrimaryGroup:      primaryGroup,
		Groups:            groups,
		SelinuxUser:       "",
		LockPasswd:        &lockPasswd,
		Inactive:          nil,
		Passwd:            "",
		SSHImportID:       "",
		SSHAuthorizedKeys: sshAuthorizedKeys,
		SSHRedirectUser:   nil,
		Sudo:              sudo,
	})
	return c
}

func (c *Cloudinit) AddCmd(cmd string) *Cloudinit {
	c.RunCMDs = append(c.RunCMDs, cmd)
	return c
}

func (c *Cloudinit) ToYamlString() (string, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return "#cloud-config\n" + string(data), nil
}
