package xva

import (
	"strings"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	xsclient "github.com/xenserver/go-xenserver-client"
	xscommon "github.com/xenserver/packer-builder-xenserver/builder/xenserver/common"
//	"log"
)

type stepRemoveExcessDrives struct {
}

func (self *stepRemoveExcessDrives) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("commonconfig").(xscommon.CommonConfig)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Step: Remove excess drives")
	client := state.Get("client").(xsclient.XenAPIClient)
	uuid := state.Get("instance_uuid").(string)
	instance, err := client.GetVMByUuid(uuid)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to get VM from UUID '%s': %s", uuid, err.Error()))
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
			err = vbd.Eject()
			if err != nil && !strings.Contains(err.Error(), "VBD_IS_EMPTY") {
				ui.Error(fmt.Sprintf("Error ejecting VBD: %s", err.Error()))
				return multistep.ActionHalt
			}
			cd_vbds = append(cd_vbds, vbd)
		}
	}
	if len(cd_vbds) == 0 {
		return multistep.ActionContinue
	}
	if config.KeepDiskDrive {
		cd_vbds = cd_vbds[:len(cd_vbds)-1]
	}
	for _, vbd := range cd_vbds {
		_ = vbd.Unplug()
		err = vbd.Destroy()
		if err != nil {
			ui.Error(fmt.Sprintf("Error destroying VBD: %s", err.Error()))
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

func (self *stepRemoveExcessDrives) Cleanup(state multistep.StateBag) {
}
