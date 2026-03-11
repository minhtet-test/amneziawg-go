package wgwrapper

import (
	"fmt"
	"os"

	// AmneziaWG-go မှ လိုအပ်သော package များကို Import လုပ်ပါ
	"github.com/amnezia-vpn/amneziawg-go/device"
	"github.com/amnezia-vpn/amneziawg-go/ipc"
	"github.com/amnezia-vpn/amneziawg-go/tun"
)

// AmneziaWG Device ကို မှတ်သားထားရန်
var wgDevice *device.Device

// StartVPN သည် Android မှ လှမ်းခေါ်မည့် Function ဖြစ်သည်
func StartVPN(fd int, configStr string) string {
	
	// ၁။ Android မှပေးသော File Descriptor (fd) ကို Go ၏ os.File အဖြစ်ပြောင်းခြင်း
	file := os.NewFile(uintptr(fd), "tun")

	// ၂။ ၎င်း File မှတဆင့် TUN Interface ဖန်တီးခြင်း
	tunDevice, err := tun.CreateTUNFromFile(file, 0)
	if err != nil {
		return "Error creating TUN device: " + err.Error()
	}

	// ၃။ AmneziaWG Device အသစ်တည်ဆောက်ခြင်း
	logger := device.NewLogger(device.LogLevelError, "MHWARPvpn")
	wgDevice = device.NewDevice(tunDevice, ipc.NewUAPIBind(), logger)

	// ၄။ Device ကို စတင် (Up) လုပ်ခြင်း
	err = wgDevice.Up()
	if err != nil {
		return "Failed to bring up device: " + err.Error()
	}

	// မှတ်ချက် - လက်တွေ့တွင် configStr (INI format) ကို UAPI format သို့ ပြောင်းလဲပြီး 
	// wgDevice.IpcSet() ဖြင့် Config သတ်မှတ်ပေးရန် လိုအပ်ပါသည်။ (အောက်တွင် ရှင်းပြထားပါသည်)

	return fmt.Sprintf("VPN Started Successfully with FD: %d", fd)
}

// StopVPN သည် VPN ကို ပိတ်ရန်ဖြစ်သည်
func StopVPN() string {
	if wgDevice != nil {
		wgDevice.Close()
		wgDevice = nil
		return "VPN Stopped"
	}
	return "No VPN running"
}
