package service

import (
	"github.com/libvirt/libvirt-go"
	"fmt"
	"encoding/xml"
	"crypto/rand"
	"errors"
	"strings"
)

type virDomainOSType struct {
	Name    string `xml:",innerxml"`
	Arch    string `xml:"arch,attr"`
	Machine string `xml:"machine,attr"`
}

type virDomainBootDevice struct {
	Device string `xml:"dev,attr"`
}

type virDomainOSElement struct {
	Type      virDomainOSType       `xml:"type"`
	BootOrder []virDomainBootDevice `xml:"boot"`
	//todo:bootmenu/bootloader/kernal/initrd
}

type virDomainInterfaceTarget struct {
	Device  string `xml:"dev,attr,omitempty"`
}

type virDomainInterfaceSource struct {
	Bridge  string `xml:"bridge,attr,omitempty"`
	Network string `xml:"network,attr,omitempty"`
}

type virDomainInterfaceMAC struct {
	Address string `xml:"address,attr"`
}

type virDomainInterfaceModel struct {
	Type string `xml:"type,attr"`
}

type virDomainInterfaceLimit struct {
	Average uint `xml:"average,attr,omitempty"`
	Peak    uint `xml:"peak,attr,omitempty"`
	Burst   uint `xml:"burst,attr,omitempty"`
}

type virDomainInterfaceBandwidth struct {
	Inbound  *virDomainInterfaceLimit `xml:"inbound,omitempty"`
	Outbound *virDomainInterfaceLimit `xml:"outbound,omitempty"`
}

type virDomainInterfaceElement struct {
	XMLName   xml.Name                     `xml:"interface"`
	Type      string                       `xml:"type,attr"`
	Source    virDomainInterfaceSource     `xml:"source,omitempty"`
	MAC       *virDomainInterfaceMAC       `xml:"mac,omitempty"`
	Model     *virDomainInterfaceModel     `xml:"model,omitempty"`
	Target    *virDomainInterfaceTarget    `xml:"target,omitempty"`
	Bandwidth *virDomainInterfaceBandwidth `xml:"bandwidth,omitempty"`
}

type virDomainGraphicsListen struct {
	Type    string `xml:"type,attr,omitempty"`
	Address string `xml:"address,attr,omitempty"`
}

type virDomainGraphicsElement struct {
	Type     string                   `xml:"type,attr"`
	Port     uint                     `xml:"port,attr"`
	Password string                   `xml:"passwd,attr"`
	Listen   *virDomainGraphicsListen `xml:"listen,omitempty"`
}

type virDomainControllerElement struct {
	Type  string `xml:"type,attr"`
	Index string `xml:"index,attr"`
	Model string `xml:"model,attr,omitempty"`
}

type virDomainMemoryStats struct {
	Period uint `xml:"period,attr"`
}

type virDomainMemoryBalloon struct {
	Model string               `xml:"model,attr"`
	Stats virDomainMemoryStats `xml:"stats"`
}

type virDomainChannelTarget struct {
	Type string `xml:"type,attr"`
	Name string `xml:"name,attr"`
}

type virDomainChannel struct {
	Type   string                 `xml:"type,attr"`
	Target virDomainChannelTarget `xml:"target"`
}

type virDomainInput struct {
	Type string `xml:"type,attr"`
	Bus  string `xml:"bus,attr"`
}

type virDomainDevicesElement struct {
	Emulator      string                       `xml:"emulator"`
	Disks         []virDomainDiskElement       `xml:"disk,omitempty"`
	Interface     []virDomainInterfaceElement  `xml:"interface,omitempty"`
	Graphics      virDomainGraphicsElement     `xml:"graphics"`
	Controller    []virDomainControllerElement `xml:"controller,omitempty"`
	Input         []virDomainInput             `xml:"input,omitempty"`
	MemoryBalloon virDomainMemoryBalloon       `xml:"memballoon"`
	Channel       virDomainChannel             `xml:"channel"`
}

type virDomainCpuElement struct {
	Topology virDomainCpuTopology `xml:"topology"`
}

type virDomainCpuTopology struct {
	Sockets uint `xml:"sockets,attr"`
	Cores   uint `xml:"cores,attr"`
	Threads uint `xml:"threads,attr"`
}

type virDomainSuspendToDisk struct {
	Enabled string `xml:"enabled,attr,omitempty"`
}

type virDomainSuspendToMem struct {
	Enabled string `xml:"enabled,attr,omitempty"`
}

type virDomainPowerElement struct {
	Disk virDomainSuspendToDisk `xml:"suspend-to-disk"`
	Mem  virDomainSuspendToMem  `xml:"suspend-to-mem"`
}

type virDomainFeaturePAE struct {
	XMLName xml.Name `xml:"pae"`
}

type virDomainFeatureACPI struct {
	XMLName xml.Name `xml:"acpi"`
}

type virDomainFeatureAPIC struct {
	XMLName xml.Name `xml:"apic"`
}


type virDomainFeatureElement struct {
	PAE  *virDomainFeaturePAE
	ACPI *virDomainFeatureACPI
	APIC *virDomainFeatureAPIC
}

type virDomainClockElement struct {
	Offset string `xml:"offset,attr,omitempty"`
}

type virDomainDiskDriver struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type virDomainDiskSourceHost struct {
	Name string `xml:"name,attr"`
	Port uint    `xml:"port,attr"`
}

type virDomainDiskSource struct {
	File     string                   `xml:"file,attr,omitempty"`
	Protocol string                   `xml:"protocol,attr,omitempty"`
	Name     string                   `xml:"name,attr,omitempty"`
	Pool     string                   `xml:"pool,attr,omitempty"`
	Volume   string                   `xml:"volume,attr,omitempty"`
	Host     *virDomainDiskSourceHost `xml:"host,omitempty"`
}

type virDomainDiskTarget struct {
	Device string `xml:"dev,attr,omitempty"`
	Bus    string `xml:"bus,attr,omitempty"`
}

type virDomainDiskTune struct {
	ReadBytePerSecond  uint `xml:"read_bytes_sec,omitempty"`
	WriteBytePerSecond uint `xml:"write_bytes_sec,omitempty"`
	ReadIOPerSecond    int  `xml:"read_iops_sec,omitempty"`
	WriteIOPerSecond   int  `xml:"write_iops_sec,omitempty"`
}

type virDomainDiskElement struct {
	XMLName  xml.Name             `xml:"disk"`
	Type     string               `xml:"type,attr"`
	Device   string               `xml:"device,attr"`
	Driver   virDomainDiskDriver  `xml:"driver"`
	Target   virDomainDiskTarget  `xml:"target,omitempty"`
	ReadOnly *bool                `xml:"readonly,omitempty"`
	Source   *virDomainDiskSource `xml:"source,omitempty"`
	IoTune   *virDomainDiskTune   `xml:"iotune,omitempty"`
}

type virDomainCPUTuneDefine struct {
	Shares uint `xml:"shares"`
	Period uint `xml:"period"`
	Quota  uint `xml:"quota"`
}

type virDomainDefine struct {
	XMLName     xml.Name                `xml:"domain"`
	Type        string                  `xml:"type,attr"`
	Name        string                  `xml:"name"`
	UUID        string                  `xml:"uuid,omitempty"`
	Memory      uint                    `xml:"memory"` //Default in KiB
	VCpu        uint                    `xml:"vcpu"`
	OS          virDomainOSElement      `xml:"os"`
	CPU         virDomainCpuElement     `xml:"cpu"`
	CPUTune     virDomainCPUTuneDefine  `xml:"cputune,omitempty"`
	Devices     virDomainDevicesElement `xml:"devices,omitempty"`
	OnPowerOff  string                  `xml:"on_poweroff,omitempty"`
	OnReboot    string                  `xml:"on_reboot,omitempty"`
	OnCrash     string                  `xml:"on_crash,omitempty""`
	PowerManage virDomainPowerElement   `xml:"pm,omitempty"`
	Features    virDomainFeatureElement `xml:"features,omitempty"`
	Clock       virDomainClockElement   `xml:"clock"`
}

type configTemplate struct {
	System       string
	Admin        string
	DiskBus      string
	NetworkModel string
	USBModel     string
	TabletBus    string
}

const (
	IDEOffsetCDROM      = iota
	IDEOffsetCIDATA
	IDEOffsetDISK
)

const (
	DiskTypeNetwork      = "network"
	DiskTypeBlock        = "block"
	DiskTypeFile         = "file"
	DiskTypeVolume       = "volume"
	DeviceCDROM          = "cdrom"
	DeviceDisk           = "disk"
	DriverNameQEMU       = "qemu"
	DriverTypeRaw        = "raw"
	DriverTypeQCOW2      = "qcow2"
	StartDeviceCharacter = 0x61 //'a'
	DevicePrefixIDE      = "hd"
	DevicePrefixSCSI     = "sd"
	DiskBusIDE           = "ide"
	DiskBusSCSI          = "scsi"
	DiskBusSATA          = "sata"
	ProtocolHTTPS        = "https"
	NetworkModelRTL8139  = "rtl8139"
	NetworkModelE1000    = "e1000"
	NetworkModelVIRTIO   = "virtio"
	USBModelXHCI         = "nec-xhci"
	USBModelDefault      = ""
	TabletBusVIRTIO      = "virtio"
	TabletBusUSB         = "usb"
	InputTablet          = "tablet"

	PCIController          = "pci"
	DefaultControllerIndex = "0"
	DefaultControllerModel = "pci-root"
	VirtioSCSIController   = "scsi"
	VirtioSCSIModel        = "virtio-scsi"
	USBController          = "usb"
)

type InstanceUtility struct {
	virConnect      *libvirt.Connect
	systemTemplates map[string]configTemplate
}

func CreateInstanceUtility(connect *libvirt.Connect) (util *InstanceUtility, err error) {
	util = &InstanceUtility{}
	util.virConnect = connect
	util.systemTemplates = map[string]configTemplate{
		SystemVersionCentOS7:    configTemplate{SystemNameLinux, AdminLinux, DiskBusSCSI, NetworkModelVIRTIO, USBModelXHCI, TabletBusUSB},
		SystemVersionCentOS6:    configTemplate{SystemNameLinux, AdminLinux, DiskBusSATA, NetworkModelVIRTIO, USBModelXHCI, TabletBusUSB},
		SystemVersionWindow2012: configTemplate{SystemNameWindows, AdminWindows, DiskBusSATA, NetworkModelE1000, USBModelXHCI, TabletBusUSB},
		SystemVersionGeneral:    configTemplate{SystemNameLinux, AdminLinux, DiskBusSATA, NetworkModelRTL8139, USBModelDefault, TabletBusUSB},
		SystemVersionLegacy:     configTemplate{SystemNameLinux, AdminLinux, DiskBusIDE, NetworkModelRTL8139, USBModelDefault, TabletBusUSB},
	}
	return util, nil
}

func (util *InstanceUtility) GetSystemTemplate(version string) (template configTemplate, err error){
	template, exists := util.systemTemplates[version]
	if !exists{
		err = fmt.Errorf("unsupport system version '%s'", version)
		return
	}
	return template, nil
}

func (util *InstanceUtility) CreateInstance(config GuestConfig) (guest GuestConfig, err error) {
	define, err := util.createDefine(config)
	if err != nil {
		return guest, err
	}
	data, err := xml.MarshalIndent(define, "", " ")
	if err != nil{
		return guest, err
	}
	virDomain, err := util.virConnect.DomainDefineXML(string(data))
	if err != nil{
		return guest, err
	}
	if config.AutoStart{
		if err = virDomain.SetAutostart(true);err != nil{
			return guest, err
		}
	}
	config.Created = true
	//todo: complete return guest
	return config, nil
}

func (util *InstanceUtility) DeleteInstance(id string) error {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	running, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if running {
		return fmt.Errorf("instance '%s' is running", id)
	}
	return virDomain.Undefine()
}

func (util *InstanceUtility) Exists(id string) bool {
	_, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return false
	}
	return true
}

func (util *InstanceUtility) GetCPUTimes(id string) (usedNanoseconds uint64, cores uint, err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return
	}
	info, err := virDomain.GetInfo()
	if err != nil{
		return
	}
	return info.CpuTime, info.NrVirtCpu, nil
}

func (util *InstanceUtility) GetIPv4Address(id, mac string) (ip string, err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return
	}
	isRunning, err := virDomain.IsActive()
	if err != nil{
		return
	}
	if !isRunning{
		err = fmt.Errorf("instance '%s' not running", id)
		return
	}
	ifs, err := virDomain.ListAllInterfaceAddresses(libvirt.DOMAIN_INTERFACE_ADDRESSES_SRC_AGENT)
	if err != nil{
		return
	}
	for _, guestInterface := range ifs{
		if guestInterface.Hwaddr == mac{
			if 0 == len(guestInterface.Addrs){
				err = fmt.Errorf("no address found in interface '%s'", guestInterface.Name)
				return "", err
			}
			for _, addr := range guestInterface.Addrs{
				if libvirt.IP_ADDR_TYPE_IPV4 == libvirt.IPAddrType(addr.Type){
					ip = addr.Addr
					return ip, nil
				}
			}
		}
	}
	//no hwaddress matched
	return "", nil
}

func (util *InstanceUtility) GetInstanceStatus(id string) (ins InstanceStatus, err error) {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return ins, err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return ins, err
	}
	ins.Running = isRunning
	if !isRunning{
		return ins, nil
	}
	//running status
	{
		//memory stats
		var statsCount = uint32(libvirt.DOMAIN_MEMORY_STAT_LAST)
		memStats, err := virDomain.MemoryStats(statsCount, 0)
		if err != nil{
			return ins, err
		}
		//size in KB
		var availableValue, rssValue uint64 = 0, 0
		for _, stats := range memStats{
			if stats.Tag == int32(libvirt.DOMAIN_MEMORY_STAT_AVAILABLE){
				availableValue = stats.Val
				break
			}else if stats.Tag == int32(libvirt.DOMAIN_MEMORY_STAT_RSS){
				rssValue = stats.Val
			}
		}
		if 0 != availableValue{
			ins.AvailableMemory = availableValue << 10
		}else if 0 != rssValue{
			maxMemory, err := virDomain.GetMaxMemory()
			if err != nil{
				return ins, err
			}

			ins.AvailableMemory = (maxMemory- rssValue) << 10
			//log.Printf("debug: max %d, rss %d, avail %d", maxMemory, rssValue, ins.AvailableMemory)
			//log.Println("<instance> warning: available memory stats not supported, using (max - rss) instead")
		}else{
			return ins, errors.New("available or rss memory stats not supported")
		}
	}
	desc, err := virDomain.GetXMLDesc(0)
	if err != nil{
		return ins, err
	}
	var define virDomainDefine
	if err = xml.Unmarshal([]byte(desc), &define); err != nil{
		return ins, err
	}

	{
		//disk io
		const (
			DiskTypeVolume  = "volume"
		)
		ins.BytesRead = 0
		ins.BytesWritten = 0
		ins.AvailableDisk = 0
		for _, virDisk := range define.Devices.Disks{
			if virDisk.Type != DiskTypeVolume{
				continue
			}
			var devName = virDisk.Target.Device
			info, err:= virDomain.GetBlockInfo(devName, 0)
			if err != nil{
				return ins, err
			}
			ins.AvailableDisk += info.Capacity - info.Allocation
			stats, err := virDomain.BlockStats(devName)
			if err != nil{
				return ins, err
			}
			ins.BytesRead += uint64(stats.RdBytes)
			ins.BytesWritten += uint64(stats.WrBytes)
		}
	}
	{
		//network io
		ins.BytesSent = 0
		ins.BytesReceived = 0
		for index, inf := range define.Devices.Interface{
			if inf.Target == nil{
				return ins, fmt.Errorf("no target available for interface %d", index)
			}

			stats, err := virDomain.InterfaceStats(inf.Target.Device)
			if err != nil{
				return ins, err
			}
			ins.BytesReceived += uint64(stats.RxBytes)
			ins.BytesSent += uint64(stats.TxBytes)
		}
	}
	return ins, nil
}

func (util *InstanceUtility) IsInstanceRunning(id string) (bool, error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return false, err
	}
	return virDomain.IsActive()
}

func (util *InstanceUtility) StartInstance(id string) error {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if isRunning {
		return fmt.Errorf("instance '%s' already started", id)
	}
	return virDomain.Create()
}

func (util *InstanceUtility) StartInstanceWithMedia(id, host, url string, port uint) error {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if isRunning {
		return fmt.Errorf("instance '%s' already started", id)
	}

	var ideDevChar = StartDeviceCharacter

	var devName = fmt.Sprintf("%s%c", DevicePrefixIDE, ideDevChar +IDEOffsetCDROM)

	var deviceWithMedia, deviceWithoutMedia string
	var readyOnly = true
	{
		//empty ide cdrom
		var emptyDriver = virDomainDiskElement{Type:DiskTypeBlock, Device: DeviceCDROM, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeRaw},
			Target: virDomainDiskTarget{devName, DiskBusIDE}, ReadOnly: &readyOnly}
		if data, err := xml.MarshalIndent(emptyDriver, "", " "); err != nil {
			return err
		}else{
			deviceWithoutMedia = string(data)
		}
	}
	{
		var mediaSource = virDomainDiskSource{Protocol: ProtocolHTTPS, Name: url, Host:&virDomainDiskSourceHost{host, port}}
		var driverWithMedia = virDomainDiskElement{Type:DiskTypeNetwork, Device: DeviceCDROM, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeRaw},
			Target: virDomainDiskTarget{devName, DiskBusIDE}, Source:&mediaSource, ReadOnly: &readyOnly}
		if data, err := xml.MarshalIndent(driverWithMedia, "", " "); err != nil {
			return err
		}else{
			deviceWithMedia = string(data)
		}
	}
	//change config before start
	if err = virDomain.UpdateDeviceFlags(deviceWithMedia, libvirt.DOMAIN_DEVICE_MODIFY_CONFIG); err != nil{
		return err
	}
	if err = virDomain.Create();err != nil{
		return err
	}
	//change live config only
	if err = virDomain.UpdateDeviceFlags(deviceWithoutMedia, libvirt.DOMAIN_DEVICE_MODIFY_CONFIG); err != nil{
		virDomain.Destroy()
		return err
	}
	return nil
}

func (util *InstanceUtility) StopInstance(id string, reboot, force bool) error {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if !isRunning {
		return fmt.Errorf("instance '%s' already stopped", id)
	}
	//todo: check & unload cdrom before restart
	if reboot {
		if force {
			return virDomain.Reset(0)
		} else {
			return virDomain.Reboot(libvirt.DOMAIN_REBOOT_DEFAULT)
		}
	} else {
		if force {
			return virDomain.Destroy()
		} else {
			return virDomain.Shutdown()
		}
	}
}

func (util *InstanceUtility) InsertMedia(id, host, url string, port uint) (err error) {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if !isRunning {
		return fmt.Errorf("instance '%s' stopped", id)
	}
	//assume first ide device as cdrom
	var ideDevChar = StartDeviceCharacter

	var devName = fmt.Sprintf("%s%c", DevicePrefixIDE, ideDevChar +IDEOffsetCDROM)

	var deviceWithMedia string
	{
		var readyOnly = true
		var mediaSource = virDomainDiskSource{Protocol: ProtocolHTTPS, Name: url, Host:&virDomainDiskSourceHost{host, port}}
		var driverWithMedia = virDomainDiskElement{Type:DiskTypeNetwork, Device: DeviceCDROM, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeRaw},
			Target: virDomainDiskTarget{devName, DiskBusIDE}, Source:&mediaSource, ReadOnly: &readyOnly}
		if data, err := xml.MarshalIndent(driverWithMedia, "", " "); err != nil {
			return err
		}else{
			deviceWithMedia = string(data)
		}
	}
	//change online device
	return virDomain.UpdateDeviceFlags(deviceWithMedia, libvirt.DOMAIN_DEVICE_MODIFY_LIVE)
}

func (util *InstanceUtility) EjectMedia(id string) (err error) {
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if !isRunning {
		return fmt.Errorf("instance '%s' stopped", id)
	}
	//assume first ide device as cdrom
	var ideDevChar = StartDeviceCharacter
	var devName = fmt.Sprintf("%s%c", DevicePrefixIDE, ideDevChar +IDEOffsetCDROM)
	var deviceWithoutMedia string
	{
		var readyOnly = true
		//empty ide cdrom
		var emptyDriver = virDomainDiskElement{Type:DiskTypeBlock, Device: DeviceCDROM, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeRaw},
			Target: virDomainDiskTarget{devName, DiskBusIDE}, ReadOnly: &readyOnly}
		if data, err := xml.MarshalIndent(emptyDriver, "", " "); err != nil {
			return err
		}else{
			deviceWithoutMedia = string(data)
		}
	}
	//change alive device
	return virDomain.UpdateDeviceFlags(deviceWithoutMedia, libvirt.DOMAIN_DEVICE_MODIFY_LIVE)
}


func (util *InstanceUtility) ModifyCPUTopology(id string, core uint, immediate bool) (err error){
	const (
		TopologyFormat = "<topology sockets='%d' cores='%d' threads='%d'/>"
	)
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	var currentDomain virDomainDefine
	xmlDesc, err := virDomain.GetXMLDesc(0)
	if err != nil{
		return err
	}
	if err = xml.Unmarshal([]byte(xmlDesc), &currentDomain); err != nil{
		return err
	}
	var previousDefine = fmt.Sprintf(TopologyFormat, currentDomain.CPU.Topology.Sockets,
		currentDomain.CPU.Topology.Cores, currentDomain.CPU.Topology.Threads)
	var newTopology = virDomainCpuTopology{}
	if err = newTopology.SetCpuTopology(core); err != nil{
		return
	}

	var replaceData = fmt.Sprintf(TopologyFormat, newTopology.Sockets,
		newTopology.Cores, newTopology.Threads)

	var modifiedData = strings.Replace(xmlDesc, previousDefine, replaceData, 1)
	if virDomain, err = util.virConnect.DomainDefineXML(modifiedData); err != nil{
		return err
	}
	if err = virDomain.SetVcpusFlags(core, libvirt.DOMAIN_VCPU_CONFIG| libvirt.DOMAIN_VCPU_MAXIMUM); err != nil{
		return err
	}
	return nil
}

func (util *InstanceUtility) ModifyCore(id string, core uint, immediate bool) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	if err = virDomain.SetVcpusFlags(core, libvirt.DOMAIN_VCPU_CONFIG); err != nil{
		return
	}
	return nil
}

func (util *InstanceUtility) ModifyMemory(id string, memory uint, immediate bool) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	var memoryInKiB = uint64(memory >> 10)
	maxMemory, err := virDomain.GetMaxMemory()
	if err != nil{
		return err
	}
	if memoryInKiB > maxMemory{
		if err = virDomain.SetMemoryFlags(memoryInKiB, libvirt.DOMAIN_MEM_CONFIG| libvirt.DOMAIN_MEM_MAXIMUM); err != nil{
			return
		}
	}
	if err = virDomain.SetMemoryFlags(memoryInKiB, libvirt.DOMAIN_MEM_CONFIG); err != nil{
		return
	}
	return nil
}

func (util *InstanceUtility) ModifyPassword(id, user, password string) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(id)
	if err != nil {
		return err
	}
	err = virDomain.SetUserPassword(user, password, 0)
	return err
}

func (util *InstanceUtility) SetCPUThreshold(guestID string, priority PriorityEnum) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(guestID)
	if err != nil {
		return err
	}
	var xmlContent string
	xmlContent, err = virDomain.GetXMLDesc(0)
	if err != nil{
		return
	}
	var newDefine virDomainDefine
	if err = xml.Unmarshal([]byte(xmlContent), &newDefine); err != nil{
		return
	}
	if err = setCPUPriority(&newDefine, priority); err != nil{
		return
	}
	data, err := xml.MarshalIndent(newDefine, "", " ")
	if err != nil{
		return err
	}
	_, err =  util.virConnect.DomainDefineXML(string(data))
	if err != nil{
		err = fmt.Errorf("define fail: %s, content: %s", err.Error(), string(data))
	}
	return
}

func setCPUPriority(domain *virDomainDefine, priority PriorityEnum) (err error){
	const (
		periodPerSecond = 1000000
		quotaPerSecond  = 1000000
		highShares      = 2000
		mediumShares    = 1000
		lowShares       = 500
	)
	domain.CPUTune.Period = periodPerSecond
	switch priority {
	case PriorityHigh:
		domain.CPUTune.Shares = highShares
		domain.CPUTune.Quota = quotaPerSecond
		break
	case PriorityMedium:
		domain.CPUTune.Shares = mediumShares
		domain.CPUTune.Quota = quotaPerSecond / 2
		break
	case PriorityLow:
		domain.CPUTune.Shares = lowShares
		domain.CPUTune.Quota = quotaPerSecond / 4
		break
	default:
		return fmt.Errorf("invalid CPU priority %d", priority)
	}
	return nil
}

func (util *InstanceUtility) SetDiskThreshold(guestID string, writeSpeed, writeIOPS, readSpeed, readIOPS uint64) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(guestID)
	if err != nil {
		return err
	}
	var currentDomain virDomainDefine
	xmlDesc, err := virDomain.GetXMLDesc(0)
	if err != nil{
		return err
	}
	if err = xml.Unmarshal([]byte(xmlDesc), &currentDomain); err != nil{
		return err
	}
	var activated bool
	if activated, err = virDomain.IsActive(); err != nil{
		return err
	}
	if activated{
		err = fmt.Errorf("instance %s ('%s') is still running", currentDomain.Name, guestID)
		return
	}
	//var affectFlag = libvirt.DOMAIN_AFFECT_CONFIG | libvirt.DOMAIN_AFFECT_LIVE
	//deprecated: unsupported configuration: the block I/O throttling group parameter is not supported with this QEMU binary
	//var affectFlag = libvirt.DOMAIN_AFFECT_CONFIG
	//var parameters libvirt.DomainBlockIoTuneParameters
	//parameters.WriteIopsSec = writeIOPS
	//parameters.WriteIopsSecSet = true
	//parameters.ReadIopsSec = readIOPS
	//parameters.ReadIopsSecSet = true
	//parameters.GroupNameSet = true
	//parameters.GroupName = "default"
	//
	//for _, dev := range currentDomain.Devices.Disks {
	//	if nil != dev.ReadOnly {
	//		//ignore readonly device
	//		continue
	//	}
	//	if err = virDomain.SetBlockIoTune(dev.Target.Device, &parameters, affectFlag); err != nil {
	//		err = fmt.Errorf("set block io tune fail for dev '%s' : %s", dev.Target.Device, err.Error())
	//		return
	//	}
	//}
	//return nil

	var impactFlag = libvirt.DOMAIN_DEVICE_MODIFY_CONFIG

	for _, dev := range currentDomain.Devices.Disks {
		if nil != dev.ReadOnly {
			//ignore readonly device
			continue
		}
		//static configure
		var tune = dev.IoTune
		if tune == nil{
			tune = &virDomainDiskTune{}
		}
		tune.WriteIOPerSecond = int(writeIOPS)
		tune.ReadIOPerSecond = int(readIOPS)
		tune.WriteBytePerSecond = uint(writeSpeed)
		tune.ReadBytePerSecond = uint(readSpeed)

		dev.IoTune = tune
		data, err := xml.MarshalIndent(dev, "", " ")
		if err != nil{
			return err
		}
		if err = virDomain.UpdateDeviceFlags(string(data), impactFlag); err != nil{
			err = fmt.Errorf("update device fail: %s, content: %s", err.Error(), string(data))
			return err
		}
	}
	return nil
}

func (util *InstanceUtility) SetNetworkThreshold(guestID string, receiveSpeed, sendSpeed uint64) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(guestID)
	if err != nil {
		return err
	}
	var currentDomain virDomainDefine
	xmlDesc, err := virDomain.GetXMLDesc(0)
	if err != nil{
		return err
	}
	if err = xml.Unmarshal([]byte(xmlDesc), &currentDomain); err != nil{
		return err
	}
	var activated bool
	if activated, err = virDomain.IsActive(); err != nil{
		return err
	}
	//var affectFlag = libvirt.DOMAIN_AFFECT_CONFIG | libvirt.DOMAIN_AFFECT_LIVE
	var affectFlag = libvirt.DOMAIN_AFFECT_LIVE
	//API params
	var parameters libvirt.DomainInterfaceParameters
	if activated{
		if 0 != receiveSpeed{
			parameters.BandwidthInAverageSet = true
			parameters.BandwidthInAverage = uint(receiveSpeed >> 10) //kbytes
			parameters.BandwidthInPeakSet = true
			parameters.BandwidthInPeak = uint(receiveSpeed >> 9) // average * 2
			parameters.BandwidthInBurstSet = true
			parameters.BandwidthInBurst = uint(receiveSpeed >> 10)
		}
		if 0 != sendSpeed{
			parameters.BandwidthOutAverageSet = true
			parameters.BandwidthOutAverage = uint(sendSpeed >> 10) //kbytes
			parameters.BandwidthOutPeakSet = true
			parameters.BandwidthOutPeak = uint(sendSpeed >> 9) // average * 2
			parameters.BandwidthOutBurstSet = true
			parameters.BandwidthOutBurst = uint(sendSpeed >> 10)
		}
	}
	//configure params
	var impactFlag = libvirt.DOMAIN_DEVICE_MODIFY_CONFIG
	var bandWidth = virDomainInterfaceBandwidth{}
	if 0 != receiveSpeed{
		var inbound = virDomainInterfaceLimit{uint(receiveSpeed >> 10), uint(receiveSpeed >> 9), uint(receiveSpeed >> 10)}
		bandWidth.Inbound = &inbound
	}
	if 0 != sendSpeed{
		var outbound = virDomainInterfaceLimit{uint(sendSpeed >> 10), uint(sendSpeed >> 9), uint(sendSpeed >> 10)}
		bandWidth.Outbound = &outbound
	}

	for _, netInf := range currentDomain.Devices.Interface{
		if activated{
			if err = virDomain.SetInterfaceParameters(netInf.Target.Device, &parameters, affectFlag); err != nil{
				return
			}
		}
		netInf.Bandwidth = &bandWidth
		data, err := xml.MarshalIndent(netInf, "", " ")
		if err != nil{
			return err
		}
		if err = virDomain.UpdateDeviceFlags(string(data), impactFlag); err != nil{
			err = fmt.Errorf("update device fail: %s, content: %s", err.Error(), string(data))
			return err
		}
	}
	return nil
}

func (util *InstanceUtility) Rename(uuid, newName string) (err error){
	virDomain, err := util.virConnect.LookupDomainByUUIDString(uuid)
	if err != nil {
		return err
	}
	isRunning, err := virDomain.IsActive()
	if err != nil {
		return err
	}
	if isRunning {
		return fmt.Errorf("instance '%s' is still running", uuid)
	}
	var currentName string
	currentName, err = virDomain.GetName()
	if err != nil{
		return
	}
	if currentName == newName{
		return  errors.New("no need to change")
	}
	return virDomain.Rename(newName, 0)
}

func (util *InstanceUtility) createDefine(config GuestConfig) (define virDomainDefine, err error) {
	const (
		BootDeviceCDROM    = "cdrom"
		BootDeviceHardDisk = "hd"
	)
	define.Initial()
	define.Name = config.Name
	define.UUID = config.ID
	define.Memory = config.Memory >> 10
	define.VCpu = config.Cores


	//cpu
	if err = setCPUPriority(&define, config.CPUPriority); err != nil{
		return
	}

	if err = define.CPU.Topology.SetCpuTopology(config.Cores); err != nil {
		return
	}
	//os & boot
	template, err := util.GetSystemTemplate(config.SystemVersion)
	if err != nil{
		return
	}


	define.OS.BootOrder = []virDomainBootDevice{virDomainBootDevice{BootDeviceCDROM}, virDomainBootDevice{BootDeviceHardDisk}}

	switch config.StorageMode {
	case StorageModeLocal:
		if err = define.SetLocalVolumes(template.DiskBus, config.StoragePool, config.StorageVolumes, config.BootImage,
			config.ReadSpeed, config.ReadIOPS, config.WriteSpeed, config.WriteIOPS); err != nil {
			return
		}
	default:
		err = fmt.Errorf("unsupported storage mode :%d", config.StorageMode)
		return
	}
	switch config.NetworkMode {
	case NetworkModePlain:
		if err = define.SetPlainNetwork(template.NetworkModel, config.NetworkSource, config.HardwareAddress, config.ReceiveSpeed, config.SendSpeed); err != nil {
			return
		}
	default:
		err = fmt.Errorf("unsupported network mode :%d", config.NetworkMode)
		return
	}
	if err = define.SetVncDisplay(config.MonitorPort, config.MonitorSecret); err != nil {
		return
	}
	//tablet
	define.Devices.Input = append(define.Devices.Input, virDomainInput{InputTablet, template.TabletBus})
	if template.USBModel != USBModelDefault{
		define.Devices.Controller = append(define.Devices.Controller, virDomainControllerElement{USBController, DefaultControllerIndex, template.USBModel})
	}
	//qos

	return
}

func (topology *virDomainCpuTopology) SetCpuTopology(totalThreads uint) error {
	const (
		//SplitThreshold = 4
		ThreadPerCore  = 2
		MaxCores       = 1 << 5
		MaxSockets     = 1 << 3
	)
	if (totalThreads > 1 ) && (0 != (totalThreads % 2)) {
		return fmt.Errorf("even core number ( %d ) is not allowed", totalThreads)
	}
	var threads, cores, sockets uint
	if totalThreads < ThreadPerCore {
		threads = 1
		sockets = 1
		cores = totalThreads
	} else {
		threads = ThreadPerCore
		cores = totalThreads / threads
		sockets = 1
		if cores > MaxCores {
			for sockets = 2; sockets < MaxSockets+1; sockets = sockets << 1 {
				cores = (totalThreads / threads) / sockets
				if cores <= MaxCores {
					break
				}
			}
			if cores > MaxCores {
				return fmt.Errorf("no proper cpu topology fit total threads %d", totalThreads)
			}
		}
	}
	topology.Threads = threads
	topology.Cores = cores
	topology.Sockets = sockets
	return nil
}

func (define *virDomainDefine) SetLocalVolumes(diskBus, pool string, volumes []string, bootImage string,
	readSpeed, readIOPS, writeSpeed, writeIOPS uint64) error {
	var readyOnly = true
	{
		//empty ide cdrom
		var devName = fmt.Sprintf("%s%c", DevicePrefixIDE, StartDeviceCharacter + IDEOffsetCDROM)
		var cdromElement = virDomainDiskElement{Type:DiskTypeBlock, Device: DeviceCDROM, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeRaw},
			Target: virDomainDiskTarget{devName, DiskBusIDE}, ReadOnly: &readyOnly}
		define.Devices.Disks = append(define.Devices.Disks, cdromElement)
	}
	if "" != bootImage{
		//cdrom for ci data
		var ciDevice = fmt.Sprintf("%s%c", DevicePrefixIDE, StartDeviceCharacter + IDEOffsetCIDATA)
		var isoSource = virDomainDiskSource{File:bootImage}
		var ciElement = virDomainDiskElement{Type:DiskTypeFile, Device: DeviceCDROM, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeRaw},
			Target: virDomainDiskTarget{ciDevice, DiskBusIDE}, Source:&isoSource, ReadOnly: &readyOnly}
		define.Devices.Disks = append(define.Devices.Disks, ciElement)
	}
	var devicePrefix string
	var devChar int
	if DiskBusIDE == diskBus{
		//ide device
		devChar = StartDeviceCharacter + IDEOffsetDISK
		devicePrefix = DevicePrefixIDE
	}else{
		//sata/scsi
		devChar = StartDeviceCharacter
		devicePrefix = DevicePrefixSCSI
	}

	var ioTune *virDomainDiskTune
	if 0 != writeSpeed || 0 != writeIOPS || 0 != readSpeed || 0 != readIOPS{
		var limit = virDomainDiskTune{}
		limit.ReadBytePerSecond = uint(readSpeed)
		limit.ReadIOPerSecond = int(readIOPS)
		limit.WriteBytePerSecond = uint(writeSpeed)
		limit.WriteIOPerSecond = int(writeIOPS)
		ioTune = &limit
	}

	for _, volumeName := range volumes{
		var devName = fmt.Sprintf("%s%c", devicePrefix, devChar)
		var source = virDomainDiskSource{Pool:pool, Volume:volumeName}
		var diskElement = virDomainDiskElement{Type:DiskTypeVolume, Device: DeviceDisk, Driver: virDomainDiskDriver{DriverNameQEMU, DriverTypeQCOW2},
			Target: virDomainDiskTarget{devName, diskBus}, Source:&source}
		if ioTune != nil{
			diskElement.IoTune = ioTune
		}
		define.Devices.Disks = append(define.Devices.Disks, diskElement)
		devChar++
	}
	return nil
}

func (define *virDomainDefine) SetPlainNetwork(netBus, bridge, mac string, receiveSpeed, sendSpeed uint64) (err error) {
	if mac == "" {
		mac, err = define.generateMacAddress()
		if err != nil {
			return err
		}
	}
	const (
		InterfaceTypeBridge = "bridge"
	)
	var i = virDomainInterfaceElement{}
	i.Type = InterfaceTypeBridge
	i.MAC = &virDomainInterfaceMAC{mac}
	i.Source = virDomainInterfaceSource{Bridge: bridge}
	i.Model = &virDomainInterfaceModel{netBus}
	if 0 != receiveSpeed || 0 != sendSpeed{
		var bandWidth = virDomainInterfaceBandwidth{}
		if 0 != receiveSpeed{
			var limit = virDomainInterfaceLimit{}
			limit.Average = uint(receiveSpeed >> 10);
			limit.Peak = uint(receiveSpeed >> 9);
			limit.Burst = limit.Average
			bandWidth.Inbound = &limit
		}
		if 0 != sendSpeed{
			var limit = virDomainInterfaceLimit{}
			limit.Average = uint(sendSpeed >> 10);
			limit.Peak = uint(sendSpeed >> 9);
			limit.Burst = limit.Average
			bandWidth.Outbound = &limit
		}
		i.Bandwidth = &bandWidth
	}
	define.Devices.Interface = []virDomainInterfaceElement{i}
	return nil
}

func (define *virDomainDefine) generateMacAddress() (string, error) {
	const (
		BufferSize = 3
		MacPrefix  = "00:16:3e"
	)
	buf := make([]byte, BufferSize)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%02x:%02x:%02x", MacPrefix, buf[0], buf[1], buf[2]), nil
}

func (define *virDomainDefine) SetVncDisplay(port uint, secret string) error {
	const (
		GraphicVNC = "vnc"
		//GraphicSPICE = "spice"
		ListenAllAddress  = "0.0.0.0"
		ListenTypeAddress = "address"
	)
	define.Devices.Graphics.Port = port
	define.Devices.Graphics.Type = GraphicVNC
	define.Devices.Graphics.Password = secret
	define.Devices.Graphics.Listen = &virDomainGraphicsListen{ListenTypeAddress, ListenAllAddress}
	return nil
}

func (define *virDomainDefine) Initial() {
	const (
		KVMInstanceType     = "kvm"
		DefaultOSName       = "hvm"
		DefaultOSArch       = "x86_64"
		DefaultOSMachine    = "pc"
		DefaultOSMachineQ35 = "q35"
		DefaultQEMUEmulator = "/usr/bin/qemu-system-x86_64"
		DestroyInstance     = "destroy"
		RestartInstance     = "restart"
		DefaultPowerEnabled = "no"
		DefaultClockOffset  = "utc"

		MemoryBalloonModel  = "virtio"
		MemoryBalloonPeriod = 2

		ChannelType       = "unix"
		ChannelTargetType = "virtio"
		ChannelTargetName = "org.qemu.guest_agent.0"
	)
	define.Type = KVMInstanceType
	define.OS.Type.Name = DefaultOSName
	define.OS.Type.Arch = DefaultOSArch
	define.OS.Type.Machine = DefaultOSMachine
	//define.OS.Type.Machine = DefaultOSMachineQ35
	define.Devices.Emulator = DefaultQEMUEmulator

	define.OnPowerOff = DestroyInstance
	define.OnReboot = RestartInstance
	define.OnCrash = RestartInstance

	define.PowerManage.Disk.Enabled = DefaultPowerEnabled
	define.PowerManage.Mem.Enabled = DefaultPowerEnabled

	define.Features.PAE = &virDomainFeaturePAE{}
	define.Features.ACPI = &virDomainFeatureACPI{}
	define.Features.APIC = &virDomainFeatureAPIC{}


	define.Clock.Offset = DefaultClockOffset

	define.Devices.MemoryBalloon.Model = MemoryBalloonModel
	define.Devices.MemoryBalloon.Stats.Period = MemoryBalloonPeriod

	define.Devices.Channel.Type = ChannelType
	define.Devices.Channel.Target = virDomainChannelTarget{ChannelTargetType, ChannelTargetName}

	define.Devices.Controller = []virDomainControllerElement{
		virDomainControllerElement{Type: PCIController, Index: DefaultControllerIndex, Model: DefaultControllerModel},
		virDomainControllerElement{Type: VirtioSCSIController, Index: DefaultControllerIndex, Model: VirtioSCSIModel}}

}

