/*
test-cec runs CEC commands.

It is a simple command line utility to test CEC commands
and to demonstrate how to use the cec package.

Usage:

	test-cec [flags] [path ...]

The flags are:

		-device
		    Send CEC commands to the specified CEC device.
	        Default is /dev/ttyACM0, but can be overridden with this flag.
		-h, --help
		    Show this help message and exit.
*/
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/chbmuc/cec"
)

// TODO:  Decide if using these hash maps is sane or not
/* cecCommands - A map of command hex code to name
   Source: http://www.cec-o-matic.com/
   Note: Some commands accept arguments.
*/
var cecCommands = map[int]string{0x82: "ActiveSource", 0x04: "ImageViewOn", 0x0D: "TextViewOn", // One Touch Play
	0x0B: "RecordOff", 0x09: "RecordOn", 0x0A: "RecordStatus", 0x0F: "RecordTVScreen", // One Touch Record
	0x1B: "DeckStatus", 0x1A: "GiveDeckStatus", 0x42: "DeckControl", 0x41: "Play", // Deck Control
	0x08: "GiveTunerDeviceStatus", 0x92: "SelectAnalogueService", 0x93: "SelectDigitalService", // Tuner Control
	0x06: "TunerStepDecrement", 0x05: "TunerStepIncrement",
	0x07: "TunerDeviceStatus",
	0x71: "GiveAudioStatus", 0x7D: "GiveSystemAudioModeStatus", 0x7A: "ReportAudioStatus", // System Audio Control
	0x72: "SetSystemAudioMode", 0x70: "SystemAudioModeRequest", 0x7E: "SystemAudioModeStatus",
	0x44: "UserControlPressed", 0x45: "UserControlReleased", // Device Menu Control & Remote Control Passthrough
	0x8D: "MenuRequest", 0x8E: "MenuStatus",
	0x33: "ClearAnalogueTimer", 0x99: "ClearDigitalTimer", 0xA1: "ClearExternalTimer", // Timer Programming
	0x34: "SetAnalogueTimer", 0x97: "SetDigitalTimer", 0xA2: "SetExternalTimer",
	0x67: "SetTimerProgramTitle", 0x43: "TimerClearedStatus", 0x35: "TimerStatus",
	0x36: "Standby"} // System Standby

// From libcec "typedef enum cec_device_type"
/*const (
    CEC_DEVICE_TYPE_TV = iota        // 0
    CEC_DEVICE_TYPE_RECORDING_DEVICE // 1
    CEC_DEVICE_TYPE_RESERVED         // 2
    CEC_DEVICE_TYPE_TUNER            // 3
    CEC_DEVICE_TYPE_PLAYBACK_DEVICE  // 4
    CEC_DEVICE_TYPE_AUDIO_SYSTEM     // 5
)*/

// TODO: Figure out if this is right, or should we just use LogicalAddress only?
var cec_device_type = map[int]string{0: "TV",
	1:   "Recording1",
	2:   "Recording2",
	3:   "Tuner1",
	4:   "Playback1",
	5:   "Audio",
	6:   "Tuner2",
	7:   "Tuner3",
	8:   "Playback2",
	9:   "Playback3",
	0xA: "Tuner4",
	0xB: "Playback3",
	0xC: "Reserved1",
	0xD: "Reserved2",
	0xE: "Free", // "Reserved3" - shown as "Free use" in libcec.go output?
	0xF: "Unregistered"}

// cecCommand - Returns CEC command by name from cecCommands hashmap above
func cecCommand(cmd string) (key int, err error) {
	for opcode, name := range cecCommands {
		fmt.Printf("Comparing cmd %s with name %s = %#x\n", cmd, name, opcode)
		if name == cmd {
			key = opcode
			err = nil
			return
		}
	}
	err = fmt.Errorf("Could not find CEC code for command: %s", cmd)
	return
}

// bool2English - Helper function to translate boolean value to english (used in deviceToString())
func bool2English(in bool) string {
	if in {
		return "yes"
	}
	return "no"
}

// deviceToString - Emulate the output of "echo scan | cec-client -s" on per-device basis
func deviceToString(name string, d cec.Device) (print_result string) {
	print_result = ""

	fmt_str := "%-15s %s\n"
	print_result += fmt.Sprintf(fmt_str, fmt.Sprintf("%s %X:", "device", d.LogicalAddress), name)
	print_result += fmt.Sprintf(fmt_str, "address:", d.PhysicalAddress)
	print_result += fmt.Sprintf(fmt_str, "active source:", bool2English(d.ActiveSource))
	print_result += fmt.Sprintf(fmt_str, "vendor:", d.Vendor)
	print_result += fmt.Sprintf(fmt_str, "osd string:", d.OSDName)
	print_result += fmt.Sprintf(fmt_str, "power status:", d.PowerStatus)
	return
}

// GetDeviceByOSDName - Take a cec.Connection & osd_name input, use Regex match to find matching (device, device_type) pair to return
func GetDeviceByOSDName(c *cec.Connection, osd_name string) (d cec.Device, d_type string) {
	device_list := c.List()

	fmt.Println("CEC bus information")
	fmt.Println("===================")
	fmt.Printf("Found %d devices!\n", len(device_list))

	for device_type, dev := range device_list {
		regex := ".*" + osd_name + ".*"
		fmt.Printf("Trying to match against Regex: %s\n", regex)
		fmt.Println(deviceToString(device_type, dev))
		re := regexp.MustCompile(regex)
		if matched := re.MatchString(dev.OSDName); matched {
			fmt.Println("MATCHED!")
			fmt.Println(matched)
			d = dev
			d_type = device_type
		} else {
			fmt.Println("NOT MATCHED!")
			fmt.Println(matched)
		}
	}
	return
}

// main - Main entrypoint to test-cec golang binary command
//
//	Just some testing example code to implement basic english -> CEC commands:
//	 - PowerOn by device name
//	 - Run "ActiveSource" command by cmd name, Autodetect this device's physical address via os.Hostname
//
// TODO?: Figure out why Golang client grabs weird `f.f.f.f` physicaladdress when it should be 4.0.0.0
func main() {
	// CEC adapter device path
	// Default to `/dev/ttyACM0`, but allow user to override with `-device` flag
	// Examples: `/dev/ttyACM0``, `/dev/ttyUSB0`, or `/dev/cec0``
	var devicePath string
	var help bool
	flag.StringVar(&devicePath, "device", "/dev/ttyACM0",
		"Path to CEC device, e.g. `/dev/ttyACM0`, `/dev/ttyUSB0`, or `/dev/cec0`")
	flag.BoolVar(&help, "help", false, "Show this help message and exit")
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	// Set this CEC client name to os.Hostname(), or default to "cec.go"
	cec_device_name, err := os.Hostname()
	if err != nil {
		cec_device_name = "cec.go"
	}
	fmt.Println("Using CEC Device Name: ", cec_device_name)
	c, err := cec.Open(cec_device_name, devicePath, true)
	if err != nil {
		fmt.Println(err)
	}
	//c.PowerOn(0)
	//    device_list := c.List()
	//    for key, value := range device_list {
	//        fmt.Printf("value: %#v\n", value)
	//       fmt.Println(deviceToString(key, value))
	//    }
	// Autodetect TV Logical Address by regex match for "TV"
	tv, _ := GetDeviceByOSDName(c, "TV")
	c.PowerOn(tv.LogicalAddress)

	// example to set active source for device at address 4.0.0.0
	// '2F:84:40:00'
	// From http://www.cec-o-matic.com/
	// Recording 2 -> TV : Set Active Source 1.0.0.0
	// '20:82:10:00'

	// Test out crafting CEC ActiveSource command by name -> opcode lookup
	cmd, err := cecCommand("ActiveSource")
	if err == nil {
		fmt.Println(fmt.Sprintf("Attempt to transmit command: %s = %#x", "ActiveSource", cmd))

		// Using Broadcast dest seems to be recommended way
		// Reference: https://blog.gordonturner.com/2016/12/14/using-cec-client-on-a-raspberry-pi/
		//
		// Playback 1 -> Broadcast : Set Active Source 1.0.0.0
		// '4F:82:10:00'

		// Get this device by the os.Hostname we detected earlier (or default to "cec.go")
		self, self_device_type := GetDeviceByOSDName(c, cec_device_name)
		// DEBUG info
		fmt.Println(self_device_type)
		fmt.Println(self.LogicalAddress)
		fmt.Printf("%q\n", strings.Split(self.PhysicalAddress, "."))

		// Autodetect this client's PhysicalAddress, split by periods ('.')
		target_address := strings.Split(self.PhysicalAddress, ".")

		//set_active_device := fmt.Sprintf("%XF:82:%s:%s", self.LogicalAddress, target_address[0:1], target_address[2:3])

		// Construct the SetActive device command using components
		set_active_device := fmt.Sprintf("%XF:82:%s:%s", self.LogicalAddress, strings.Join(target_address[0:2], ""), strings.Join(target_address[2:4], ""))

		// Debug output the constructed command string (HEX as literal hex char string, NOT bytes or int)
		fmt.Println(set_active_device)

		// This does work, usually...
		//c.Transmit("4F:82:40:00")

		// This detects `f.f.f.f` address incorrectly... but constructed command looks ok
		c.Transmit(set_active_device)
	} else {
		// Throw error
		fmt.Println("Error: ", err)
	}
}

/* sample Golang cec.Device type data structure contents

device         2: Playback2
value: cec.Device{OSDName:"BD\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00", Vendor:"Sony", LogicalAddress:8, ActiveSource:false, PowerStatus:"standby", PhysicalAddress:"2.0.0.0"}



sample output for: echo "scan" | cec-client -s

CEC bus information
===================
device #1: Recorder 1
address:       1.0.0.0
active source: no
vendor:        Pulse Eight
osd string:    Kodi
CEC version:   1.4
power status:  on
language:      eng


device #2: Recorder 2
address:       1.0.0.0
active source: no
vendor:        Pulse Eight
osd string:    CECTester
CEC version:   1.4
power status:  on
language:      eng


device #8: Playback 2
address:       2.0.0.0
active source: no
vendor:        Sony
osd string:    BD
CEC version:   1.4
power status:  standby
language:      ???


device #E: Free use
address:       f.f.f.f
active source: no
vendor:        Unknown
osd string:    Free use
CEC version:   unknown
power status:  unknown
language:      ???


^^^^ This last one is from this code??  Why "Free use" and `f.f.f.f`??
*/
