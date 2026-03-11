package wgwrapper

import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/amnezia-vpn/amneziawg-go/conn"
	"github.com/amnezia-vpn/amneziawg-go/device"
	"github.com/amnezia-vpn/amneziawg-go/tun"
)

var wgDevice *device.Device

// StartVPN သည် Android မှ လှမ်းခေါ်မည့် Function ဖြစ်သည်
func StartVPN(fd int, configStr string) string {
	
	// ၁။ Android မှပေးသော File Descriptor (fd) ဖြင့် TUN ဖွင့်ခြင်း
	file := os.NewFile(uintptr(fd), "tun")
	tunDevice, err := tun.CreateTUNFromFile(file, 0)
	if err != nil {
		return "Error creating TUN device: " + err.Error()
	}

	logger := device.NewLogger(device.LogLevelError, "MHWARPvpn")
	bind := conn.NewDefaultBind()
	wgDevice = device.NewDevice(tunDevice, bind, logger)

	// ၂။ API မှရသော INI Config ကို UAPI Format သို့ ပြောင်းလဲခြင်း
	uapiConfig, err := iniToUAPI(configStr)
	if err != nil {
		return "Config parse error: " + err.Error()
	}

	// ၃။ ပြောင်းလဲထားသော UAPI Config ကို Device သို့ ထည့်သွင်းခြင်း
	err = wgDevice.IpcSet(uapiConfig)
	if err != nil {
		return "Failed to set config: " + err.Error()
	}

	// ၄။ VPN Device ကို စတင်အလုပ်လုပ်ခိုင်းခြင်း
	err = wgDevice.Up()
	if err != nil {
		return "Failed to bring up device: " + err.Error()
	}

	return fmt.Sprintf("VPN Started Successfully with Config length: %d", len(configStr))
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

// INI Format မှ UAPI သို့ ပြောင်းလဲပေးသော Helper Function
func iniToUAPI(ini string) (string, error) {
	var uapi strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(ini))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Section headers နှင့် Comments များကို ကျော်သွားပါ
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch strings.ToLower(key) {
		case "privatekey":
			hexKey, _ := base64ToHex(val)
			uapi.WriteString(fmt.Sprintf("private_key=%s\n", hexKey))
		case "publickey":
			hexKey, _ := base64ToHex(val)
			uapi.WriteString(fmt.Sprintf("public_key=%s\n", hexKey))
		case "endpoint":
			uapi.WriteString(fmt.Sprintf("endpoint=%s\n", val))
		case "allowedips":
			ips := strings.Split(val, ",")
			for _, ip := range ips {
				uapi.WriteString(fmt.Sprintf("allowed_ip=%s\n", strings.TrimSpace(ip)))
			}
		case "persistentkeepalive":
			uapi.WriteString(fmt.Sprintf("persistent_keepalive_interval=%s\n", val))
		
		// AmneziaWG ၏ အထူး Parameters များ (J1-J4, S1, S2, H1-H4 စသည်တို့)
		case "jc", "jmin", "jmax", "s1", "s2", "h1", "h2", "h3", "h4", "i1":
			uapi.WriteString(fmt.Sprintf("%s=%s\n", strings.ToLower(key), val))
		}
	}
	return uapi.String(), nil
}

// Base64 ကို Hex သို့ ပြောင်းပေးသော Function (UAPI က Keys များကို Hex ဖြင့်သာ လက်ခံသောကြောင့်ဖြစ်သည်)
func base64ToHex(b64 string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(decoded), nil
}
