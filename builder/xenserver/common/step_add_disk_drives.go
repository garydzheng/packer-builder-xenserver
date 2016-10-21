package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	xsclient "github.com/xenserver/go-xenserver-client"
)

type StepRemoveAllDiskDrives struct {
}

func (self *StepRemoveAllDiskDrives) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Step: Removing all disk drives")
	client := state.Get("client").(xsclient.XenAPIClient)
	uuid := state.Get("instance_uuid").(string)
	instance, err := client.GetVMByUuid(uuid)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to get VM from UUID '%s': %s", uuid, err.Error()))
		return multistep.ActionHalt
	}


	vdi, err := client.GetVdiByUuid(vdiUuid)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to get VDI from UUID '%s': %s", vdiUuid, err.Error()))
		return multistep.ActionHalt
	}


	vbds, err := instance.GetVBDs()
	if err != nil {
		ui.Error(fmt.Sprintf("Error getting VBDs: %s", err.Error()))
		return multistep.ActionHalt
	}
	var cd_vbds []xsclient.VBD
	for _, vbd := range vbds {
		vbd_rec, err := vbd.GetRecord()
		if err != nil {
			ui.Error(fmt.Sprintf("Error getting VBD record: %s", err.Error()))
			return multistep.ActionHalt
		}
		if strings.ToLower(vbd_rec["type"].(string)) == "cd" {
			cd_vbds = append(cd_vbds, vbd)
		}
	}
	for _, vbd := range cd_vbds {
		_ = vbd.Eject()
		_ = vbd.Unplug()
		err = vbd.Destroy()
		if err != nil {
			ui.Error(fmt.Sprintf("Error destroying VBD: %s", err.Error()))
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

func (self *StepRemoveAllDiskDrives) Cleanup(state multistep.StateBag) {
}