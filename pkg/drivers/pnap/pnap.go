package pnap

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/PNAP/bmc-api-sdk/client"
	"github.com/PNAP/bmc-api-sdk/command"
	"github.com/PNAP/bmc-api-sdk/dto"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
)

// Driver is the implementation of BaseDriver interface
type Driver struct {
	*drivers.BaseDriver
	client *client.PNAPClient

	//APIToken         string
	//UserAgentPrefix  string
	ID                 string
	Status             string
	Name               string
	ServerDescription  string
	PrivateIPAddresses []string
	PublicIPAddresses  []string
	ServerOs           string
	ServerType         string
	ServerLocation     string
	CPU                string
	RAM                string
	Storage            string
	ClientIdentifier   string
	ClientSecret       string
}

const (
	defaultOS       = "ubuntu/bionic"
	defaultType     = "s1.c1.medium"
	defaultLocation = "PHX"
)

// NewDriver creates and returns a new instance of the PNAP driver
func NewDriver() *Driver {
	return &Driver{
		ServerOs:       defaultOS,
		ServerType:     defaultType,
		ServerLocation: defaultLocation,

		BaseDriver: &drivers.BaseDriver{},
	}
}

// getClient creates the pnap API Client
func (d *Driver) getClient() (*client.PNAPClient, error) {
	if d.client == nil {
		var pnapClient client.PNAPClient
		var confErr error

		if (d.ClientIdentifier != "") && (d.ClientSecret != "") {
			pnapClient = client.NewPNAPClient(d.ClientIdentifier, d.ClientSecret)
		} else {
			pnapClient, confErr = client.Create()
			if confErr != nil {
				return nil, errors.Wrap(confErr, "PNAP API client can not be created")

			}
		}

		d.client = &pnapClient
	}
	return d.client, nil
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_OS",
			Name:   "pnap-server-os",
			Usage:  "The server’s OS ID used when the server was created (e.g., ubuntu/bionic, centos/centos7).",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_LOCATION",
			Name:   "pnap-server-location",
			Usage:  "Server Location ID. Cannot be changed once a server is created",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_TYPE",
			Name:   "pnap-server-type",
			Usage:  "Server type ID. Cannot be changed once a server is created",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_DESCRIPTION",
			Name:   "pnap-server-description",
			Usage:  "Server description",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_SERVER_HOSTNAME",
			Name:   "pnap-server-hostname",
			Usage:  "Server hostname",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_CLIENT_ID",
			Name:   "pnap-client-identifier",
			Usage:  "Client ID from Application Credentials",
		},
		mcnflag.StringFlag{
			EnvVar: "PNAP_CLIENT_SECRET",
			Name:   "pnap-client-secret",
			Usage:  "Client Secret from Application Credentials",
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "pnap"
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Name = flags.String("pnap-server-hostname")
	d.ServerLocation = flags.String("pnap-server-location")
	d.ServerOs = flags.String("pnap-server-os")
	d.ServerType = flags.String("pnap-server-type")
	d.ServerDescription = flags.String("pnap-server-description")
	d.ClientIdentifier = flags.String("pnap-client-identifier")
	d.ClientSecret = flags.String("pnap-client-secret")

	return nil
}
func (d *Driver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

// publicSSHKeyPath is always SSH Key Path appended with ".pub"
func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Info("Creating pnap machine instance...")
	log.Infof("Driver params host:%s clientID:%s type:%s os:%s", d.Name, d.ClientIdentifier, d.ServerType, d.ServerOs)
	publicKey, err := d.createSSHKey()
	if err != nil {
		return err
	}

	client, err := d.getClient()
	if err != nil {
		return err
	}

	request := &dto.ProvisionedServer{}
	request.Name = d.MachineName
	request.Description = d.ServerDescription
	request.Os = d.ServerOs
	request.Type = d.ServerType
	request.Location = d.ServerLocation

	request.SSHKeys = []string{strings.TrimSpace(publicKey)}

	requestCommand := command.NewCreateServerCommand(client, *request)

	resp, err := requestCommand.Execute()

	if err != nil {
		return err
	}
	code := resp.StatusCode
	if code == 200 {
		response := &dto.LongServer{}
		response.FromBytes(resp)
		d.ID = (response.ID)
		d.Name = d.MachineName
		//d.MachineName = (response.ID)
		d.PrivateIPAddresses = response.PrivateIPAddresses
		d.PublicIPAddresses = response.PublicIPAddresses
		d.IPAddress = response.PublicIPAddresses[0]
		d.RAM = response.RAM
		d.Storage = response.Storage
		d.CPU = response.CPU
	} else {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code %v Message: %s Validation Errors: %s", code, response.Message, response.ValidationErrors)
	}

	if err := d.waitForStatus(state.Running); err != nil {
		return fmt.Errorf("wait for machine running failed: %s", err)
	}

	return nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {

	client, err := d.getClient()
	if err != nil {
		return state.Error, err
	}
	requestCommand := command.NewGetServerCommand(client, d.ID)
	resp, err := requestCommand.Execute()
	if err != nil {
		return state.Error, err
	}
	code := resp.StatusCode
	if code != 200 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return state.Error, fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	}
	response := &dto.LongServer{}
	response.FromBytes(resp)
	d.ID = (response.ID)
	d.Status = response.Status

	switch d.Status {
	case "powered-on":
		return state.Running, nil
	case "creating",
		"resetting",
		"rebooting":
		return state.Starting, nil
	case "powered-off":
		return state.Stopped, nil
	}
	return state.None, nil
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}

	return fmt.Sprintf("tcp://%s:%d", ip, 2376), nil
}

func (d *Driver) waitForStatus(a state.State) error {
	for {
		log.Infof("Waiting for Machine %s...", a.String())
		act, err := d.GetState()
		if err != nil {
			return errors.Wrap(err, "Could not get Server state.")
		}

		if act == a {
			log.Infof("Created pnap machine reached state %s.", a.String())
			break
		} else if act == state.Error {
			return errors.Wrap(err, "Server state could not be retrived.")
		}

		time.Sleep(10 * time.Second)
	}
	return nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	log.Info("Killing pnap machine instance...")
	client, err := d.getClient()
	if err != nil {
		return err
	}

	var requestCommand command.Executor
	requestCommand = command.NewShutDownCommand(client, d.ID)
	err = run(requestCommand)
	if err != nil {
		return err
	}
	if err := d.waitForStatus(state.Stopped); err != nil {
		return fmt.Errorf("wait for machine stopping failed: %s", err)
	}
	return err
}

// Remove a host
func (d *Driver) Remove() error {
	log.Infof("Removing pnap machine instance with id %s", d.ID)
	if d.ID == "" {
		return nil
	}
	client, err := d.getClient()
	if err != nil {
		return err
	}

	var requestCommand command.Executor
	requestCommand = command.NewDeleteServerCommand(client, d.ID)
	resp, err := requestCommand.Execute()
	if err != nil {
		return err
	}
	code := resp.StatusCode
	if code != 200 && code != 404 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	}
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	log.Info("Rebooting pnap machine instance...")
	client, err := d.getClient()
	if err != nil {
		return err
	}

	var requestCommand command.Executor
	requestCommand = command.NewRebootCommand(client, d.ID)
	err = run(requestCommand)
	if err != nil {
		return err
	}
	if err := d.waitForStatus(state.Running); err != nil {
		return fmt.Errorf("wait for machine reboot failed: %s", err)
	}
	return err
}

// Start a host
func (d *Driver) Start() error {
	log.Info("Starting pnap machine instance...")
	client, err := d.getClient()
	if err != nil {
		return err
	}

	var requestCommand command.Executor
	requestCommand = command.NewPowerOnCommand(client, d.ID)
	err = run(requestCommand)
	if err != nil {
		return err
	}
	if err := d.waitForStatus(state.Running); err != nil {
		return fmt.Errorf("wait for machine to start failed: %s", err)
	}
	return err
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	log.Info("Stopping pnap machine instance...")
	client, err := d.getClient()
	if err != nil {
		return err
	}

	var requestCommand command.Executor
	requestCommand = command.NewShutDownCommand(client, d.ID)
	err = run(requestCommand)
	if err != nil {
		return err
	}
	if err := d.waitForStatus(state.Stopped); err != nil {
		return fmt.Errorf("wait for machine to shut down failed: %s", err)
	}
	return err
}

func run(command command.Executor) error {
	resp, err := command.Execute()
	if err != nil {
		return err
	}
	code := resp.StatusCode
	if code != 200 {
		response := &dto.ErrorMessage{}
		response.FromBytes(resp)
		return fmt.Errorf("API Returned Code: %v, Message: %v, Validation Errors: %v", code, response.Message, response.ValidationErrors)
	}
	return nil
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	if d.ServerLocation == "" {
		log.Info("Location has not been set, will be used PHX as default location.")
		d.ServerLocation = defaultLocation
	}
	if d.ServerType == "" {
		log.Info("Type has not been set, will be used s1.c1.medium as default type.")
		d.ServerType = defaultType
	}
	if d.ServerOs == "" {
		log.Info("OS has not been set, will be used ubuntu/bionic as default type.")
		d.ServerOs = defaultOS
	}

	return nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {

	if strings.Contains(d.ServerOs, "ubuntu") {
		d.SSHUser = "ubuntu"
	} else if strings.Contains(d.ServerOs, "centos") {
		d.SSHUser = "centos"
	}

	return d.SSHUser
}
