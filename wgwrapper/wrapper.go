package wgwrapper

import (
	"fmt"
	"os"

	// ipc အစား conn ကို ပြောင်းလဲ Import လုပ်ပါ
	"github.com/amnezia-vpn/amneziawg-go/conn"
	"github.com/amnezia-vpn/amneziawg-go/device"
	"github.com/amnezia-vpn/amneziawg-go/tun"
)

var wgDevice *device.Device

func StartVPN(fd int, configStr string) string {
	
	file := os.NewFile(uintptr(fd), "tun")

	tunDevice, err := tun.CreateTUNFromFile(file, 0)
	if err != nil {
		return "Error creating TUN device: " + err.Error()
	}

	logger := device.NewLogger(device.LogLevelError, "MHWARPvpn")
	
	// ဒီနေရာမှာ conn.NewDefaultBind() ကို ပြင်ဆင်အသုံးပြုထားပါတယ်
	bind := conn.NewDefaultBind()
	wgDevice = device.NewDevice(tunDevice, bind, logger)

	err = wgDevice.Up()
	if err != nil {
		return "Failed to bring up device: " + err.Error()
	}

	// နောက်တစ်ဆင့်အနေဖြင့် configStr ကို UAPI Format ပြောင်းပြီး wgDevice.IpcSet() ကို ခေါ်ရပါမည်

	return fmt.Sprintf("VPN Started Successfully with FD: %d", fd)
}

func StopVPN() string {
	if wgDevice != nil {
		wgDevice.Close()
		wgDevice = nil
		return "VPN Stopped"
	}
	return "No VPN running"
}
