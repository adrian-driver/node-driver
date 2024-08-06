package poc

import (
	// "encoding/base64"
	// "encoding/json"
	"fmt"
	// "io/ioutil"
	// "strings"
	// "time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	// "github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"

	// "github.com/pkg/errors"
)

// Driver is the implementation of BaseDriver interface
type Driver struct {
	*drivers.BaseDriver
}


// NewDriver creates and returns a new instance of the PNAP driver
func NewDriver() *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{},
	}
}

// GetCreateFlags returns the mcnflag.Flag slice representing the flags
// that can be set, their descriptions and defaults.
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "poc"
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return nil
}
func (d *Driver) createSSHKey() (string, error) {
	return "", nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Info("Creating pnap machine instance...")

	return nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetIP returns IP to use in communication
func (d *Driver) GetIP() (string, error) {
	log.Debug("pnap.GetIP()")

	return d.IPAddress, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	return state.Running, nil
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

	return nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	return nil
}

// Start a host
func (d *Driver) Start() error {
	return nil
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return nil
}

/* func run(command command.Executor) error {
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
} */

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {

	return nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	d.SSHUser = "ubuntu"
	return d.SSHUser
}

// setTokenToEmptySTring invalidates token.
// Token is definitelly expired after one hour, and this method enables other ways of authentication.
func (d *Driver) setTokenToEmptySTring() {
}
